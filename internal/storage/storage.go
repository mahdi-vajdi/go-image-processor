package storage

import (
	"context"
	"io"
)

type Storage interface {
	Save(ctx context.Context, filename string, data io.Reader) (string, error)
	Get(ctx context.Context, filename string) (io.ReadCloser, error)
	Delete(ctx context.Context, filename string) error
}
