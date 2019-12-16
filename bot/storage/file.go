package storage

import (
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
	dir := filepath.Join(s.dir, collection)

	// read all the files in the transaction.Collection; an error here just means
	// the collection is either empty or doesn't exist
	files, _ := ioutil.ReadDir(dir)

	keys := make([]string, 0, len(files))
	for _, file := range files {
		keys = append(keys, strings.TrimSuffix(file.Name(), ".json"))
	}

	return keys, nil
}
