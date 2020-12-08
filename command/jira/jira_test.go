package jira

import (
	"github.com/andygrunwald/go-jira"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestJira(t *testing.T) {
	t.Run("Id", func(t *testing.T) {
		assert.Equal(t, ":question:", idToIcon(nil))
		assert.Equal(t, ":jira_blocker:", idToIcon(&jira.Priority{Name: "Blocker"}))
	})
}
