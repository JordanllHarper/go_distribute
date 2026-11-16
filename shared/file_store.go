package shared

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
)

type FileStorer interface {
	Store(filename string, b io.Reader) (err error)
}

type FileGetter interface {
	Get(filename string) ([]byte, error)
}

type FileStore interface {
	FileStorer
	FileGetter
}

type FileStorage struct {
	dir string
}

func NewFileStorage(dir string) FileStorage {
	dir = strings.TrimSpace(dir)
	if dir == "" {
		panic("Invalid dir: empty")
	}
	return FileStorage{dir}
}

type ErrDoesntExist struct {
	file string
}

func (err ErrDoesntExist) Error() string {
	return fmt.Sprintf("file %s doesn't exist", err.file)
}

func (fs FileStorage) getFullPath(filename string) string {
	return fmt.Sprintf("%s/%s", fs.dir, filename)
}

func (fs FileStorage) Get(filename string) ([]byte, error) {
	fullPath := fs.getFullPath(filename)
	f, err := os.ReadFile(fullPath)
	switch {
	case err == nil:
		return f, nil
	case errors.Is(err, os.ErrNotExist):
		return nil, ErrDoesntExist{filename}
	default:
		return nil, err
	}
}

func (fs FileStorage) Store(filename string, b io.Reader) error {
	fullPath := fmt.Sprintf("%s/%s", fs.dir, filename)
	f, err := os.Create(fullPath)
	if err != nil {
		return err
	}
	_, err = io.Copy(f, b)
	return err
}
