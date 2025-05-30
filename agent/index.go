package agent

import (
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"go-mod.ewintr.nl/henk/internal"
)

type Index struct {
	fileRepo internal.FileIndex
}

func NewIndex(fileRepo internal.FileIndex) *Index {
	return &Index{
		fileRepo: fileRepo,
	}
}

func (i *Index) Refresh() error {
	dir := "."
	currentFiles := make(map[string]internal.File, 0)
	if err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(dir, path)
		if err != nil {
			return err
		}

		if !strings.HasPrefix(relPath, ".") && !info.IsDir() {
			hash, err := calculateMD5(relPath)
			if err != nil {
				return fmt.Errorf("could not calculate md5 of %s: %v", relPath, err)
			}
			currentFiles[relPath] = internal.File{
				Path:    relPath,
				Updated: info.ModTime(),
				Hash:    hash,
			}
		}

		return nil
	}); err != nil {
		return fmt.Errorf("could not list files: %v", err)
	}

	knownFiles, err := i.fileRepo.FindAll()
	if err != nil {
		return err
	}
	needsDeletion := make(map[string]bool, len(knownFiles))
	for _, kf := range knownFiles {
		needsDeletion[kf.Path] = true
	}

	needsUpdate := make([]string, 0)
	for _, cf := range currentFiles {
		knownVersion, ok := knownFiles[cf.Path]
		if !ok {
			needsUpdate = append(needsUpdate, cf.Path)
			continue
		}
		if knownVersion.Hash != cf.Hash {
			needsUpdate = append(needsUpdate, cf.Path)
			delete(needsDeletion, cf.Path)
		}
	}

	for _, p := range needsUpdate {
		if err := i.fileRepo.Store(currentFiles[p]); err != nil {
			return fmt.Errorf("could not store file %s: %v", p, err)
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
