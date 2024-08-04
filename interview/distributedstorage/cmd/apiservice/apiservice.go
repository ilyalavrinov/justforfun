package main

import (
	"context"
	"flag"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"os"

	"github.com/ilyalavrinov/justforfun/interview/distributedstorage/internal/chunkmaster"
	pb "github.com/ilyalavrinov/justforfun/interview/distributedstorage/internal/proto/storageinventory"
	"github.com/ilyalavrinov/justforfun/interview/distributedstorage/internal/storage"
	"google.golang.org/grpc"
)

func main() {
	argInventoryPort := flag.Int("inventory-port", 3609, "port where we listen for grpc info about storages")
	argChunksNum := flag.Int("chunks-num", 6, "number of chunks to split incoming file")
	flag.Parse()
	if *argInventoryPort <= 0 {
		slog.Error("inventory port is bad", "port", *argInventoryPort)
		os.Exit(1)
	}
	if *argChunksNum <= 0 {
		slog.Error("nunmber of chunks is bad", "port", *argChunksNum)
		os.Exit(1)
	}

	chunkMaster, err := startChunkMaster(*argInventoryPort, *argChunksNum)
	if err != nil {
		slog.Error("cannot start chunk master", "err", err)
		os.Exit(1)
	}

	retriever := &retrieveHandler{chunkMaster: chunkMaster}
	storer := &storeHandler{chunkMaster: chunkMaster}

	http.Handle("GET /{filepath}", retriever)
	http.Handle("POST /{filepath}", storer)

	slog.Info("apiservice started")
	err = http.ListenAndServe("", nil)
	if err != nil {
		slog.Error("server exit with error", "err", err)
	}
}

func startChunkMaster(storageInventoryPort int, chunksNum int) (chunkmaster.ChunkMaster, error) {
	connectToRemoteStorage := func(storageId string) (storage.Storage, error) {
		return storage.NewRemoteStorage(storageId)
	}
	chunkMaster := chunkmaster.NewTemporaryChunkMaster(chunksNum, connectToRemoteStorage)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", storageInventoryPort))
	if err != nil {
		return nil, fmt.Errorf("listen for chunkmaster failed: %w", err)
	}
	gsrv := grpc.NewServer()
	pb.RegisterStorageInventoryServer(gsrv, chunkMaster)

	go func() {
		slog.Info("inventory service listening", "port", storageInventoryPort)
		if err := gsrv.Serve(listener); err != nil {
			slog.Error("cannot serve grpc for chunkmaster", "err", err)
			os.Exit(2)
		}
	}()

	return chunkMaster, nil
}

type storeHandler struct {
	chunkMaster chunkmaster.ChunkMaster
}

func (h *storeHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	filepath := req.PathValue("filepath")
	slog.Info("incoming store request", "filepath", filepath, "size", req.ContentLength)
	chunks, err := h.chunkMaster.SplitToChunks(filepath, req.ContentLength)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("split to chunks error", "err", err, "filepath", filepath)
		return
	}

	ctx := req.Context()
	err = distributeData(ctx, req.Body, chunks, h.chunkMaster.Storages())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("data distribution error", "err", err, "filepath", filepath)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func distributeData(ctx context.Context, body io.Reader, chunks []chunkmaster.Chunk, storages map[string]storage.Storage) error {
	for i, chunk := range chunks {
		if i != int(chunk.Order) {
			panic("chunks are not ordered")
		}

		storage, found := storages[chunk.StorageInstance]
		if !found {
			rollbackSave(ctx)
			return fmt.Errorf("storage instance %s missing", chunk.StorageInstance)
		}
		// I don't think it is worth paralleling things here. Concurrent execution would help only if access to our storages is a bottleneck
		chunkReader := io.LimitReader(body, chunk.Size)
		err := storage.StoreChunk(ctx, chunk.FileId, chunkReader)
		if err != nil {
			rollbackSave(ctx)
			return fmt.Errorf("cannot save chunk %d on instance %s with error: %w", chunk.Order, chunk.StorageInstance, err)
		}
	}
	return nil
}

func rollbackSave(ctx context.Context) {
	// panic("TODO implement it")
}

type retrieveHandler struct {
	chunkMaster chunkmaster.ChunkMaster
}

func (h *retrieveHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	filepath := req.PathValue("filepath")
	slog.Info("incoming retrieve request", "filepath", filepath)
	chunks, err := h.chunkMaster.ChunksToRestore(filepath)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("chunks restoration error", "err", err, "filepath", filepath)
		return
	}

	ctx := req.Context()
	err = reconstructData(ctx, chunks, h.chunkMaster.Storages(), w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("data distribution error", "err", err, "filepath", filepath)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func reconstructData(ctx context.Context, chunks []chunkmaster.Chunk, storages map[string]storage.Storage, writer io.Writer) error {
	for i, chunk := range chunks {
		if i != int(chunk.Order) {
			panic("incorrect chunk order")
		}

		storage, found := storages[chunk.StorageInstance]
		if !found {
			return fmt.Errorf("storage instance %s missing", chunk.StorageInstance)
		}

		err := storage.RetrieveChunk(ctx, chunk.FileId, writer)
		if err != nil {
			return fmt.Errorf("cannot retrieve chunk %d on instance %s with error: %w", chunk.Order, chunk.StorageInstance, err)
		}
	}
	return nil
}
