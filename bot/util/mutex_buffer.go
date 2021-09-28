package util

import (
	"bytes"
	"sync"
)

// MutexBuffer is a goroutine safe bytes.Buffer
type MutexBuffer struct {
	buffer bytes.Buffer
	mutex  sync.RWMutex
}

// Write appends the contents of p to the buffer, growing the buffer as needed. It returns
// the number of bytes written.
func (s *MutexBuffer) Write(p []byte) (n int, err error) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	return s.buffer.Write(p)
}

// Write returns the current buffer
func (s *MutexBuffer) Read(p []byte) (n int, err error) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.buffer.Read(p)
}

// String returns the contents of the unread portion of the buffer
// as a string.  If the MutexBuffer is a nil pointer, it returns "<nil>".
func (s *MutexBuffer) String() string {
	s.mutex.RLock()
	defer s.mutex.RUnlock()

	return s.buffer.String()
}
