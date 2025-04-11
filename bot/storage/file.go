package storage

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/innogames/slack-bot/v2/bot/util"
)

func newFileStorage(dir string) (Storage, error) {
	driver := &fileStorage{
		dir:   dir,
		locks: util.NewGroupedLogger(),
	}

	// if the database already exists, just use it
	if _, err := os.Stat(dir); err == nil {
		return driver, nil
	}

	// if the database doesn't exist create it
	return driver, os.MkdirAll(dir, 0o755) // #nosec G301
}

type fileStorage struct {
	dir   string
	locks util.GroupedLock[string]
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

// Write locks the database and attempts to write the record to the database under
// the [collection] specified with the [resource] name given
func (s *fileStorage) Write(collection, resource string, v any) error {
	lock := s.locks.GetLock(collection)
	defer lock.Unlock()

	dir := filepath.Join(s.dir, collection)
	fnlPath := filepath.Join(dir, resource+".json")
	tmpPath := fnlPath + ".tmp"

	return write(dir, tmpPath, fnlPath, v)
}

func write(dir, tmpPath, dstPath string, v any) error {
	// create collection directory
	if err := os.MkdirAll(dir, 0o750); err != nil { // #nosec G301
		return err
	}

	// marshal the pointer to a non-struct and indent with tab
	b, err := json.MarshalIndent(v, "", "\t")
	if err != nil {
		return err
	}

	// add newline to the end
	b = append(b, byte('\n'))

	// write marshaled data to the temp file
	if err := os.WriteFile(tmpPath, b, 0o600); err != nil {
		return err
	}

	// move final file into place
	return os.Rename(tmpPath, dstPath)
}

// Read a record from the database
func (s *fileStorage) Read(collection, resource string, v any) error {
	record := filepath.Join(s.dir, collection, resource+".json")

	b, err := os.ReadFile(record)
	if err != nil {
		return err
	}

	return json.Unmarshal(b, v)
}

// Delete locks the database then attempts to remove the collection/resource
// specified by [path]
func (s *fileStorage) Delete(collection, resource string) error {
	lock := s.locks.GetLock(collection)
	defer lock.Unlock()

	path := filepath.Join(collection, resource)
	dir := filepath.Join(s.dir, path)

	switch fi, err := stat(dir); {
	// if fi is nil or error is not nil return
	case fi == nil, err != nil:
		return fmt.Errorf("unable to find file named %v", path)

	// remove file
	case fi.Mode().IsRegular():
		return os.RemoveAll(dir + ".json")
	}

	return nil
}

func stat(path string) (fi os.FileInfo, err error) {
	// check for dir, if path isn't a directory check to see if it's a file
	if fi, err = os.Stat(path); os.IsNotExist(err) {
		fi, err = os.Stat(path + ".json")
	}

	return
}
