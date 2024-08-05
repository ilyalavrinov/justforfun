package chunkmaster

import (
	"context"
	"fmt"
	"testing"

	pb "github.com/ilyalavrinov/justforfun/interview/distributedstorage/internal/proto/storageinventory"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newReadyForTestTmpChunker(numberOfChunks int) ChunkMaster {
	chunker := NewTemporaryChunkMaster(numberOfChunks)
	cm := chunker.(*TemporaryChunkMaster)
	for i := range numberOfChunks {
		nodeInfo := &pb.StorageInfo{
			Iam:            fmt.Sprintf("tempstorage-%d", i),
			AvailableBytes: 0,
		}
		cm.UpdateStorageInfo(context.Background(), nodeInfo)
	}
	return chunker
}

func TestNotEnoughStorageHosts(t *testing.T) {
	chunker := NewTemporaryChunkMaster(6)
	cm := chunker.(*TemporaryChunkMaster)
	for i := range 5 {
		nodeInfo := &pb.StorageInfo{
			Iam:            fmt.Sprintf("tempstorage-%d", i),
			AvailableBytes: 0,
		}
		cm.UpdateStorageInfo(context.Background(), nodeInfo)
	}
	chunks, err := chunker.SplitToChunks("some/path", 9000)
	assert.Nil(t, chunks)
	assert.ErrorIs(t, err, ErrNotEnoughStorageNodes)
}

func TestSplitDataForOnlyOneChunk(t *testing.T) {
	chunker := newReadyForTestTmpChunker(6)
	chunks, err := chunker.SplitToChunks("some/path", 3)
	require.NoError(t, err)
	require.Len(t, chunks, 1)
	assert.EqualValues(t, 0, chunks[0].Order)
	assert.EqualValues(t, 0, chunks[0].OriginalFileStart)
	assert.EqualValues(t, 3, chunks[0].Size)
}

func TestSplitDataSmallSize(t *testing.T) {
	chunker := newReadyForTestTmpChunker(6)
	chunks, err := chunker.SplitToChunks("some/path", 8)
	require.NoError(t, err)
	assert.Len(t, chunks, 6)
	var sumChunks int64 = 0
	nextExpectedOrder := 0
	for _, chunk := range chunks {
		sumChunks += chunk.Size

		assert.EqualValues(t, nextExpectedOrder, chunk.Order)
		nextExpectedOrder++
	}
	assert.EqualValues(t, 8, sumChunks)
}

func TestSplitData(t *testing.T) {
	chunker := newReadyForTestTmpChunker(6)
	chunks, err := chunker.SplitToChunks("some/path", 9007)
	require.NoError(t, err)
	assert.Len(t, chunks, 6)
	var sumChunks int64 = 0
	nextExpectedOrder := 0
	for _, chunk := range chunks {
		sumChunks += chunk.Size

		assert.EqualValues(t, nextExpectedOrder, chunk.Order)
		nextExpectedOrder++
	}
	assert.EqualValues(t, 9007, sumChunks)
}

func TestDuplicatesNotAllowed(t *testing.T) {
	chunker := newReadyForTestTmpChunker(6)
	filepath := "same/path"
	chunks, err := chunker.SplitToChunks(filepath, 9007)
	require.NoError(t, err)
	assert.Len(t, chunks, 6)
	_, err = chunker.SplitToChunks(filepath, 1035)
	require.ErrorIs(t, err, ErrFileDuplicate)
}

func TestSplitAndRetrieve(t *testing.T) {
	chunker := newReadyForTestTmpChunker(6)
	filepath := "this/is/my/path123"
	chunksSplit, err := chunker.SplitToChunks(filepath, 54623)
	require.NoError(t, err)
	chunksRestore, err := chunker.ChunksToRestore(filepath)
	require.NoError(t, err)
	require.EqualValues(t, chunksSplit, chunksRestore)
}

func TestFileNotFound(t *testing.T) {
	chunker := newReadyForTestTmpChunker(6)
	_, err := chunker.ChunksToRestore("abc/3424/ty")
	require.ErrorIs(t, err, ErrFileNotFound)
}
