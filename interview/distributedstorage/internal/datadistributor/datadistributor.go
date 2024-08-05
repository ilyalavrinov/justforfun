package datadistributor

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"math"
	"sync"

	"github.com/ilyalavrinov/justforfun/interview/distributedstorage/internal/chunkmaster"
	inventorypb "github.com/ilyalavrinov/justforfun/interview/distributedstorage/internal/proto/storageinventory"
	"github.com/ilyalavrinov/justforfun/interview/distributedstorage/internal/storage"
	"google.golang.org/protobuf/types/known/emptypb"
)

type storageMeta struct {
	storageID      string
	storage        storage.Storage
	availableBytes int64
}

type ConnectStorageFunc func(string) (storage.Storage, error)

type DataDistributor struct {
	inventorypb.UnsafeStorageInventoryServer
	knownStorages  map[string]*storageMeta
	storageCreator ConnectStorageFunc
	storageMutex   sync.Mutex

	chunkMaster chunkmaster.ChunkMaster
}

func NewDataDistributor(chunkMaster chunkmaster.ChunkMaster, connectFunc ConnectStorageFunc) *DataDistributor {
	return &DataDistributor{
		chunkMaster:    chunkMaster,
		storageCreator: connectFunc,
		knownStorages:  make(map[string]*storageMeta),
	}
}

func (dd *DataDistributor) DistributeData(ctx context.Context, inputFilename string, size int64, reader io.Reader) error {
	chunks, err := dd.determineChunksReserveQuota(inputFilename, size)
	if err != nil {
		return fmt.Errorf("quoting failed: %w", err)
	}

	for i, chunk := range chunks {
		if i != int(chunk.Order) {
			panic("chunks are not ordered")
		}

		storage, found := dd.knownStorages[chunk.StorageInstance]
		if !found {
			dd.rollbackSave(ctx, inputFilename, chunks, i)
			return fmt.Errorf("storage instance %s missing", chunk.StorageInstance)
		}
		// I don't think it is worth paralleling things here. Concurrent execution would help only if access to our storages is a bottleneck
		chunkReader := io.LimitReader(reader, chunk.Size)
		chunkFileId := incomingFilenameToChunkFileId(inputFilename, chunk.Order)
		err := storage.storage.StoreChunk(ctx, chunkFileId, chunkReader)
		if err != nil {
			dd.rollbackSave(ctx, inputFilename, chunks, i)
			return fmt.Errorf("cannot save chunk %d on instance %s with error: %w", chunk.Order, chunk.StorageInstance, err)
		}
	}
	return nil
}

func (dd *DataDistributor) determineChunksReserveQuota(inputFilename string, size int64) ([]chunkmaster.Chunk, error) {
	// we're going to reserve the quotas from instances
	dd.storageMutex.Lock()
	defer dd.storageMutex.Unlock()
	storageInfo := make(map[string]chunkmaster.StorageInfo, len(dd.knownStorages))
	for _, storageMeta := range dd.knownStorages {
		storageInfo[storageMeta.storageID] = chunkmaster.StorageInfo{
			StorageID:      storageMeta.storageID,
			AvailableBytes: storageMeta.availableBytes,
		}
	}

	chunks, err := dd.chunkMaster.SplitToChunks(inputFilename, size, storageInfo)
	if err != nil {
		return nil, fmt.Errorf("split to chunks failed for %s: %w", inputFilename, err)
	}

	for _, chunk := range chunks {
		storageMeta := dd.knownStorages[chunk.StorageInstance]
		storageMeta.availableBytes -= chunk.Size
		// it is likely OK that we have negative here, though it's undesirable.
		// The storage will reject the payload
	}

	return chunks, nil
}

func (dd *DataDistributor) rollbackSave(ctx context.Context, inputFilename string, chunks []chunkmaster.Chunk, failedChunk int) {
	dd.storageMutex.Lock()
	for i, chunk := range chunks {
		storage := dd.knownStorages[chunk.StorageInstance]
		storage.availableBytes -= chunk.Size
		if i < failedChunk {
			storage.storage.DeleteChunk(ctx, incomingFilenameToChunkFileId(inputFilename, uint32(i)))
		}
	}
	dd.storageMutex.Unlock()
	dd.chunkMaster.DeleteChunks(inputFilename)
}

func (dd *DataDistributor) ReconstructData(ctx context.Context, inputFilename string, writer io.Writer) error {
	chunks, err := dd.chunkMaster.ChunksToRestore(inputFilename)
	if err != nil {
		return fmt.Errorf("cannot restore chunks for %s: %w", inputFilename, err)
	}

	for i, chunk := range chunks {
		if i != int(chunk.Order) {
			panic("incorrect chunk order")
		}

		storageMeta, found := dd.knownStorages[chunk.StorageInstance]
		if !found {
			return fmt.Errorf("storage instance %s missing", chunk.StorageInstance)
		}

		chunkFileId := incomingFilenameToChunkFileId(inputFilename, chunk.Order)
		err := storageMeta.storage.RetrieveChunk(ctx, chunkFileId, writer)
		if err != nil {
			return fmt.Errorf("cannot retrieve chunk %d on instance %s with error: %w", chunk.Order, chunk.StorageInstance, err)
		}
	}
	return nil
}

func (dd *DataDistributor) UpdateStorageInfo(_ context.Context, info *inventorypb.StorageInfo) (*emptypb.Empty, error) {
	dd.storageMutex.Lock()
	defer dd.storageMutex.Unlock()

	storageID := info.GetIam()
	meta, found := dd.knownStorages[storageID]
	if !found {
		rs, err := dd.storageCreator(storageID)
		if err != nil {
			slog.Error("cannot add new storage", "storage_id", storageID, "err", err)
			return nil, err
		}
		meta = &storageMeta{
			storageID:      storageID,
			storage:        rs,
			availableBytes: math.MaxInt64,
		}
		dd.knownStorages[storageID] = meta
		slog.Info("added new storage", "storage_id", storageID)
	}
	availBytesNow := meta.availableBytes
	slog.Info("heartbeat received", "from", storageID, "available_bytes_received", info.GetAvailableBytes(), "available_bytes_known", availBytesNow)
	// TODO: here we'll be getting a race condition when a chunk is being uploaded/removed, which can easily lead to overbooking of space.
	// But it should self-recover! ..probably.
	// Using our quotation calculation should be prevailing over what we have reported from host if we detect shortage
	if info.GetAvailableBytes() < availBytesNow {
		meta.availableBytes = info.GetAvailableBytes()
	}
	return nil, nil
}

func incomingFilenameToChunkFileId(incomingFilename string, chunk uint32) string {
	return fmt.Sprintf("%s.part.%d", base64.StdEncoding.EncodeToString([]byte(incomingFilename)), chunk)
}
