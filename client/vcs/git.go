package vcs

import (
	"os/exec"
	"regexp"
)

// git fetcher to load all branch names from a remote repository
type git struct {
	repoURL string
}

var gitBranchRe = regexp.MustCompile(`refs/(remotes/origin|heads)/(.*)\n`)

// LoadBranches will load the branches from a (remote) git repository
func (f git) LoadBranches() (branchNames []string, err error) {
	/* #nosec */
	cmd := exec.Command("git", "ls-remote", "--refs", f.repoURL) //nolint:gosec
	output, err := cmd.Output()
	if err != nil {
		return
	}

	for _, match := range gitBranchRe.FindAllStringSubmatch(string(output), -1) {
		branchNames = append(branchNames, match[2])
	}

	return
}
