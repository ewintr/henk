package agent

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"go-mod.ewintr.nl/henk/storage"
)

type Index struct {
	fileRepo storage.FileIndex
}

func NewIndex(fileRepo storage.FileIndex) *Index {
	return &Index{
		fileRepo: fileRepo,
	}
}

func (i *Index) Refresh() error {
	dir := "."
	var files []storage.File
	if err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		if !strings.HasPrefix(relPath, ".") && !info.IsDir() {
			files = append(files, storage.File{
				Path:    relPath,
				Updated: info.ModTime(),
			})
		}

		return nil
	}); err != nil {
		return fmt.Errorf("could not list files: %v", err)
	}

	// fmt.Printf("%+v\n", files)

	// TODO find deleted

	for _, f := range files {
		// fmt.Println(p)
		hash, err := calculateMD5(f.Path)
		if err != nil {
			return fmt.Errorf("could not calculate md5 of %s: %v", f, err)
		}
		f.Hash = hash

		if err := i.fileRepo.Store(f); err != nil {
			return fmt.Errorf("could not store file %s: %v", f, err)
		}

	}
	return nil
}

func (i *Index) processFile(path string) error {

	return nil
}

func calculateMD5(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}
