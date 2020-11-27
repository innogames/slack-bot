package bot

// Help command can provide help objects which are searchable by keywords
type Help struct {
	Command     string
	Description string
	HelpURL     string
	Category    Category
	Examples    []string
}

// GetKeywords crates a string slice of help keywords -> used by fuzzy search
func (h *Help) GetKeywords() []string {
	keywords := make([]string, 0, len(h.Examples)+1)
	keywords = append(keywords, h.Command)
	keywords = append(keywords, h.Examples...)

	if h.Category.Name != "" {
		keywords = append(keywords, h.Category.Name)
	}

	return keywords
}

// Category of Help entries. -> Groups command in help command by "Jenkins", "Pull request" etc
type Category struct {
	Name        string
	Description string
	HelpURL     string
}
