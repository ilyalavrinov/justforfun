package storage

import (
	"context"
	"fmt"
	"io"

	pb "github.com/ilyalavrinov/justforfun/interview/distributedstorage/internal/proto/storage"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type remoteStorage struct {
	conn   *grpc.ClientConn
	client pb.StorageClient
}

var _ Storage = (*remoteStorage)(nil)

func NewRemoteStorage(addr string) (Storage, error) {
	conn, err := grpc.NewClient(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
	if err != nil {
		return nil, fmt.Errorf("remote storage cannot connect: %w", err)
	}
	return &remoteStorage{
		conn:   conn,
		client: pb.NewStorageClient(conn),
	}, nil
}

func (rs *remoteStorage) AcceptChunk(ctx context.Context, fileId string, reader io.Reader) error {
	data, err := io.ReadAll(reader) // TODO: get rid of it. Too much reading, I think I can pass reader directly somehow
	if err != nil {
		return fmt.Errorf("cannot readall: %w", err)
	}
	unit := &pb.StoredUnit{
		FileId: fileId,
		Data:   data,
	}
	_, err = rs.client.StoreData(ctx, unit)
	if err != nil {
		return fmt.Errorf("remote store data failed: %w", err)
	}
	return nil
}

func (rs *remoteStorage) RetrieveChunk(ctx context.Context, fileId string, writer io.Writer) error {
	return fmt.Errorf("remote retrieve not implemented")
}
