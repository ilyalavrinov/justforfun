package storage

import (
	"context"
	"fmt"
	"io"
	"log/slog"

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

func (rs *remoteStorage) StoreChunk(ctx context.Context, fileId string, reader io.Reader) error {
	stream, err := rs.client.StoreData(ctx)
	if err != nil {
		return fmt.Errorf("store stream open failed for %s: %w", fileId, err)
	}
	var totalWritten int
	done := false
	const portionSize int = 1024 * 1024
	for !done {
		data := make([]byte, portionSize)
		readCnt, readErr := reader.Read(data)
		unit := &pb.StoredUnit{
			FileInfo: &pb.FileInfo{
				FileId: fileId,
			},
			Data: data[:readCnt],
		}
		err := stream.Send(unit)
		if err != nil {
			stream.CloseSend()
			return fmt.Errorf("stream send failed for %s: %w", fileId, err)
		}
		totalWritten += readCnt
		if readErr == io.EOF || readCnt == 0 {
			done = true
		}
	}
	_, err = stream.CloseAndRecv()
	slog.Info("chunk sent", "file_id", fileId, "written", totalWritten, "err", err)
	return err
}

func (rs *remoteStorage) RetrieveChunk(ctx context.Context, fileId string, writer io.Writer) error {
	info := &pb.FileInfo{
		FileId: fileId,
	}
	stream, err := rs.client.RetrieveData(ctx, info)
	if err != nil {
		return fmt.Errorf("remote retrieve data failed: %w", err)
	}
	var totalWritten int
	for {
		unit, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			stream.CloseSend()
			return fmt.Errorf("stream receive failed for %s: %w", fileId, err)
		}
		written, err := writer.Write(unit.GetData())
		if err != nil {
			stream.CloseSend()
			return fmt.Errorf("write portion failed for %s: %w", fileId, err)
		}
		totalWritten += written
	}
	slog.Info("remote retrieve done", "file_id", fileId, "written", totalWritten)
	return stream.CloseSend()
}

func (rs *remoteStorage) DeleteChunk(ctx context.Context, fileId string) error {
	info := &pb.FileInfo{
		FileId: fileId,
	}
	_, err := rs.client.DeleteData(ctx, info)
	if err != nil {
		return fmt.Errorf("remote delete data failed: %w", err)
	}
	slog.Info("remote delete done", "file_id", fileId, "err", err)
	return err
}
