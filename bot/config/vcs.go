package config

type VCS struct {
	Type       string // stash/bitbucket/git/null
	Repository string
}

func (c VCS) IsEnabled() bool {
	return c.Type != ""
}
