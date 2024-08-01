package chunkmaster

import "math/rand/v2"

type TemporaryChunkMaster struct {
	knownStorages map[string]bool
	chunkCatalog  map[string][]Chunk

	splitNumber int
}

var _ ChunkMaster = (*TemporaryChunkMaster)(nil)

func NewTemporaryChunkMaster(chunkSplitNumber int) ChunkMaster {
	return &TemporaryChunkMaster{
		knownStorages: make(map[string]bool),
		chunkCatalog:  make(map[string][]Chunk),
		splitNumber:   chunkSplitNumber,
	}
}

func (tmp *TemporaryChunkMaster) NodeUp(fqdn string) {
	tmp.knownStorages[fqdn] = true
}

func (tmp *TemporaryChunkMaster) SplitToChunks(filepath string, size int64) ([]Chunk, error) {
	// equal dumb strategy here just for first version
	if len(tmp.knownStorages) < tmp.splitNumber {
		return nil, ErrNotEnoughStorageNodes
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
		})
	}
	chunks[len(chunks)-1].Size = size - (chunkSize * int64(tmp.splitNumber-1))

	return chunks, nil
}

func (tmp *TemporaryChunkMaster) ChunksToRestore(filepath string) ([]Chunk, error) {
	chunks, found := tmp.chunkCatalog[filepath]
	if !found {
		return nil, ErrFileNotFound
	}
	return chunks, nil
}
