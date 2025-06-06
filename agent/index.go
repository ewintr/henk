package agent

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"

	"go-mod.ewintr.nl/henk/llm"
	"go-mod.ewintr.nl/henk/storage"
	"go-mod.ewintr.nl/henk/tool"
)

type Index struct {
	fileRepo storage.FileIndex
	llm      llm.LLM
	out      chan Message
	busy     bool
	sync.RWMutex
}

func NewIndex(fileRepo storage.FileIndex, llm llm.LLM, out chan Message) *Index {
	return &Index{
		fileRepo: fileRepo,
		llm:      llm,
		out:      out,
	}
}

func (i *Index) Refresh(full bool) error {
	i.Lock()
	if i.busy {
		i.out <- Message{
			Type: TypeGeneral,
			Body: "indexer currently working, try again later",
		}
		i.Unlock()
		return nil
	}
	i.busy = true
	i.Unlock()

	dir := "."
	currentFiles := make(map[string]storage.File, 0)
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
			currentFiles[relPath] = storage.File{
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
		if full {
			needsUpdate = append(needsUpdate, cf.Path)
			continue
		}

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

	if len(needsUpdate) == 0 {
		i.out <- Message{Type: TypeGeneral, Body: "nothing to index"}
		i.busy = false
		return nil
	}

	i.out <- Message{
		Type: TypeGeneral,
		Body: fmt.Sprintf("indexing %d file(s)...", len(needsUpdate)),
	}
	for _, p := range needsUpdate {
		curFile := currentFiles[p]
		summary, err := i.summarizeFile(p)
		if err != nil {
			return fmt.Errorf("could not get summary for file %s: %v", p, err)
		}
		curFile.Summary = summary
		if err := i.fileRepo.Store(curFile); err != nil {
			return fmt.Errorf("could not store file %s: %v", p, err)
		}

	}
	i.out <- Message{
		Type: TypeGeneral,
		Body: fmt.Sprintf("indexed %d file(s)", len(needsUpdate)),
	}
	i.busy = false
	return nil
}

func (i *Index) summarizeFile(path string) (string, error) {
	content, err := os.ReadFile(path)
	if err != nil {
		return "", fmt.Errorf("could not read file %s: %v", path, err)
	}

	if len(content) > 10000 {
		content = content[:10000] // Take first 10K chars
	}

	ctx := context.Background()
	conv := summaryConversation(path, string(content))
	response, err := i.llm.RunInference(ctx, []tool.Tool{}, conv)
	if err != nil {
		return "", fmt.Errorf("could not generate summary for %s: %v", path, err)
	}

	summary := ""
	for _, block := range response.Content {
		if block.Type == llm.ContentTypeText {
			summary = block.Text
			break
		}
	}

	return summary, nil
}

func summaryConversation(path string, content string) []llm.Message {
	// Determine file type based on extension
	ext := filepath.Ext(path)

	var prompt string
	switch ext {
	case ".go", ".js", ".py", ".java", ".c", ".cpp":
		prompt = fmt.Sprintf(
			"This is a code file: %s\nPlease provide a concise summary (under 200 words) of what this code does, including key functions, purpose, and any notable patterns or techniques:\n\n%s",
			path, content)
	case ".md", ".txt", ".adoc":
		prompt = fmt.Sprintf(
			"This is a text document: %s\nPlease provide a concise summary (under 200 words) of the main ideas and content in this document:\n\n%s",
			path, content)
	case ".json", ".yaml", ".yml", ".toml":
		prompt = fmt.Sprintf(
			"This is a configuration file: %s\nPlease provide a concise summary (under 200 words) of what this configuration defines or controls:\n\n%s",
			path, content)
	default:
		prompt = fmt.Sprintf(
			"This is a file: %s\nPlease provide a concise summary (under 200 words) of what this file contains or defines:\n\n%s",
			path, content)
	}

	conv := []llm.Message{
		{
			Role: llm.RoleUser,
			Content: []llm.ContentBlock{
				{
					Type: llm.ContentTypeText,
					Text: prompt,
				},
			},
		},
	}

	return conv
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
