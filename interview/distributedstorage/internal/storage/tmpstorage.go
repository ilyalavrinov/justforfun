package storage

import (
	"errors"
	"fmt"
	"io"
	"io/fs"
	"log/slog"
	"os"
	"path"
)

// saves stuff to local /tmp
type TmpStorage struct {
	rootDir string
}

var _ Storage = (*TmpStorage)(nil)

func NewTmpStorage() Storage {
	tmpdir, err := os.MkdirTemp(os.TempDir(), "diststorage-")
	if err != nil {
		panic(fmt.Sprintf("cannot create temp storage: %s", err))
	}
	return &TmpStorage{
		rootDir: tmpdir,
	}
}

func (ts *TmpStorage) AcceptChunk(filepath string, reader io.Reader) error {
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

func (ts *TmpStorage) RetrieveChunk(filepath string, writer io.Writer) error {
	fullpath := path.Join(ts.rootDir, filepath)

	f, err := os.Open(fullpath)
	if err != nil {
		return fmt.Errorf("cannot open chunk at %s, err: %w", filepath, err)
	}

	written, err := io.Copy(writer, f)
	slog.Info("retrieve chunk done", "fullpath", fullpath, "written", written)
	return err
}
