package chunkmaster

import (
	"errors"

	"github.com/ilyalavrinov/justforfun/interview/distributedstorage/internal/storage"
)

type Chunk struct {
	Order             int32
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

type ChunkMaster interface {
	// node registry - TODO move it away?
	NodeUp(fqdn string, storage storage.Storage)
	Storages() map[string]storage.Storage

	// splitting functionality
	SplitToChunks(filepath string, size int64) ([]Chunk, error)
	ChunksToRestore(filepath string) ([]Chunk, error)
}
