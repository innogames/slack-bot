package bot

import "sync"

var globalLock sync.Mutex

func (b *Bot) getUserLock(userID string) *sync.Mutex {
	var userLock *sync.Mutex

	globalLock.Lock()
	userLock, ok := b.userLocks[userID]
	if !ok {
		userLock = &sync.Mutex{}
		b.userLocks[userID] = userLock
	}
	globalLock.Unlock()

	userLock.Lock()

	return userLock
}
