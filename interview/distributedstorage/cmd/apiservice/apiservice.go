package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"os"

	"github.com/ilyalavrinov/justforfun/interview/distributedstorage/internal/chunkmaster"
	"github.com/ilyalavrinov/justforfun/interview/distributedstorage/internal/datadistributor"
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

	dataDistributor, err := startDataDistributor(*argInventoryPort, *argChunksNum)
	if err != nil {
		slog.Error("cannot start chunk master", "err", err)
		os.Exit(1)
	}

	retriever := &retrieveHandler{dd: dataDistributor}
	storer := &storeHandler{dd: dataDistributor}

	http.Handle("GET /{filepath}", retriever)
	http.Handle("POST /{filepath}", storer)

	slog.Info("apiservice started")
	err = http.ListenAndServe("", nil)
	if err != nil {
		slog.Error("server exit with error", "err", err)
	}
}

func startDataDistributor(storageInventoryPort int, chunksNum int) (*datadistributor.DataDistributor, error) {
	connectToRemoteStorage := func(storageId string) (storage.Storage, error) {
		return storage.NewRemoteStorage(storageId)
	}
	chunkMaster := chunkmaster.NewTemporaryChunkMaster(chunksNum)
	dataDistributor := datadistributor.NewDataDistributor(chunkMaster, connectToRemoteStorage)

	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", storageInventoryPort))
	if err != nil {
		return nil, fmt.Errorf("listen for storage inventory failed: %w", err)
	}
	gsrv := grpc.NewServer()
	pb.RegisterStorageInventoryServer(gsrv, dataDistributor)

	go func() {
		slog.Info("inventory service listening", "port", storageInventoryPort)
		if err := gsrv.Serve(listener); err != nil {
			slog.Error("cannot serve grpc for storage inventory", "err", err)
			os.Exit(2)
		}
	}()

	return dataDistributor, nil
}

type storeHandler struct {
	dd *datadistributor.DataDistributor
}

func (h *storeHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	filepath := req.PathValue("filepath")
	slog.Info("incoming store request", "filepath", filepath, "size", req.ContentLength)
	err := h.dd.DistributeData(req.Context(), filepath, req.ContentLength, req.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("distribute data error", "err", err, "filepath", filepath)
		return
	}
	w.WriteHeader(http.StatusOK)
}

type retrieveHandler struct {
	dd *datadistributor.DataDistributor
}

func (h *retrieveHandler) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	filepath := req.PathValue("filepath")
	slog.Info("incoming retrieve request", "filepath", filepath)
	err := h.dd.ReconstructData(req.Context(), filepath, w)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		slog.Error("reconstruct data error", "err", err, "filepath", filepath)
		return
	}

	w.WriteHeader(http.StatusOK)
}
