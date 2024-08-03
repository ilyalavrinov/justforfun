package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"math/rand"
	"os"
	"path"
)

type localStorage struct {
	rootDir string
}

var _ Storage = (*localStorage)(nil)

func NewLocalStorage(saveDir string) (Storage, error) {
	err := os.MkdirAll(saveDir, 0o700)
	if err != nil {
		return nil, err
	}
	return &localStorage{
		rootDir: saveDir,
	}, nil
}

func (ts *localStorage) AcceptChunk(_ context.Context, filepath string, reader io.Reader) error {
	fullpath := path.Join(ts.rootDir, filepath)
	_, err := os.Stat(fullpath)
	if err == nil || !errors.Is(err, os.ErrNotExist) {
		return fmt.Errorf("file already exists at %s", filepath)
	}

	err = os.MkdirAll(path.Dir(fullpath), fs.FileMode(0o700))
	if err != nil {
		return fmt.Errorf("cannot mkdirall: %w", err)
	}

	f, err := os.Create(fullpath)
	if err != nil {
		return fmt.Errorf("cannot create file: %w", err)
	}

	written, err := io.Copy(f, reader)
	slog.Info("accept chunk done", "fullpath", fullpath, "written", written)
	return err
}

func (ts *localStorage) RetrieveChunk(_ context.Context, filepath string, writer io.Writer) error {
	fullpath := path.Join(ts.rootDir, filepath)

	f, err := os.Open(fullpath)
	if err != nil {
		return fmt.Errorf("cannot open chunk at %s, err: %w", filepath, err)
	}

	written, err := io.Copy(writer, f)
	slog.Info("retrieve chunk done", "fullpath", fullpath, "written", written)
	return err
}

// creates a new directory in /tmp. Panics if something is wrong
func NewTmpStorage() Storage {
	tmpdir := path.Join(os.TempDir(), fmt.Sprintf("diststorage-%d", 100000+rand.Intn(100000)))
	storage, err := NewLocalStorage(tmpdir)
	if err != nil {
		panic(fmt.Sprintf("cannot create temp storage: %s", err))
	}
	return storage
}
