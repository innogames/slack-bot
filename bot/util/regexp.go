package util

import (
	"regexp"
	"strings"
)

// FullMatch is the key of the "Result" which contains the initial full string of the matching process
// todo remove this old shitty magic...
const FullMatch = "match"

// RegexpResultToParams converts a regexp result into a simple string map...kind of deprecated
func RegexpResultToParams(re *regexp.Regexp, match []string) map[string]string {
	params := make(map[string]string, len(match))
	if match == nil {
		return params
	}

	for i, name := range re.SubexpNames() {
		if i == 0 || len(match) < i {
			// store the full match as well
			name = FullMatch
		}
		params[name] = match[i]
	}

	return params
}

// CompileRegexp compiles a regexp to a unified regexp layout:
// - case insensitive
// - always match a full line -> implicitly adds "^" and "$"
func CompileRegexp(pattern string) *regexp.Regexp {
	if pattern == "" {
		return nil
	}

	// add ^...$ to all regexp when not given
	if !strings.HasPrefix(pattern, "^") {
		pattern = "^" + pattern
	}
	if !strings.HasSuffix(pattern, "$") {
		pattern += "$"
	}

	// make all regexp case insensitive by default
	if !strings.Contains(pattern, "(?i)") {
		pattern = "(?i)" + pattern
	}

	return regexp.MustCompile(pattern)
}
