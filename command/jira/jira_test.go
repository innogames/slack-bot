package jira

import (
	"testing"

	"github.com/andygrunwald/go-jira"
	"github.com/stretchr/testify/assert"
)

func TestJira(t *testing.T) {
	t.Run("Id", func(t *testing.T) {
		assert.Equal(t, ":question:", idToIcon(nil))
		assert.Equal(t, ":jira_blocker:", idToIcon(&jira.Priority{Name: "Blocker"}))
		assert.Equal(t, ":jira_minor:", idToIcon(&jira.Priority{Name: "Minor"}))
	})
}
