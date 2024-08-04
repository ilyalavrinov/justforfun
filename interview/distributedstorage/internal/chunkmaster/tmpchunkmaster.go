package chunkmaster

import (
	"encoding/base64"
	"fmt"
	"math/rand/v2"

	"github.com/ilyalavrinov/justforfun/interview/distributedstorage/internal/storage"
)

type TemporaryChunkMaster struct {
	knownStorages map[string]storage.Storage
	chunkCatalog  map[string][]Chunk

	splitNumber int
}

var _ ChunkMaster = (*TemporaryChunkMaster)(nil)

func NewTemporaryChunkMaster(chunkSplitNumber int) ChunkMaster {
	return &TemporaryChunkMaster{
		knownStorages: make(map[string]storage.Storage),
		chunkCatalog:  make(map[string][]Chunk),
		splitNumber:   chunkSplitNumber,
	}
}

func (tmp *TemporaryChunkMaster) NodeUp(fqdn string, storage storage.Storage) {
	tmp.knownStorages[fqdn] = storage
}

func (tmp *TemporaryChunkMaster) Storages() map[string]storage.Storage {
	return tmp.knownStorages
}

func (tmp *TemporaryChunkMaster) SplitToChunks(filepath string, size int64) ([]Chunk, error) {
	if len(tmp.knownStorages) < tmp.splitNumber {
		return nil, ErrNotEnoughStorageNodes
	}

	fullFileId := incomingFilePathToId(filepath)
	if _, found := tmp.chunkCatalog[fullFileId]; found {
		return nil, ErrFileDuplicate
	}

	storages := make([]string, 0, tmp.splitNumber)
	for fqdn := range tmp.knownStorages {
		storages = append(storages, fqdn)
	}
	rand.Shuffle(tmp.splitNumber, func(i, j int) {
		storages[i], storages[j] = storages[j], storages[i]
	})

	if size < int64(tmp.splitNumber) {
		return []Chunk{{
			Order:             0,
			StorageInstance:   storages[0],
			OriginalFileStart: 0,
			Size:              size,
		}}, nil
	}

	chunks := make([]Chunk, 0, tmp.splitNumber)
	chunkSize := size / int64(tmp.splitNumber)
	for i := range tmp.splitNumber {
		chunks = append(chunks, Chunk{
			Order:             int32(i),
			StorageInstance:   storages[i],
			OriginalFileStart: int64(i) * chunkSize,
			Size:              chunkSize,
			FileId:            fmt.Sprintf("%s.part.%d", fullFileId, i),
		})
	}
	chunks[len(chunks)-1].Size = size - (chunkSize * int64(tmp.splitNumber-1))
	tmp.chunkCatalog[fullFileId] = chunks

	return chunks, nil
}

func (tmp *TemporaryChunkMaster) ChunksToRestore(filepath string) ([]Chunk, error) {
	fullFileId := incomingFilePathToId(filepath)
	chunks, found := tmp.chunkCatalog[fullFileId]
	if !found {
		return nil, ErrFileNotFound
	}
	return chunks, nil
}

func incomingFilePathToId(filepath string) string {
	return base64.StdEncoding.EncodeToString([]byte(filepath))
}
