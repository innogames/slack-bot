package storage

import (
	"fmt"
	"io/ioutil"
	"path/filepath"
	"strings"

	scribble "github.com/nanobox-io/golang-scribble"
)

func newFileStorage(dir string) (storage, error) {
	driver, err := scribble.New(dir, &scribble.Options{})

	return &fileStorage{driver, dir}, err
}

type fileStorage struct {
	*scribble.Driver
	dir string
}

func (s fileStorage) GetKeys(collection string) ([]string, error) {
	// todo check security by passing malformatted keys/collections into scribble
	dir := filepath.Join(s.dir, collection)
	files, _ := ioutil.ReadDir(dir)

	if len(files) == 0 {
		return nil, fmt.Errorf("collection is empty")
	}

	keys := make([]string, 0, len(files))
	for _, file := range files {
		keys = append(keys, strings.TrimSuffix(file.Name(), ".json"))
	}

	return keys, nil
}
