package matcher

import (
	"regexp"
	"strconv"

	"github.com/innogames/slack-bot/bot/util"
)

type Result interface {
	GetString(key string) string
	GetInt(key string) int
	MatchedString() string
	Matched() bool
}

type MapResult map[string]string

func (m MapResult) GetString(key string) string {
	return m[key]
}

func (m MapResult) GetInt(key string) int {
	number, _ := strconv.Atoi(m[key])
	return number
}

func (m MapResult) MatchedString() string {
	return m[util.FullMatch]
}

func (m MapResult) Matched() bool {
	return len(m) > 0
}

type ReResult struct {
	match []string
	re    *regexp.Regexp
}

func (m ReResult) GetString(key string) string {
	for idx, name := range m.re.SubexpNames() {
		if name == key {
			return m.match[idx]
		}
	}

	return ""
}

func (m ReResult) GetInt(key string) int {
	number, _ := strconv.Atoi(m.GetString(key))
	return number
}

func (m ReResult) MatchedString() string {
	value := m.GetString(util.FullMatch)
	if value != "" {
		return value
	}

	return m.match[0]
}

func (m ReResult) Matched() bool {
	return len(m.match) > 0
}
