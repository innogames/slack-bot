package vcs

import (
	"os/exec"
	"regexp"
)

type git struct {
	repo string
}

var gitBranchRe = regexp.MustCompile(`refs\/(remotes\/origin|heads)\/(.*)\n`)

// LoadBranches will load the branches from a (remote) git repository
func (f git) LoadBranches() ([]string, error) {
	var branchNames []string

	cmd := exec.Command("git", "ls-remote", "--refs", f.repo)
	output, err := cmd.Output()
	if err != nil {
		return branchNames, err
	}

	for _, match := range gitBranchRe.FindAllStringSubmatch(string(output), -1) {
		if match[2] == "HEAD" {
			continue
		}
		branchNames = append(branchNames, match[2])
	}

	return branchNames, err
}
