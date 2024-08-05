package chunkmaster

import (
	"context"
	"errors"

	pb "github.com/ilyalavrinov/justforfun/interview/distributedstorage/internal/proto/storageinventory"
	"github.com/ilyalavrinov/justforfun/interview/distributedstorage/internal/storage"
	"google.golang.org/protobuf/types/known/emptypb"
)

type Chunk struct {
	Order             int32
	StorageInstance   string
	OriginalFileStart int64
	Size              int64
	FileId            string
}

var (
	ErrFileDuplicate = errors.New("duplicate file")
	ErrFileNotFound  = errors.New("not found")

	ErrNotEnoughStorageNodes     = errors.New("not enough storage nodes")
	ErrNotEnoughAvailableStorage = errors.New("not enough free space")
)

type ChunkMaster interface {
	// storage inventory things
	// TODO: move it to separate thing? Make ChunkMaster more monolithinc and black-boxy?
	pb.UnsafeStorageInventoryServer
	UpdateStorageInfo(context.Context, *pb.StorageInfo) (*emptypb.Empty, error)
	Storages() map[string]storage.Storage

	// splitting functionality
	SplitToChunks(filepath string, size int64) ([]Chunk, error)
	ChunksToRestore(filepath string) ([]Chunk, error)
}
