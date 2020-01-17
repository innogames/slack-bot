package bot

import "sync"

var globalLock sync.Mutex

func (b bot) getUserLock(userId string) *sync.Mutex {
	var userLock *sync.Mutex

	globalLock.Lock()
	userLock, ok := b.userLocks[userId]
	if !ok {
		userLock = &sync.Mutex{}
		b.userLocks[userId] = userLock
	}
	globalLock.Unlock()

	userLock.Lock()

	return userLock
}
