package storage

import (
	"context"
	"io"
)

type Storage interface {
	AcceptChunk(context.Context, string, io.Reader) error
	RetrieveChunk(context.Context, string, io.Writer) error
}
