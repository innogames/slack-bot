package pool

import (
	"fmt"
	"sort"
	"sync"
	"time"

	"github.com/innogames/slack-bot/v2/bot/util"

	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/storage"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
)

const (
	storageKey = "pool"
)

var (
	ErrResourceLockedByDifferentUser = fmt.Errorf("resources locked by different user")
	ErrNoLockedResourceFound         = fmt.Errorf("no locked resource found")
	ErrNoResourceAvailable           = fmt.Errorf("no resource available")
)

// ResourceLock struct to hold and store the current locks
type ResourceLock struct {
	Resource    config.Resource `json:"-"`
	User        string
	Reason      string
	WarningSend bool `json:"-"`
	LockUntil   time.Time
}

type pool struct {
	locks        map[*config.Resource]*ResourceLock
	lockDuration time.Duration
	mu           sync.RWMutex
}

// getNewPool create a new pool and initialize it by the local storage
func getNewPool(cfg *config.Pool) *pool {
	var p pool

	p.lockDuration = cfg.LockDuration

	p.locks = make(map[*config.Resource]*ResourceLock)
	for _, resource := range cfg.Resources {
		p.locks[resource] = nil
	}

	keys, _ := storage.GetKeys(storageKey)
	for _, key := range keys {
		var lock ResourceLock
		if err := storage.Read(storageKey, key, &lock); err != nil {
			log.Errorf("[Pool] unable to restore lock for '%s': %s", key, err)
			continue
		}

		for k := range p.locks {
			if k.Name == key {
				lock.Resource = *k
				p.locks[k] = &lock
				break
			}
		}
	}
	return &p
}

// Lock a resource in the pool for a user
func (p *pool) Lock(user, reason, resourceName string) (*ResourceLock, error) {
	specificResource := len(resourceName) > 0

	for k, v := range p.locks {
		if v != nil {
			// it's already in used
			continue
		}

		if !specificResource && k.ExplicitLock {
			// resource can be locked only specifically
			continue
		}

		if specificResource && k.Name != resourceName {
			// specific resource should be locked but it's not this one.
			continue
		}

		resourceLock := &ResourceLock{
			Resource:  *k,
			User:      user,
			Reason:    reason,
			LockUntil: time.Now().Add(p.lockDuration),
		}

		p.mu.Lock()
		defer p.mu.Unlock()

		p.locks[k] = resourceLock

		if err := storage.Write(storageKey, k.Name, resourceLock); err != nil {
			log.Error(errors.Wrap(err, "error while storing pool lock entry"))
		}
		return resourceLock, nil
	}

	return nil, ErrNoResourceAvailable
}

// ExtendLock extends the lock of a resource in the pool for a user
func (p *pool) ExtendLock(user, resourceName, duration string) (*ResourceLock, error) {
	for k, v := range p.locks {
		if v == nil {
			continue
		}

		if k.Name != resourceName {
			continue
		}

		if v.User != user {
			return nil, ErrResourceLockedByDifferentUser
		}

		d, err := util.ParseDuration(duration)
		if err != nil {
			return nil, err
		}

		v.LockUntil = v.LockUntil.Add(d)
		v.WarningSend = false

		p.locks[k] = v

		if err := storage.Delete(storageKey, k.Name); err != nil {
			log.Error(errors.Wrap(err, "error while storing pool lock entry"))
		}
		if err := storage.Write(storageKey, k.Name, v); err != nil {
			log.Error(errors.Wrap(err, "error while storing pool lock entry"))
		}

		return v, nil
	}

	return nil, ErrNoLockedResourceFound
}

// Unlock a resource of a user
func (p *pool) Unlock(user, resourceName string) error {
	for k, v := range p.locks {
		if v == nil {
			continue
		}

		if k.Name != resourceName {
			continue
		}

		if v.User != user {
			return ErrResourceLockedByDifferentUser
		}

		p.locks[k] = nil

		if err := storage.Delete(storageKey, k.Name); err != nil {
			log.Error(errors.Wrap(err, "error while storing pool lock entry"))
		}
	}

	return nil
}

// GetLocks returns a sorted list of all active locks of a user / all users if userName = ""
func (p *pool) GetLocks(userName string) []*ResourceLock {
	var locked []*ResourceLock
	byUser := len(userName) > 0
	for _, v := range p.locks {
		if v != nil && (!byUser || userName == v.User) {
			locked = append(locked, v)
		}
	}
	sort.Slice(locked, func(i, j int) bool {
		return locked[i].Resource.Name < locked[j].Resource.Name
	})

	return locked
}

// GetFree returns a sorted list of all free / unlocked resources
func (p *pool) GetFree() []*config.Resource {
	var free []*config.Resource
	for k, v := range p.locks {
		if v == nil {
			free = append(free, k)
		}
	}
	sort.Slice(free, func(i, j int) bool {
		return free[i].Name < free[j].Name
	})
	return free
}
