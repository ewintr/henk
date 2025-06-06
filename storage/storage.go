package storage

import (
	"errors"
	"time"
)

var (
	ErrNotFound = errors.New("not found")
)

type File struct {
	Path     string
	Hash     string
	FileType string
	Updated  time.Time
	Summary  string
}

type FileIndex interface {
	FindByPath(path string) (File, error)
	FindAll() (map[string]File, error)
	ListPaths() ([]string, error)
	Store(file File) error
	Delete(path string) error
}
