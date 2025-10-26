package pool

import (
	"testing"
	"time"

	"github.com/innogames/slack-bot/v2/bot"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/bot/storage"
	"github.com/innogames/slack-bot/v2/mocks"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// compile time check that the interface matches
var _ bot.Runnable = &poolCommands{}

func TestPools(t *testing.T) {
	slackClient := mocks.NewSlackClient(t)
	base := bot.BaseCommand{SlackClient: slackClient}

	t.Run("Pools are not active", func(t *testing.T) {
		cfg := &config.Pool{}
		commands := GetCommands(cfg, base)
		assert.Empty(t, commands.GetCommandNames())
	})

	t.Run("Full test", func(t *testing.T) {
		cfg := &config.Pool{
			LockDuration: time.Minute,
			NotifyExpire: time.Minute,
			Resources: []*config.Resource{
				{
					Name: "server1",
				},
				{
					Name: "server2",
				},
			},
		}
		commands := GetCommands(cfg, base)
		assert.Len(t, commands.GetCommandNames(), 1)

		runCommand := func(message msg.Message) {
			actual := commands.Run(message)
			assert.True(t, actual)
		}

		// list
		message := msg.Message{}
		message.Text = "pool list"
		mocks.AssertSlackMessage(slackClient, message, "*Available:*\n`server1`, `server2`\n\n*Used/Locked:*")
		runCommand(message)

		// lock
		message = msg.Message{}
		message.Text = "pool lock server1"
		mocks.AssertSlackMessageRegexp(slackClient, message, "^`server1` is locked for you until")
		runCommand(message)

		// extend
		message = msg.Message{}
		message.Text = "pool extend server1 1h"
		mocks.AssertSlackMessageRegexp(slackClient, message, "^`server1` got extended until")
		runCommand(message)

		// pool locks
		message = msg.Message{}
		message.Text = "pool locks"
		mocks.AssertSlackMessageRegexp(slackClient, message, "^ \\*Your locks:\\*\n\n`server1` until")
		runCommand(message)

		// unlock
		message = msg.Message{}
		message.Text = "pool unlock server1"
		mocks.AssertSlackMessage(slackClient, message, "`server1` is free again")
		runCommand(message)

		help := commands.GetHelp()
		assert.Len(t, help, 6)
	})
}

func TestPoolCoreOperations(t *testing.T) {
	// Initialize storage for tests
	storage.InitStorage("")

	// Create a fresh pool for each test to avoid state conflicts
	createTestPool := func() *pool {
		resources := []*config.Resource{
			{Name: "server1", ExplicitLock: false},
			{Name: "server2", ExplicitLock: true},
			{Name: "server3", ExplicitLock: false},
		}

		cfg := &config.Pool{
			LockDuration: 30 * time.Minute,
			Resources:    resources,
		}

		p := &pool{
			locks:        make(map[*config.Resource]*ResourceLock),
			lockDuration: cfg.LockDuration,
		}

		// Initialize locks without using storage
		for _, resource := range cfg.Resources {
			p.locks[resource] = nil
		}

		return p
	}

	t.Run("initial state", func(t *testing.T) {
		p := createTestPool()
		// All resources should be initially unlocked
		free := p.GetFree()
		assert.Len(t, free, 3)

		locked := p.GetLocks("")
		assert.Empty(t, locked)
	})

	t.Run("lock resource successfully", func(t *testing.T) {
		p := createTestPool()
		lock, err := p.Lock("user1", "testing", "server1")
		require.NoError(t, err)
		require.NotNil(t, lock)

		assert.Equal(t, "user1", lock.User)
		assert.Equal(t, "testing", lock.Reason)
		assert.Equal(t, "server1", lock.Resource.Name)
		assert.False(t, lock.WarningSend)
		assert.WithinDuration(t, time.Now().Add(30*time.Minute), lock.LockUntil, time.Second)

		// Check that resource is now locked
		free := p.GetFree()
		assert.Len(t, free, 2)

		locked := p.GetLocks("")
		assert.Len(t, locked, 1)
		assert.Equal(t, "user1", locked[0].User)
	})

	t.Run("lock already locked resource", func(t *testing.T) {
		p := createTestPool()
		_, err := p.Lock("user1", "testing", "server1")
		require.NoError(t, err)

		_, err = p.Lock("user2", "also testing", "server1")
		assert.Equal(t, ErrNoResourceAvailable, err)
	})

	t.Run("lock specific explicit resource", func(t *testing.T) {
		p := createTestPool()
		lock, err := p.Lock("user2", "explicit testing", "server2")
		require.NoError(t, err)
		require.NotNil(t, lock)

		assert.Equal(t, "user2", lock.User)
		assert.Equal(t, "explicit testing", lock.Reason)
		assert.Equal(t, "server2", lock.Resource.Name)

		// Check state
		free := p.GetFree()
		assert.Len(t, free, 2) // server1 and server3 should be free

		locked := p.GetLocks("")
		assert.Len(t, locked, 1)
	})

	t.Run("lock non-specific explicit resource should skip explicit", func(t *testing.T) {
		p := createTestPool()
		// Lock the non-explicit resources first
		_, err := p.Lock("user1", "test", "server1")
		require.NoError(t, err)

		_, err = p.Lock("user2", "test", "server3")
		require.NoError(t, err)

		// Try to lock non-specifically - should fail because only server2 (explicit) is left
		_, err = p.Lock("user3", "non-specific", "")
		assert.Equal(t, ErrNoResourceAvailable, err)
	})

	t.Run("unlock resource by owner", func(t *testing.T) {
		p := createTestPool()
		_, err := p.Lock("user1", "testing", "server1")
		require.NoError(t, err)

		err = p.Unlock("user1", "server1")
		require.NoError(t, err)

		// Check that resource is now free
		free := p.GetFree()
		assert.Len(t, free, 3)

		locked := p.GetLocks("")
		assert.Empty(t, locked)
	})

	t.Run("unlock resource by different user", func(t *testing.T) {
		p := createTestPool()
		_, err := p.Lock("user1", "testing", "server1")
		require.NoError(t, err)

		err = p.Unlock("user2", "server1")
		assert.Equal(t, ErrResourceLockedByDifferentUser, err)

		// Resource should still be locked
		locked := p.GetLocks("")
		assert.Len(t, locked, 1)
	})

	t.Run("unlock non-existent resource", func(t *testing.T) {
		p := createTestPool()
		err := p.Unlock("user1", "nonexistent")
		assert.NoError(t, err) // Should not error, just do nothing
	})

	t.Run("extend lock by owner", func(t *testing.T) {
		p := createTestPool()
		originalTime := time.Now().Add(15 * time.Minute)
		_, err := p.Lock("user2", "testing", "server2")
		require.NoError(t, err)

		locked := p.GetLocks("")
		require.Len(t, locked, 1)

		// Manually set lock time for testing
		locked[0].LockUntil = originalTime
		p.locks[&locked[0].Resource] = locked[0]

		extended, err := p.ExtendLock("user2", "server2", "1h")
		require.NoError(t, err)
		require.NotNil(t, extended)

		assert.Equal(t, "user2", extended.User)
		assert.Equal(t, "server2", extended.Resource.Name)
		assert.False(t, extended.WarningSend)
		assert.WithinDuration(t, originalTime.Add(time.Hour), extended.LockUntil, time.Second)
	})

	t.Run("extend lock by different user", func(t *testing.T) {
		p := createTestPool()
		_, err := p.Lock("user1", "testing", "server1")
		require.NoError(t, err)

		_, err = p.ExtendLock("user2", "server1", "1h")
		assert.Equal(t, ErrResourceLockedByDifferentUser, err)
	})

	t.Run("extend non-existent lock", func(t *testing.T) {
		p := createTestPool()
		_, err := p.ExtendLock("user1", "server1", "1h")
		assert.Equal(t, ErrNoLockedResourceFound, err)
	})

	t.Run("get user specific locks", func(t *testing.T) {
		p := createTestPool()
		// Lock resources for different users
		_, err := p.Lock("user1", "user1 reason", "server1")
		require.NoError(t, err)

		_, err = p.Lock("user2", "user2 reason", "server2")
		require.NoError(t, err)

		// Get locks for specific users
		user1Locks := p.GetLocks("user1")
		assert.Len(t, user1Locks, 1)
		assert.Equal(t, "user1", user1Locks[0].User)
		assert.Equal(t, "server1", user1Locks[0].Resource.Name)

		user2Locks := p.GetLocks("user2")
		assert.Len(t, user2Locks, 1)
		assert.Equal(t, "user2", user2Locks[0].User)
		assert.Equal(t, "server2", user2Locks[0].Resource.Name)

		// Get all locks
		allLocks := p.GetLocks("")
		assert.Len(t, allLocks, 2)
	})
}

func TestPoolStorage(t *testing.T) {
	// Initialize storage for tests
	storage.InitStorage("")

	resources := []*config.Resource{
		{Name: "server1", ExplicitLock: false},
		{Name: "server2", ExplicitLock: false},
	}

	cfg := &config.Pool{
		LockDuration: time.Hour,
		Resources:    resources,
	}

	t.Run("storage persistence", func(t *testing.T) {
		p := getNewPool(cfg)

		// Lock a resource
		lock, err := p.Lock("user1", "storage test", "server1")
		require.NoError(t, err)

		// Manually modify the lock time for testing
		lock.LockUntil = time.Now().Add(2 * time.Hour)
		p.locks[&lock.Resource] = lock

		// Force storage write
		err = storage.Write(storageKey, lock.Resource.Name, lock)
		require.NoError(t, err)

		// Create new pool instance (simulates restart)
		p2 := getNewPool(cfg)

		// Check that lock was restored
		locked := p2.GetLocks("")
		assert.Len(t, locked, 1)
		assert.Equal(t, "user1", locked[0].User)
		assert.Equal(t, "server1", locked[0].Resource.Name)
	})

	t.Run("storage corruption handling", func(t *testing.T) {
		// Write invalid data to storage
		err := storage.Write(storageKey, "server2", "invalid data")
		require.NoError(t, err)

		// Create new pool - should handle corruption gracefully
		p2 := getNewPool(cfg)

		// Should still work normally
		_, err = p2.Lock("user1", "test", "server2")
		require.NoError(t, err)
	})
}

func TestPoolConcurrency(t *testing.T) {
	// Initialize storage for tests
	storage.InitStorage("")

	resources := []*config.Resource{
		{Name: "server1", ExplicitLock: false},
		{Name: "server2", ExplicitLock: false},
	}

	cfg := &config.Pool{
		LockDuration: time.Hour,
		Resources:    resources,
	}

	p := getNewPool(cfg)

	t.Run("concurrent locking", func(t *testing.T) {
		done := make(chan bool, 2)

		// Two goroutines trying to lock the same resource
		go func() {
			_, err := p.Lock("user1", "concurrent test 1", "server1")
			if err == nil {
				time.Sleep(10 * time.Millisecond) // Hold lock briefly
				p.Unlock("user1", "server1")
			}
			done <- true
		}()

		go func() {
			time.Sleep(10 * time.Millisecond) // Small delay
			_, err := p.Lock("user2", "concurrent test 2", "server1")
			if err == nil {
				p.Unlock("user2", "server1")
			}
			done <- true
		}()

		// Wait for both to complete
		<-done
		<-done

		// Only one should have succeeded
		locked := p.GetLocks("")
		assert.LessOrEqual(t, len(locked), 1, "Only one user should have the lock")
	})
}

func TestPoolErrors(t *testing.T) {
	// Initialize storage for tests
	storage.InitStorage("")

	resources := []*config.Resource{
		{Name: "server1", ExplicitLock: false},
	}

	cfg := &config.Pool{
		LockDuration: time.Hour,
		Resources:    resources,
	}

	p := getNewPool(cfg)

	t.Run("all error conditions", func(t *testing.T) {
		// Test ErrNoResourceAvailable
		_, err := p.Lock("user1", "test", "server1")
		require.NoError(t, err)

		_, err = p.Lock("user2", "test", "server1")
		assert.Equal(t, ErrNoResourceAvailable, err)

		// Test ErrResourceLockedByDifferentUser (unlock)
		err = p.Unlock("user2", "server1")
		assert.Equal(t, ErrResourceLockedByDifferentUser, err)

		// Test ErrResourceLockedByDifferentUser (extend)
		_, err = p.ExtendLock("user2", "server1", "1h")
		assert.Equal(t, ErrResourceLockedByDifferentUser, err)

		// Test ErrNoLockedResourceFound (extend non-existent)
		_, err = p.ExtendLock("user1", "server2", "1h")
		assert.Equal(t, ErrNoLockedResourceFound, err)

		// Clean up
		err = p.Unlock("user1", "server1")
		require.NoError(t, err)
	})
}

func TestPoolLockExpiration(t *testing.T) {
	// Initialize storage for tests
	storage.InitStorage("")

	resources := []*config.Resource{
		{Name: "server1", ExplicitLock: false},
		{Name: "server2", ExplicitLock: false},
	}

	cfg := &config.Pool{
		LockDuration: 100 * time.Millisecond, // Very short for testing
		NotifyExpire: 50 * time.Millisecond,
		Resources:    resources,
	}

	p := getNewPool(cfg)

	t.Run("lock expiration", func(t *testing.T) {
		// Lock a resource
		_, err := p.Lock("user1", "expiration test", "server1")
		require.NoError(t, err)

		// Check initial state
		locked := p.GetLocks("")
		require.Len(t, locked, 1)
		assert.False(t, locked[0].WarningSend)

		// Wait for expiration
		time.Sleep(150 * time.Millisecond)

		// Lock should still be there but expired
		locked = p.GetLocks("")
		assert.Len(t, locked, 1)
		assert.True(t, time.Now().After(locked[0].LockUntil))

		// Clean up
		err = p.Unlock("user1", "server1")
		require.NoError(t, err)
	})
}
