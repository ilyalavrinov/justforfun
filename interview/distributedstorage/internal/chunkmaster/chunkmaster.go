package chunkmaster

import (
	"context"
	"errors"
	"io"

	pb "github.com/ilyalavrinov/justforfun/interview/distributedstorage/internal/proto/storageinventory"
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
	pb.UnsafeStorageInventoryServer
	UpdateStorageInfo(context.Context, *pb.StorageInfo) (*emptypb.Empty, error)

	// splitting functionality
	DistributeData(ctx context.Context, filepath string, reader io.Reader, size int64) error
	ReconstructData(ctx context.Context, filepath string, writer io.Writer) error
}
