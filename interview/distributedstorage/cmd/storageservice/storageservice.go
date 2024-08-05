package main

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net"
	"os"
	"path"
	"time"

	storagepb "github.com/ilyalavrinov/justforfun/interview/distributedstorage/internal/proto/storage"
	inventorypb "github.com/ilyalavrinov/justforfun/interview/distributedstorage/internal/proto/storageinventory"
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
	storageLocation string
}

var _ storagepb.StorageServer = (*storageServer)(nil)

func newStorageServer(storageLocation string) (*storageServer, error) {
	err := os.MkdirAll(storageLocation, 0o700)
	if err != nil {
		return nil, err
	}

	return &storageServer{
		storageLocation: storageLocation,
	}, nil
}

func (ssrv *storageServer) StoreData(stream grpc.ClientStreamingServer[storagepb.StoredUnit, emptypb.Empty]) error {
	var (
		f            *os.File
		fullpath     string
		totalWritten int
	)
	defer func() {
		if f != nil {
			f.Close()
		}
	}()
	for {
		unit, err := stream.Recv()
		if err == io.EOF {
			break
		}
		if err != nil {
			return err
		}
		if f == nil {
			fullpath = path.Join(ssrv.storageLocation, unit.FileInfo.FileId)
			slog.Debug("creating new file", "fullpath", fullpath)
			_, err = os.Stat(fullpath)
			if err == nil || !errors.Is(err, os.ErrNotExist) {
				return fmt.Errorf("file already exists at %s", fullpath)
			}

			f, err = os.Create(fullpath)
			if err != nil {
				return fmt.Errorf("cannot create file: %w", err)
			}
		}

		written, err := f.Write(unit.GetData())
		totalWritten += written
		if err != nil {
			return fmt.Errorf("data portion copy error for %s: %w", fullpath, err)
		}
	}
	slog.Info("accept full data done", "fullpath", fullpath, "written", totalWritten)
	return stream.SendAndClose(nil)
}

func (ssrv *storageServer) RetrieveData(in *storagepb.FileInfo, gsrv grpc.ServerStreamingServer[storagepb.StoredUnit]) error {
	fullpath := path.Join(ssrv.storageLocation, in.GetFileId())

	f, err := os.Open(fullpath)
	if err != nil {
		return fmt.Errorf("cannot open data at %s, err: %w", fullpath, err)
	}

	var totalWritten int64
	done := false
	for !done {
		portionReader := io.LimitReader(f, 1024*1024)
		var buffer bytes.Buffer
		written, copyErr := io.Copy(&buffer, portionReader)
		unit := &storagepb.StoredUnit{
			FileInfo: in,
			Data:     buffer.Bytes(),
		}
		err := gsrv.Send(unit)
		totalWritten += written
		if err != nil {
			return fmt.Errorf("file stream send failed: %w", err)
		}
		if copyErr == io.EOF || written == 0 {
			done = true
		}
	}
	slog.Info("send complete", "fullpath", fullpath, "written", totalWritten)
	return nil
}

func (ssrv *storageServer) DeleteData(ctx context.Context, in *storagepb.FileInfo) (*emptypb.Empty, error) {
	fullpath := path.Join(ssrv.storageLocation, in.GetFileId())

	err := os.Remove(fullpath)
	if err != nil {
		return nil, fmt.Errorf("cannot delete data at %s, err: %w", fullpath, err)
	}
	slog.Info("delete data done", "fullpath", fullpath)
	return nil, nil
}

func runHeartbeatSender(iam, storageDir, inventoryServerAddr string) {
	ticker := time.NewTicker(1 * time.Second)
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
		} else {
			slog.Debug("heartbeat successfully sent", "iam", iam, "to", inventoryServerAddr, "available_bytes", availableBytes)
		}
		cancel()
	}
}
