package matcher

import (
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
