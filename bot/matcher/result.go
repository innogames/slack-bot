package matcher

import (
	"regexp"
	"strconv"
)

// Result implements the Result interface and is a wrapper for map[string]string
type Result map[string]string

// GetString returns a parameter, casted as string
func (m Result) GetString(key string) string {
	return m[key]
}

// GetInt returns a parameter, casted as int
func (m Result) GetInt(key string) int {
	number, _ := strconv.Atoi(m[key])
	return number
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
