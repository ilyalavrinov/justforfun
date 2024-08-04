package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"

	pb "github.com/ilyalavrinov/justforfun/interview/distributedstorage/internal/proto/storage"
	"github.com/ilyalavrinov/justforfun/interview/distributedstorage/internal/storage"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/emptypb"
)

func main() {
	argStorageLocation := flag.String("storage-location", "", "location where all files will be stored locally")
	argPort := flag.Int("port", 45346, "port for listening for incoming data")
	flag.Parse()
	if *argStorageLocation == "" {
		slog.Error("missing storage location arg")
		os.Exit(1)
	}

	if *argPort <= 0 {
		slog.Error("port arg is incorrect", "port", *argPort)
		os.Exit(1)
	}

	err := runServer(*argStorageLocation, *argPort)
	if err != nil {
		slog.Error("server exited with error", "err", err)
	}
}

func runServer(storageLocation string, port int) error {
	storageSrv, err := newStorageServer(storageLocation)
	if err != nil {
		return fmt.Errorf("cannot create storage server: %w", err)
	}

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("listen failed: %w", err)
	}
	gsrv := grpc.NewServer()
	pb.RegisterStorageServer(gsrv, storageSrv)

	slog.Info("storage service listening", "port", port)
	if err := gsrv.Serve(listener); err != nil {
		return fmt.Errorf("listen filed: %w", err)
	}
	return nil
}

type storageServer struct {
	pb.UnsafeStorageServer
	storage storage.Storage
}

func newStorageServer(storageLocation string) (*storageServer, error) {
	storage, err := storage.NewLocalStorage(storageLocation)
	if err != nil {
		return nil, err
	}

	return &storageServer{
		storage: storage,
	}, nil
}

func (ssrv *storageServer) StoreData(ctx context.Context, in *pb.StoredUnit) (*emptypb.Empty, error) {
	reader := bytes.NewReader(in.GetData())
	err := ssrv.storage.StoreChunk(ctx, in.GetFileInfo().GetFileId(), reader)
	return nil, err
}

func (ssrv *storageServer) RetrieveData(ctx context.Context, in *pb.FileInfo) (*pb.StoredUnit, error) {
	var buffer bytes.Buffer
	err := ssrv.storage.RetrieveChunk(ctx, in.GetFileId(), &buffer)
	if err != nil {
		return nil, err
	}
	unit := &pb.StoredUnit{
		FileInfo: in,
		Data:     buffer.Bytes(),
	}
	return unit, nil
}
