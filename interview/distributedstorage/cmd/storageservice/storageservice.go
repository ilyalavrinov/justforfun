package main

import (
	"bytes"
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net"
	"os"
	"time"

	storagepb "github.com/ilyalavrinov/justforfun/interview/distributedstorage/internal/proto/storage"
	inventorypb "github.com/ilyalavrinov/justforfun/interview/distributedstorage/internal/proto/storageinventory"
	"github.com/ilyalavrinov/justforfun/interview/distributedstorage/internal/storage"
	"golang.org/x/sys/unix"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/protobuf/types/known/emptypb"
)

func main() {
	argStorageLocation := flag.String("storage-location", "", "location where all files will be stored locally")
	argPort := flag.Int("port", 45346, "port for listening for incoming data")
	argInventoryHost := flag.String("inventory-host", "localhost:3609", "address to connect to notify that this storage is up")
	flag.Parse()
	if *argStorageLocation == "" {
		slog.Error("missing storage location arg")
		os.Exit(1)
	}

	if *argPort <= 0 {
		slog.Error("port arg is incorrect", "port", *argPort)
		os.Exit(1)
	}

	hostname, err := os.Hostname()
	if err != nil {
		slog.Error("error getting hostname", "err", err)
		os.Exit(1)
	}

	go runHeartbeatSender(fmt.Sprintf("%s:%d", hostname, *argPort), *argStorageLocation, *argInventoryHost)

	err = runServer(*argStorageLocation, *argPort)
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
	storagepb.RegisterStorageServer(gsrv, storageSrv)

	slog.Info("storage service listening", "port", port)
	if err := gsrv.Serve(listener); err != nil {
		return fmt.Errorf("listen filed: %w", err)
	}
	return nil
}

type storageServer struct {
	storagepb.UnsafeStorageServer
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

func (ssrv *storageServer) StoreData(ctx context.Context, in *storagepb.StoredUnit) (*emptypb.Empty, error) {
	reader := bytes.NewReader(in.GetData())
	err := ssrv.storage.StoreChunk(ctx, in.GetFileInfo().GetFileId(), reader)
	return nil, err
}

func (ssrv *storageServer) RetrieveData(ctx context.Context, in *storagepb.FileInfo) (*storagepb.StoredUnit, error) {
	var buffer bytes.Buffer
	err := ssrv.storage.RetrieveChunk(ctx, in.GetFileId(), &buffer)
	if err != nil {
		return nil, err
	}
	unit := &storagepb.StoredUnit{
		FileInfo: in,
		Data:     buffer.Bytes(),
	}
	return unit, nil
}

func (ssrv *storageServer) DeleteData(ctx context.Context, in *storagepb.FileInfo) (*emptypb.Empty, error) {
	return nil, ssrv.storage.DeleteChunk(ctx, in.GetFileId())
}

func runHeartbeatSender(iam, storageDir, inventoryServerAddr string) {
	ticker := time.NewTicker(10 * time.Second)
	for {
		<-ticker.C
		conn, err := grpc.NewClient(inventoryServerAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			slog.Error("storage inventory cannot connect", "err", err)
			continue
		}

		var stats unix.Statfs_t
		err = unix.Statfs(storageDir, &stats)
		if err != nil {
			slog.Error("cannot stat storage dir", "err", err)
			continue
		}
		availableBytes := int64(stats.Bavail) * int64(stats.Bsize)

		client := inventorypb.NewStorageInventoryClient(conn)
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		info := &inventorypb.StorageInfo{
			Iam:            iam,
			AvailableBytes: availableBytes,
		}
		_, err = client.UpdateStorageInfo(ctx, info)
		if err != nil {
			slog.Error("cannot update storage info", "err", err)
		}
		cancel()
		slog.Info("heartbeat successfully sent", "iam", iam, "to", inventoryServerAddr, "available_bytes", availableBytes)
	}
}
