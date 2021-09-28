package bot

import "sync"

var (
	globalLock sync.Mutex
	userLocks  = map[string]*sync.Mutex{}
)

// get and hold a mutex for each requested user.
// the user-lock it's locked by default
func getUserLock(userID string) *sync.Mutex {
	var userLock *sync.Mutex

	globalLock.Lock()
	defer globalLock.Unlock()

	userLock, ok := userLocks[userID]
	if !ok {
		userLock = &sync.Mutex{}
		userLocks[userID] = userLock
	}

	userLock.Lock()

	return userLock
}
