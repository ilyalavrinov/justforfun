package chunkmaster

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newReadyForTestTmpChunker(numberOfChunks int) ChunkMaster {
	chunker := NewTemporaryChunkMaster(numberOfChunks)
	for i := range numberOfChunks {
		chunker.NodeUp(fmt.Sprintf("node_%d", i))
	}
	return chunker
}

func TestNotEnoughStorageHosts(t *testing.T) {
	chunker := NewTemporaryChunkMaster(6)
	for i := range 5 {
		chunker.NodeUp(fmt.Sprintf("node_%d", i))
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
