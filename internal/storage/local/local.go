package localStorage

import (
	"context"
	"fmt"
	"io"
	"os"
	"path"
	"path/filepath"
	"time"

	"github.com/mahdi-vajdi/go-image-processor/internal/storage"
)

type LocalStore struct {
	baseDir string
}

var _ storage.Storage = (*LocalStore)(nil)

func NewLocalStore(baseDir string) (storage.Storage, error) {
	// Create directory if it doesn't exist
	if err := os.MkdirAll(baseDir, os.ModePerm); err != nil {
		return nil, fmt.Errorf("failed to create directory %s, %w", baseDir, err)
	}

	return &LocalStore{
		baseDir: baseDir,
	}, nil
}

func (s *LocalStore) generateLocalPath(originalFilename string) string {
	extension := filepath.Ext(originalFilename)
	base := originalFilename[:len(originalFilename)-len(extension)]
	uniqueFilename := fmt.Sprintf("%s_%d%s", base, time.Now().UnixNano(), extension)

	return uniqueFilename
}

func (s *LocalStore) Save(ctx context.Context, originalFilename string, data io.Reader) (string, error) {
	// Generate a unique filename
	uniqueFilename := s.generateLocalPath(originalFilename)
	filePath := filepath.Join(s.baseDir, uniqueFilename)

	// Create the file
	outFile, err := os.Create(filePath)
	if err != nil {
		return "", fmt.Errorf("failed to craete file %s: %w", filePath, err)
	}
	defer outFile.Close()

	// Copy the data into the file
	if _, err := io.Copy(outFile, data); err != nil {
		os.Remove(filePath)
		return "", fmt.Errorf("failed to copy data to the file %s: %w", filePath, err)
	}

	return uniqueFilename, nil
}

func (s *LocalStore) Get(ctx context.Context, filename string) (io.ReadCloser, error) {
	filePath := path.Join(s.baseDir, filename)

	file, err := os.Open(filePath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, fmt.Errorf("file %s not found: %w", filePath, err)
		}
		return nil, fmt.Errorf("failed to open file %s: %w", filePath, err)
	}

	return file, nil
}

func (s *LocalStore) Delete(ctx context.Context, filename string) error {
	filePath := path.Join(s.baseDir, filename)

	if err := os.Remove(filePath); err != nil {
		if os.IsNotExist(err) {
			return fmt.Errorf("file %s not found: %w", filePath, os.ErrNotExist)
		}
		return fmt.Errorf("falied to delete file %s: %w", filePath, err)
	}

	return nil
}
