package chunkmaster

import (
	"context"
	"encoding/base64"
	"fmt"
	"io"
	"log/slog"
	"math"
	"sort"
	"sync"

	pb "github.com/ilyalavrinov/justforfun/interview/distributedstorage/internal/proto/storageinventory"
	"github.com/ilyalavrinov/justforfun/interview/distributedstorage/internal/storage"
	"google.golang.org/protobuf/types/known/emptypb"
)

type storageMeta struct {
	storageID      string
	storage        storage.Storage
	availableBytes int64
}

type TemporaryChunkMaster struct {
	// storage inventory
	pb.UnsafeStorageInventoryServer
	storageMutex   sync.RWMutex
	storages       map[string]*storageMeta
	storageCreator ConnectStorageFunc

	// chunky logic
	chunkMutex   sync.RWMutex
	chunkCatalog map[string][]Chunk

	splitNumber int
}

var _ ChunkMaster = (*TemporaryChunkMaster)(nil)

type ConnectStorageFunc func(string) (storage.Storage, error)

func NewTemporaryChunkMaster(chunkSplitNumber int, connectFunc ConnectStorageFunc) ChunkMaster {
	return &TemporaryChunkMaster{
		storages:       make(map[string]*storageMeta),
		storageCreator: connectFunc,
		chunkCatalog:   make(map[string][]Chunk),
		splitNumber:    chunkSplitNumber,
	}
}

func (cm *TemporaryChunkMaster) Storages() map[string]storage.Storage {
	cm.storageMutex.RLock()
	defer cm.storageMutex.RUnlock()
	res := make(map[string]storage.Storage, len(cm.storages))
	for id, storage := range cm.storages {
		res[id] = storage.storage
	}
	return res
}

func (cm *TemporaryChunkMaster) DistributeData(ctx context.Context, filepath string, reader io.Reader, size int64) error {
	chunks, err := cm.splitToChunks(filepath, size)
	if err != nil {
		return fmt.Errorf("split to chunks failed: %w", err)
	}

	return cm.doDistributeData(ctx, reader, filepath, chunks)
}

func (cm *TemporaryChunkMaster) doDistributeData(ctx context.Context, reader io.Reader, filepath string, chunks []Chunk) error {
	cm.storageMutex.RLock()
	defer cm.storageMutex.RUnlock()
	for i, chunk := range chunks {
		if i != int(chunk.Order) {
			panic("chunks are not ordered")
		}

		storageMeta, found := cm.storages[chunk.StorageInstance]
		if !found {
			cm.rollbackSave(ctx, filepath, chunks)
			return fmt.Errorf("storage instance %s missing", chunk.StorageInstance)
		}
		// I don't think it is worth paralleling things here. Concurrent execution would help only if access to our storages is a bottleneck
		chunkReader := io.LimitReader(reader, chunk.Size)
		err := storageMeta.storage.StoreChunk(ctx, chunk.FileId, chunkReader)
		if err != nil {
			cm.rollbackSave(ctx, filepath, chunks)
			return fmt.Errorf("cannot save chunk %d on instance %s with error: %w", chunk.Order, chunk.StorageInstance, err)
		}
	}
	return nil
}

func (cm *TemporaryChunkMaster) splitToChunks(filepath string, size int64) ([]Chunk, error) {
	if len(cm.storages) < cm.splitNumber {
		return nil, ErrNotEnoughStorageNodes
	}

	cm.chunkMutex.Lock()
	defer cm.chunkMutex.Unlock()

	fullFileId := incomingFilePathToId(filepath)
	if _, found := cm.chunkCatalog[fullFileId]; found {
		return nil, ErrFileDuplicate
	}

	cm.storageMutex.Lock() // TODO: 2 mutexes taken at the same time are bad.
	defer cm.storageMutex.Unlock()

	prioritizedIds := prioritizeStorages(cm.storages)

	// special case when we cannot split even by 1 byte to each storage
	if size < int64(cm.splitNumber) {
		targetStorageId := prioritizedIds[0]
		availMem := cm.storages[targetStorageId].availableBytes
		if availMem < size {
			return nil, ErrNotEnoughAvailableStorage
		}
		return []Chunk{{
			Order:             0,
			StorageInstance:   targetStorageId,
			OriginalFileStart: 0,
			Size:              size,
		}}, nil
	}

	chunks := make([]Chunk, 0, cm.splitNumber)
	chunkSize := size / int64(cm.splitNumber)
	for i := range cm.splitNumber {
		chunks = append(chunks, Chunk{
			Order:             int32(i),
			StorageInstance:   prioritizedIds[i],
			OriginalFileStart: int64(i) * chunkSize,
			Size:              chunkSize,
			FileId:            fmt.Sprintf("%s.part.%d", fullFileId, i),
		})
	}
	chunks[len(chunks)-1].Size = size - (chunkSize * int64(cm.splitNumber-1))

	// checking that we have enough memory
	isAllMemGood := true
	for _, chunk := range chunks {
		storageAvailable := cm.storages[chunk.StorageInstance].availableBytes
		if storageAvailable < chunk.Size {
			isAllMemGood = false
		}
	}
	if !isAllMemGood {
		return nil, ErrNotEnoughAvailableStorage
	}

	// all good, reserve quota! iterating again, this chunky function is so ugly
	for _, chunk := range chunks {
		cm.storages[chunk.StorageInstance].availableBytes -= chunk.Size
	}

	cm.chunkCatalog[fullFileId] = chunks

	return chunks, nil
}

func prioritizeStorages(storages map[string]*storageMeta) []string {
	list := make([]*storageMeta, 0, len(storages))
	for _, meta := range storages {
		list = append(list, meta)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].availableBytes > list[j].availableBytes
	})

	ids := make([]string, 0, len(list))
	for _, meta := range list {
		ids = append(ids, meta.storageID)
	}
	return ids
}

func (cm *TemporaryChunkMaster) ReconstructData(ctx context.Context, filepath string, writer io.Writer) error {
	chunks, err := cm.chunksToRestore(filepath)
	if err != nil {
		return fmt.Errorf("cannot find chunks for restore %s: %w", filepath, err)
	}

	err = cm.doReconstructData(ctx, chunks, writer)
	if err != nil {
		return fmt.Errorf("cannot reconstruct data for %s: %w", filepath, err)
	}
	return nil
}

func (cm *TemporaryChunkMaster) chunksToRestore(filepath string) ([]Chunk, error) {
	cm.chunkMutex.RLock()
	defer cm.chunkMutex.RUnlock()

	fullFileId := incomingFilePathToId(filepath)
	chunks, found := cm.chunkCatalog[fullFileId]
	if !found {
		return nil, ErrFileNotFound
	}
	return chunks, nil
}

func (cm *TemporaryChunkMaster) doReconstructData(ctx context.Context, chunks []Chunk, writer io.Writer) error {
	cm.storageMutex.RLock()
	defer cm.storageMutex.RUnlock()
	for i, chunk := range chunks {
		if i != int(chunk.Order) {
			panic("incorrect chunk order")
		}

		storageMeta, found := cm.storages[chunk.StorageInstance]
		if !found {
			return fmt.Errorf("storage instance %s missing", chunk.StorageInstance)
		}

		err := storageMeta.storage.RetrieveChunk(ctx, chunk.FileId, writer)
		if err != nil {
			return fmt.Errorf("cannot retrieve chunk %d on instance %s with error: %w", chunk.Order, chunk.StorageInstance, err)
		}
	}
	return nil
}

func incomingFilePathToId(filepath string) string {
	return base64.StdEncoding.EncodeToString([]byte(filepath))
}

func (cm *TemporaryChunkMaster) UpdateStorageInfo(_ context.Context, info *pb.StorageInfo) (*emptypb.Empty, error) {
	cm.storageMutex.Lock()
	defer cm.storageMutex.Unlock()

	storageID := info.GetIam()
	meta, found := cm.storages[storageID]
	if !found {
		rs, err := cm.storageCreator(storageID)
		if err != nil {
			slog.Error("cannot add new storage", "storage_id", storageID, "err", err)
			return nil, err
		}
		meta = &storageMeta{
			storageID:      storageID,
			storage:        rs,
			availableBytes: math.MaxInt64,
		}
		cm.storages[storageID] = meta
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

func (cm *TemporaryChunkMaster) rollbackSave(ctx context.Context, filepath string, chunks []Chunk) {
	for _, chunk := range chunks {
		cm.storages[chunk.StorageInstance].availableBytes -= chunk.Size
	}
	cm.storageMutex.Unlock()

	cm.chunkMutex.Lock()
	delete(cm.chunkCatalog, filepath)
	cm.chunkMutex.Unlock()
}
