package jenkins

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"

	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/msg"
	"github.com/innogames/slack-bot/v2/command/jenkins/client"
)

type pendingApproval struct {
	id        string
	jobName   string
	jobConfig config.JobConfig
	params    client.Parameters
	message   msg.Message // original message for posting results back to the original channel
	createdAt time.Time
	expiresAt time.Time
}

type approvalStore struct {
	mu      sync.Mutex
	pending map[string]*pendingApproval
}

func newApprovalStore() *approvalStore {
	return &approvalStore{
		pending: make(map[string]*pendingApproval),
	}
}

func (s *approvalStore) add(approval *pendingApproval) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.pending[approval.id] = approval
}

// get retrieves a pending approval by ID. Returns nil if not found or expired.
func (s *approvalStore) get(id string) *pendingApproval {
	s.mu.Lock()
	defer s.mu.Unlock()

	approval, ok := s.pending[id]
	if !ok {
		return nil
	}

	if time.Now().After(approval.expiresAt) {
		delete(s.pending, id)
		return nil
	}

	return approval
}

func (s *approvalStore) remove(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.pending, id)
}

func (s *approvalStore) cleanup() {
	s.mu.Lock()
	defer s.mu.Unlock()

	now := time.Now()
	for id, approval := range s.pending {
		if now.After(approval.expiresAt) {
			delete(s.pending, id)
		}
	}
}

func generateApprovalID() string {
	b := make([]byte, 4)
	_, _ = rand.Read(b)
	return hex.EncodeToString(b)
}
