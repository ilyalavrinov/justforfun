package chunkmaster

import (
	"context"
	"encoding/base64"
	"fmt"
	"log/slog"
	"math/rand/v2"
	"sync"

	pb "github.com/ilyalavrinov/justforfun/interview/distributedstorage/internal/proto/storageinventory"
	"github.com/ilyalavrinov/justforfun/interview/distributedstorage/internal/storage"
	"google.golang.org/protobuf/types/known/emptypb"
)

type TemporaryChunkMaster struct {
	// storage inventory
	pb.UnsafeStorageInventoryServer
	storageMutex   sync.RWMutex
	storages       map[string]storage.Storage
	storageMem     map[string]int64
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
		storages:       make(map[string]storage.Storage),
		storageMem:     make(map[string]int64),
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
		res[id] = storage
	}
	return res
}

func (cm *TemporaryChunkMaster) SplitToChunks(filepath string, size int64) ([]Chunk, error) {
	if len(cm.storages) < cm.splitNumber {
		return nil, ErrNotEnoughStorageNodes
	}

	cm.chunkMutex.Lock()
	defer cm.chunkMutex.Unlock()

	fullFileId := incomingFilePathToId(filepath)
	if _, found := cm.chunkCatalog[fullFileId]; found {
		return nil, ErrFileDuplicate
	}

	storages := make([]string, 0, cm.splitNumber)
	for fqdn := range cm.storages {
		storages = append(storages, fqdn)
	}
	// TODO: do something better like greedy approach, or other better distribution
	rand.Shuffle(cm.splitNumber, func(i, j int) {
		storages[i], storages[j] = storages[j], storages[i]
	})

	if size < int64(cm.splitNumber) {
		return []Chunk{{
			Order:             0,
			StorageInstance:   storages[0],
			OriginalFileStart: 0,
			Size:              size,
		}}, nil
	}

	chunks := make([]Chunk, 0, cm.splitNumber)
	chunkSize := size / int64(cm.splitNumber)
	for i := range cm.splitNumber {
		chunks = append(chunks, Chunk{
			Order:             int32(i),
			StorageInstance:   storages[i],
			OriginalFileStart: int64(i) * chunkSize,
			Size:              chunkSize,
			FileId:            fmt.Sprintf("%s.part.%d", fullFileId, i),
		})
	}
	chunks[len(chunks)-1].Size = size - (chunkSize * int64(cm.splitNumber-1))
	cm.chunkCatalog[fullFileId] = chunks

	return chunks, nil
}

func (cm *TemporaryChunkMaster) ChunksToRestore(filepath string) ([]Chunk, error) {
	cm.chunkMutex.RLock()
	defer cm.chunkMutex.RUnlock()

	fullFileId := incomingFilePathToId(filepath)
	chunks, found := cm.chunkCatalog[fullFileId]
	if !found {
		return nil, ErrFileNotFound
	}
	return chunks, nil
}

func incomingFilePathToId(filepath string) string {
	return base64.StdEncoding.EncodeToString([]byte(filepath))
}

func (cm *TemporaryChunkMaster) UpdateStorageInfo(_ context.Context, info *pb.StorageInfo) (*emptypb.Empty, error) {
	cm.storageMutex.Lock()
	defer cm.storageMutex.Unlock()

	storageID := info.GetIam()
	_, found := cm.storages[storageID]
	if !found {
		rs, err := cm.storageCreator(storageID)
		if err != nil {
			slog.Error("cannot add new storage", "storage_id", storageID, "err", err)
			return nil, err
		}
		cm.storages[storageID] = rs
		slog.Info("added new storage", "storage_id", storageID)
	}
	cm.storageMem[storageID] = info.GetAvailableBytes()
	return nil, nil
}
