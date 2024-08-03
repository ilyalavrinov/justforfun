package storage

import (
	"bytes"
	"context"
	"crypto/rand"
	"io/fs"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTempStorageSendRetrieve(t *testing.T) {
	storage := NewTmpStorage()

	randdata := make([]byte, 100)
	rand.Read(randdata)
	reader := bytes.NewReader(randdata)

	filename := "this/is/my/file1"
	ctx := context.Background()
	err := storage.AcceptChunk(ctx, filename, reader)
	require.NoError(t, err)

	var retrieved bytes.Buffer
	err = storage.RetrieveChunk(ctx, filename, &retrieved)
	require.NoError(t, err)

	assert.EqualValues(t, randdata, retrieved.Bytes())
}

func TestTempStorageSendTwice(t *testing.T) {
	storage := NewTmpStorage()

	randdata1 := make([]byte, 100)
	rand.Read(randdata1)
	reader := bytes.NewReader(randdata1)

	filename := "this/is/my/file2"
	ctx := context.Background()
	err := storage.AcceptChunk(ctx, filename, reader)
	require.NoError(t, err)
	err = storage.AcceptChunk(ctx, filename, reader)
	assert.Error(t, err)
}

func TestTempStorageTryRetrieveAbsent(t *testing.T) {
	storage := NewTmpStorage()

	ctx := context.Background()
	var retrieved bytes.Buffer
	err := storage.RetrieveChunk(ctx, "hello/this/is/patrick", &retrieved)
	require.ErrorIs(t, err, fs.ErrNotExist)
}
