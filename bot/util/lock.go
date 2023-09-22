package util

import "sync"

func NewGroupedLogger[T string]() GroupedLock[T] {
	return GroupedLock[T]{
		locks:      make(map[T]*sync.RWMutex),
		globalLock: sync.Mutex{},
	}
}

type GroupedLock[T string] struct {
	globalLock sync.Mutex
	locks      map[T]*sync.RWMutex
}

func (l *GroupedLock[T]) GetLock(name T) sync.Locker {
	lock := l.getLock(name)

	lock.Lock()

	return lock
}

func (l *GroupedLock[T]) GetRLock(name T) sync.Locker {
	lock := l.getLock(name)

	lock.RLock()

	return lock.RLocker()
}

func (l *GroupedLock[T]) getLock(name T) *sync.RWMutex {
	l.globalLock.Lock()
	defer l.globalLock.Unlock()

	lock, ok := l.locks[name]
	if !ok {
		lock = &sync.RWMutex{}
		l.locks[name] = lock
	}

	return lock
}
