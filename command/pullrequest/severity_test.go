package pullrequest

import (
	"net/http"
	"net/http/httptest"
	"testing"

	gojira "github.com/andygrunwald/go-jira"
	"github.com/innogames/slack-bot/v2/bot/config"
	"github.com/innogames/slack-bot/v2/bot/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExtractTicketKey(t *testing.T) {
	testCases := []struct {
		name     string
		pr       pullRequest
		expected string
	}{
		{
			name:     "key in branch is preferred",
			pr:       pullRequest{Branch: "bugfix/TEST-123-fix-xyz", Name: "PROJ-999: unrelated title"},
			expected: "TEST-123",
		},
		{
			name:     "key at the start of the branch",
			pr:       pullRequest{Branch: "TEST-123-bugfix"},
			expected: "TEST-123",
		},
		{
			name:     "fallback to title when branch has no key",
			pr:       pullRequest{Branch: "feature/some-cleanup", Name: "TEST-456: add feature"},
			expected: "TEST-456",
		},
		{
			name:     "no key at all",
			pr:       pullRequest{Branch: "feature/cleanup", Name: "just a title"},
			expected: "",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.expected, extractTicketKey(tc.pr))
		})
	}
}

func TestGetSeverityReaction(t *testing.T) {
	priorityReactions := map[string]util.Reaction{
		"Blocker": "jira_blocker",
		"Major":   "jira_major",
	}

	t.Run("no jira client returns empty", func(t *testing.T) {
		cmd := command{cfg: config.PullRequest{JiraPriorityReactions: priorityReactions}}
		assert.Empty(t, cmd.getSeverityReaction(pullRequest{Branch: "TEST-1-fix"}))
	})

	// mock Jira server returning a fixed priority for any issue request
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/rest/api/2/issue/TEST-123")
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"key":"TEST-123","fields":{"priority":{"name":"Blocker"}}}`))
	}))
	defer server.Close()

	jiraClient, err := gojira.NewClient(http.DefaultClient, server.URL)
	require.NoError(t, err)

	cmd := command{
		cfg:  config.PullRequest{JiraPriorityReactions: priorityReactions},
		jira: jiraClient,
	}

	t.Run("maps priority to reaction", func(t *testing.T) {
		reaction := cmd.getSeverityReaction(pullRequest{Branch: "bugfix/TEST-123-fix"})
		assert.Equal(t, util.Reaction("jira_blocker"), reaction)
	})

	t.Run("no ticket key returns empty", func(t *testing.T) {
		assert.Empty(t, cmd.getSeverityReaction(pullRequest{Branch: "feature/cleanup", Name: "no key"}))
	})
}

func TestGetSeverityReactionUnmappedPriority(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"key":"TEST-123","fields":{"priority":{"name":"Trivial"}}}`))
	}))
	defer server.Close()

	jiraClient, err := gojira.NewClient(http.DefaultClient, server.URL)
	require.NoError(t, err)

	cmd := command{
		cfg: config.PullRequest{JiraPriorityReactions: map[string]util.Reaction{
			"Blocker": "jira_blocker",
		}},
		jira: jiraClient,
	}

	assert.Empty(t, cmd.getSeverityReaction(pullRequest{Branch: "TEST-123-fix"}))
}
