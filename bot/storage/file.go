package storage

import (
	"os"
	"path/filepath"
	"strings"

	scribble "github.com/sdomino/golang-scribble"
)

func newFileStorage(dir string) (Storage, error) {
	driver, err := scribble.New(dir, &scribble.Options{})

	return &fileStorage{driver, dir}, err
}

type fileStorage struct {
	*scribble.Driver
	dir string
}

func (s *fileStorage) GetKeys(collection string) ([]string, error) {
	dir := filepath.Join(s.dir, collection)
	files, _ := os.ReadDir(dir)

	keys := make([]string, 0, len(files))

	for _, file := range files {
		keys = append(keys, strings.TrimSuffix(file.Name(), ".json"))
	}

	return keys, nil
}
