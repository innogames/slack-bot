package vcs

type null struct{}

// LoadBranches will just return nothing
func (f null) LoadBranches() ([]string, error) {
	var branchNames []string

	return branchNames, nil
}
