package jira

import (
	"testing"

	"github.com/andygrunwald/go-jira"
	"github.com/stretchr/testify/assert"
)

// errTicketNotAccessibleRe matches the error message for a non existing ticket.
// The public Apache Jira answers with a 404 or - for anonymous users - with a 401,
// so both variants are accepted here.
const errTicketNotAccessibleRe = `^(Issue Does Not Exist|You do not have the permission to see the specified issue\.): request failed\.`

func TestJira(t *testing.T) {
	t.Run("Id", func(t *testing.T) {
		assert.Equal(t, ":question:", idToIcon(nil))
		assert.Equal(t, ":jira_blocker:", idToIcon(&jira.Priority{Name: "Blocker"}))
		assert.Equal(t, ":jira_minor:", idToIcon(&jira.Priority{Name: "Minor"}))
	})
}
