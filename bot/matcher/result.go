package matcher

import (
	"regexp"
	"strconv"

	"github.com/innogames/slack-bot/bot/util"
)

// Result is passed into a matched Command and contains the full command of the user and parameters
type Result interface {
	GetString(key string) string
	GetInt(key string) int
	MatchedString() string
	Matched() bool
}

// MapResult implements the Result interface and is a wrapper for map[string]string
type MapResult map[string]string

// GetString returns a parameter, casted as string
func (m MapResult) GetString(key string) string {
	return m[key]
}

// GetInt returns a parameter, casted as int
func (m MapResult) GetInt(key string) int {
	number, _ := strconv.Atoi(m[key])
	return number
}

// MatchedString will return the full command, provided by the user
func (m MapResult) MatchedString() string {
	return m[util.FullMatch]
}

// Matched will return true if a given commands matches against a Matcher
func (m MapResult) Matched() bool {
	return len(m) > 0
}

// ReResult is a regexp based result which is returned by RegexpMatcher
type ReResult struct {
	match []string
	re    *regexp.Regexp
}

// GetString returns a parameter, casted as string
func (m ReResult) GetString(key string) string {
	for idx, name := range m.re.SubexpNames() {
		if name == key {
			return m.match[idx]
		}
	}

	return ""
}

// GetInt returns a parameter, casted as int
func (m ReResult) GetInt(key string) int {
	number, _ := strconv.Atoi(m.GetString(key))
	return number
}

// MatchedString will return the full command, provided by the user
func (m ReResult) MatchedString() string {
	value := m.GetString(util.FullMatch)
	if value != "" {
		return value
	}

	return m.match[0]
}

// Matched will return true if a given commands matches against a Matcher
func (m ReResult) Matched() bool {
	return len(m.match) > 0
}
