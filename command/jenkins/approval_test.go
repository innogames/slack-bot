package jenkins

import (
	"testing"
	"time"

	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/command/jenkins/client"
	"github.com/stretchr/testify/assert"
)

func TestApprovalStore(t *testing.T) {
	store := newApprovalStore()

	t.Run("add and get", func(t *testing.T) {
		approval := &pendingApproval{
			id:        "abc123",
			jobName:   "TestJob",
			jobConfig: config.JobConfig{},
			params:    client.Parameters{"BRANCH": "master"},
			message:   msg.Message{},
			createdAt: time.Now(),
			expiresAt: time.Now().Add(5 * time.Minute),
		}
		store.add(approval)

		got := store.get("abc123")
		assert.NotNil(t, got)
		assert.Equal(t, "TestJob", got.jobName)
		assert.Equal(t, "master", got.params["BRANCH"])
	})

	t.Run("get non-existent", func(t *testing.T) {
		got := store.get("doesnotexist")
		assert.Nil(t, got)
	})

	t.Run("get expired", func(t *testing.T) {
		approval := &pendingApproval{
			id:        "expired1",
			jobName:   "ExpiredJob",
			jobConfig: config.JobConfig{},
			params:    client.Parameters{},
			message:   msg.Message{},
			createdAt: time.Now().Add(-10 * time.Minute),
			expiresAt: time.Now().Add(-1 * time.Minute),
		}
		store.add(approval)

		got := store.get("expired1")
		assert.Nil(t, got)
	})

	t.Run("remove", func(t *testing.T) {
		approval := &pendingApproval{
			id:        "toremove",
			jobName:   "RemoveJob",
			jobConfig: config.JobConfig{},
			params:    client.Parameters{},
			message:   msg.Message{},
			createdAt: time.Now(),
			expiresAt: time.Now().Add(5 * time.Minute),
		}
		store.add(approval)

		got := store.get("toremove")
		assert.NotNil(t, got)

		store.remove("toremove")

		got = store.get("toremove")
		assert.Nil(t, got)
	})

	t.Run("cleanup removes expired", func(t *testing.T) {
		cleanStore := newApprovalStore()

		cleanStore.add(&pendingApproval{
			id:        "valid1",
			expiresAt: time.Now().Add(5 * time.Minute),
		})
		cleanStore.add(&pendingApproval{
			id:        "expired2",
			expiresAt: time.Now().Add(-1 * time.Minute),
		})
		cleanStore.add(&pendingApproval{
			id:        "expired3",
			expiresAt: time.Now().Add(-5 * time.Minute),
		})

		cleanStore.cleanup()

		assert.NotNil(t, cleanStore.get("valid1"))
		assert.Nil(t, cleanStore.get("expired2"))
		assert.Nil(t, cleanStore.get("expired3"))
	})
}

func TestGenerateApprovalID(t *testing.T) {
	id1 := generateApprovalID()
	id2 := generateApprovalID()

	assert.Len(t, id1, 8)
	assert.Len(t, id2, 8)
	assert.NotEqual(t, id1, id2)
}
