package chunkmaster

import (
	"errors"
)

type Chunk struct {
	Order             uint32
	StorageInstance   string
	OriginalFileStart int64
	Size              int64
}

var (
	ErrFileDuplicate = errors.New("duplicate file")
	ErrFileNotFound  = errors.New("not found")

	ErrNotEnoughStorageNodes     = errors.New("not enough storage nodes")
	ErrNotEnoughAvailableStorage = errors.New("not enough free space")
)

type StorageInfo struct {
	StorageID      string
	AvailableBytes int64
}

type ChunkMaster interface {
	// splitting functionality
	SplitToChunks(filepath string, size int64, storages map[string]StorageInfo) ([]Chunk, error)
	ChunksToRestore(filepath string) ([]Chunk, error)
	DeleteChunks(filepath string)
}
