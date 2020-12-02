package bot

import "sync"

var globalLock sync.Mutex

// get and hold a mutex for each requested user.
// the user-lock it's locked by default
func (b *Bot) getUserLock(userID string) *sync.Mutex {
	var userLock *sync.Mutex

	globalLock.Lock()
	userLock, ok := b.userLocks[userID]
	if !ok {
		userLock = &sync.Mutex{}
		b.userLocks[userID] = userLock
	}

	userLock.Lock()
	globalLock.Unlock()

	return userLock
}
