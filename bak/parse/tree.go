package parse

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

var (
	ErrNotATextFile = errors.New("not a text file")
)

type ElementType string

type Element struct {
	Type        ElementType
	Description string
	Content     string
}

type File struct {
	Binary      bool
	Description string
	Path        string
	Content     string
	Elements    []Element
}

func (f *File) Name() string { return filepath.Base(f.Path) }

type Directory struct {
	Module bool
	Path   string
	Subs   map[string]*Directory
	Files  map[string]*File
}

type Project struct {
	Dirs  map[string]*Directory
	Files map[string]*File
}

func NewProject(path string) (*Project, error) {
	fileInfo, err := os.Stat(path)
	if err != nil {
		return nil, err
	}
	if !fileInfo.IsDir() {
		return nil, fmt.Errorf("path is not a directory")
	}

	project := &Project{
		Dirs:  make(map[string]*Directory),
		Files: make(map[string]*File),
	}

	files, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	for _, file := range files {
		// skip hidden files for now
		if strings.HasPrefix(file.Name(), ".") {
			continue
		}

		childPath := filepath.Join(path, file.Name())
		childInfo, err := os.Stat(childPath)
		if err != nil {
			return nil, err
		}
		if childInfo.IsDir() {
			m, err := NewDirectory(childPath)
			if err != nil {
				return nil, err
			}
			project.Dirs[childPath] = m
			continue
		}
		f, err := NewFile(childPath)
		if err != nil {
			return nil, err
		}
		project.Files[childPath] = f
	}

	return project, nil
}

func (p *Project) Tree() string {
	res := make([]string, 0)
	for _, d := range p.Dirs {
		res = append(res, d.Tree(2)...)
	}
	for _, f := range p.Files {
		res = append(res, f.Name())
	}

	return strings.Join(res, "\n")
}

func NewFile(path string) (*File, error) {
	file := &File{
		Path: path,
	}
	txt, err := readTextFile(path)
	switch {
	case errors.Is(err, ErrNotATextFile):
		file.Binary = true
	case err != nil:
		return nil, err
	default:
		file.Binary = false
		file.Content = txt
	}

	return file, nil
}

func NewDirectory(path string) (*Directory, error) {
	dir := &Directory{
		Path:  path,
		Subs:  make(map[string]*Directory),
		Files: make(map[string]*File),
	}

	paths, err := os.ReadDir(path)
	if err != nil {
		return nil, err
	}
	for _, p := range paths {
		childPath := filepath.Join(path, p.Name())
		childInfo, err := os.Stat(childPath)
		if err != nil {
			return nil, err
		}
		if childInfo.IsDir() {
			d, err := NewDirectory(childPath)
			if err != nil {
				return nil, err
			}
			dir.Subs[childPath] = d
			continue
		}
		f, err := NewFile(childPath)
		if err != nil {
			return nil, err
		}
		dir.Files[childPath] = f
	}

	return dir, nil
}

func (d *Directory) Tree(indent int) []string {
	in := ""
	for i := 0; i < indent; i++ {
		in += " "
	}
	res := []string{d.Path}
	for _, d := range d.Subs {
		res = append(res, d.Tree(indent+2)...)
	}
	for _, f := range d.Files {
		res = append(res, fmt.Sprintf("%s%s", in, f.Name()))
	}

	return res

}

func readTextFile(path string) (string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return "", err
	}

	for i := 0; i < len(data); i++ {
		r, size := utf8.DecodeRune(data[i:])
		i += size - 1
		if r == utf8.RuneError && !strings.ContainsRune("\r\n\t ", r) {
			return "", ErrNotATextFile
		}
	}

	return string(data), nil
}
