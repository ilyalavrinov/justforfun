package chunkmaster

import (
	"sort"
	"sync"
)

type TemporaryChunkMaster struct {
	chunkMutex   sync.RWMutex
	chunkCatalog map[string][]Chunk

	splitNumber int
}

var _ ChunkMaster = (*TemporaryChunkMaster)(nil)

func NewTemporaryChunkMaster(chunkSplitNumber int) ChunkMaster {
	return &TemporaryChunkMaster{
		chunkCatalog: make(map[string][]Chunk),
		splitNumber:  chunkSplitNumber,
	}
}

func (cm *TemporaryChunkMaster) SplitToChunks(filepath string, size int64, storages map[string]StorageInfo) ([]Chunk, error) {
	if len(storages) < cm.splitNumber {
		return nil, ErrNotEnoughStorageNodes
	}

	cm.chunkMutex.Lock()
	defer cm.chunkMutex.Unlock()

	if _, found := cm.chunkCatalog[filepath]; found {
		return nil, ErrFileDuplicate
	}

	prioritizedIds := prioritizeStorages(storages)

	// special case when we cannot split even by 1 byte to each storage
	if size < int64(cm.splitNumber) {
		targetStorageId := prioritizedIds[0]
		availMem := storages[targetStorageId].AvailableBytes
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
			Order:             uint32(i),
			StorageInstance:   prioritizedIds[i],
			OriginalFileStart: int64(i) * chunkSize,
			Size:              chunkSize,
		})
	}
	chunks[len(chunks)-1].Size = size - (chunkSize * int64(cm.splitNumber-1))

	// checking that we have enough memory
	isAllMemGood := true
	for _, chunk := range chunks {
		storageAvailable := storages[chunk.StorageInstance].AvailableBytes
		if storageAvailable < chunk.Size {
			isAllMemGood = false
		}
	}
	if !isAllMemGood {
		return nil, ErrNotEnoughAvailableStorage
	}

	cm.chunkCatalog[filepath] = chunks

	return chunks, nil
}

func prioritizeStorages(storages map[string]StorageInfo) []string {
	list := make([]StorageInfo, 0, len(storages))
	for _, info := range storages {
		list = append(list, info)
	}
	sort.Slice(list, func(i, j int) bool {
		return list[i].AvailableBytes > list[j].AvailableBytes
	})

	ids := make([]string, 0, len(list))
	for _, info := range list {
		ids = append(ids, info.StorageID)
	}
	return ids
}

func (cm *TemporaryChunkMaster) ChunksToRestore(filepath string) ([]Chunk, error) {
	cm.chunkMutex.RLock()
	defer cm.chunkMutex.RUnlock()

	chunks, found := cm.chunkCatalog[filepath]
	if !found {
		return nil, ErrFileNotFound
	}
	return chunks, nil
}

func (cm *TemporaryChunkMaster) DeleteChunks(filepath string) {
	cm.chunkMutex.Lock()
	defer cm.chunkMutex.Unlock()
	delete(cm.chunkCatalog, filepath)
}
