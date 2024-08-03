package main

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"

	"github.com/ilyalavrinov/justforfun/interview/distributedstorage/internal/chunkmaster"
	"github.com/ilyalavrinov/justforfun/interview/distributedstorage/internal/storage"
)

func main() {
	chunkMaster := chunkmaster.NewTemporaryChunkMaster(2)
	addr1 := "localhost:45001"
	remote1, err := storage.NewRemoteStorage(addr1)
	if err != nil {
		slog.Error("cannot connect to remote1", "err", err)
		os.Exit(1)
	}
	addr2 := "localhost:45002"
	remote2, err := storage.NewRemoteStorage(addr2)
	if err != nil {
		slog.Error("cannot connect to remote2", "err", err)
		os.Exit(1)
	}
	chunkMaster.NodeUp(addr1, remote1)
	chunkMaster.NodeUp(addr2, remote2)

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
	err = distributeData(ctx, req.Body, filepath, chunks, h.chunkMaster.Storages())
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("data distribution error", "err", err, "filepath", filepath)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func distributeData(ctx context.Context, body io.Reader, filepath string, chunks []chunkmaster.Chunk, storages map[string]storage.Storage) error {
	for i, chunk := range chunks {
		if i != int(chunk.Order) {
			panic("chunks are not ordered")
		}

		storage, found := storages[chunk.StorageInstance]
		if !found {
			rollbackSave(ctx)
			return fmt.Errorf("storage instance %s missing", chunk.StorageInstance)
		}
		chunkReader := io.LimitReader(body, chunk.Size)
		err := storage.AcceptChunk(ctx, chunkId(filepath, chunk.Order), chunkReader)
		if err != nil {
			rollbackSave(ctx)
			return fmt.Errorf("cannot save chunk %d on instance %s with error: %w", chunk.Order, chunk.StorageInstance, err)
		}
	}
	return nil
}

func rollbackSave(ctx context.Context) {
	panic("TODO implement it")
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
	err = reconstructData(ctx, filepath, chunks, h.chunkMaster.Storages(), w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("data distribution error", "err", err, "filepath", filepath)
		return
	}

	w.WriteHeader(http.StatusOK)
}

func reconstructData(ctx context.Context, filepath string, chunks []chunkmaster.Chunk, storages map[string]storage.Storage, writer io.Writer) error {
	for i, chunk := range chunks {
		if i != int(chunk.Order) {
			panic("incorrect chunk order")
		}

		storage, found := storages[chunk.StorageInstance]
		if !found {
			return fmt.Errorf("storage instance %s missing", chunk.StorageInstance)
		}

		err := storage.RetrieveChunk(ctx, chunkId(filepath, chunk.Order), writer)
		if err != nil {
			return fmt.Errorf("cannot retrieve chunk %d on instance %s with error: %w", chunk.Order, chunk.StorageInstance, err)
		}
	}
	return nil
}

// TODO probably should use separate type for chunkIDs for typesafety. Or it is storage's responsibility to do it internally
func chunkId(filepath string, order int32) string {
	return fmt.Sprintf("%s.part.%d", filepath, order)
}
