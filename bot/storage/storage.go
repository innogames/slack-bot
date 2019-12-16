package storage

var currentStorage storage

type storage interface {
	Write(collection, resource string, v interface{}) error
	Read(collection, resource string, v interface{}) error
	GetKeys(collection string) ([]string, error)
	Delete(collection, resource string) error
}

// InitStorage registers a local directory as JSON file storage
func InitStorage(path string) (storage, error) {
	var err error
	if path == "" {
		currentStorage = newMemoryStorage()
	} else {
		currentStorage, err = newFileStorage(path)
	}

	return currentStorage, err
}

// Write stores one value in the persistent storage
func Write(collection string, resource string, v interface{}) error {
	return getStorage().Write(collection, resource, v)
}

// Read will load the stored data for one entry (using reference) to avoid allocation
func Read(collection string, resource string, v interface{}) error {
	return getStorage().Read(collection, resource, v)
}

// GetKeys will return the (json) strings of a collection
func GetKeys(collection string) ([]string, error) {
	return getStorage().GetKeys(collection)
}

// Delete returns one entry
func Delete(collection string, resource string) error {
	return getStorage().Delete(collection, resource)
}

func getStorage() storage {
	if currentStorage == nil {
		currentStorage = newMemoryStorage()
	}

	return currentStorage
}
