package storage

import (
	"context"
	"io"
)

type Storage interface {
	StoreChunk(context.Context, string, io.Reader) error
	RetrieveChunk(context.Context, string, io.Writer) error
	DeleteChunk(context.Context, string) error
}
