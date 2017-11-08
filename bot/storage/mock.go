// +build !test

package storage

// MockStorage is only used in tests to work within a fresh storage which gets removed later
func MockStorage() func() error {
	InitStorage("./storage_test")
	DeleteAll()

	return DeleteAll
}
