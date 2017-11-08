package bot

// Help command can provide help objects which are searchable by keywords
type Help struct {
	Command     string
	Description string
	Examples    []string
}

// GetKeywords crates a string slice of help keywords -> used by fuzzy search
func (h Help) GetKeywords() []string {
	keywords := make([]string, 0, len(h.Examples)+1)
	keywords = append(keywords, h.Command)
	keywords = append(keywords, h.Examples...)

	return keywords
}
