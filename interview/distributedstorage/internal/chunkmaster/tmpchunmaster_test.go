package chunkmaster

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func randomStorages(n int) map[string]StorageInfo {
	res := make(map[string]StorageInfo, n)
	for i := range n {
		storageId := fmt.Sprintf("tempstorage-%d", i)
		res[storageId] = StorageInfo{
			StorageID:      storageId,
			AvailableBytes: 900000,
		}
	}
	return res
}

func newReadyForTestTmpChunker(numberOfChunks int) (ChunkMaster, map[string]StorageInfo) {
	return NewTemporaryChunkMaster(numberOfChunks), randomStorages(numberOfChunks)
}

func TestNotEnoughStorageHosts(t *testing.T) {
	chunker := NewTemporaryChunkMaster(6)
	storages := randomStorages(5)
	chunks, err := chunker.SplitToChunks("some/path", 9000, storages)
	assert.Nil(t, chunks)
	assert.ErrorIs(t, err, ErrNotEnoughStorageNodes)
}

func TestSplitDataForOnlyOneChunk(t *testing.T) {
	chunker, storages := newReadyForTestTmpChunker(6)
	chunks, err := chunker.SplitToChunks("some/path", 3, storages)
	require.NoError(t, err)
	require.Len(t, chunks, 1)
	assert.EqualValues(t, 0, chunks[0].Order)
	assert.EqualValues(t, 0, chunks[0].OriginalFileStart)
	assert.EqualValues(t, 3, chunks[0].Size)
}

func TestSplitDataSmallSize(t *testing.T) {
	chunker, storages := newReadyForTestTmpChunker(6)
	chunks, err := chunker.SplitToChunks("some/path", 8, storages)
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
	chunker, storages := newReadyForTestTmpChunker(6)
	chunks, err := chunker.SplitToChunks("some/path", 9007, storages)
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
	chunker, storages := newReadyForTestTmpChunker(6)
	fileref := "same/path"
	chunks, err := chunker.SplitToChunks(fileref, 9007, storages)
	require.NoError(t, err)
	assert.Len(t, chunks, 6)
	_, err = chunker.SplitToChunks(fileref, 1035, storages)
	require.ErrorIs(t, err, ErrFileDuplicate)
}

func TestSplitAndRetrieve(t *testing.T) {
	chunker, storages := newReadyForTestTmpChunker(6)
	fileref := "this/is/my/path123"
	chunksSplit, err := chunker.SplitToChunks(fileref, 54623, storages)
	require.NoError(t, err)
	chunksRestore, err := chunker.ChunksToRestore(fileref)
	require.NoError(t, err)
	require.EqualValues(t, chunksSplit, chunksRestore)
}

func TestFileNotFound(t *testing.T) {
	chunker, _ := newReadyForTestTmpChunker(6)
	_, err := chunker.ChunksToRestore("abc/3424/ty")
	require.ErrorIs(t, err, ErrFileNotFound)
}
