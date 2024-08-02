package storage

import "io"

type Storage interface {
	AcceptChunk(string, io.Reader) error
	RetrieveChunk(string, io.Writer) error
}
