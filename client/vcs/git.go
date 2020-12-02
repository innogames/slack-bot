package vcs

import (
	"os/exec"
	"regexp"
)

type git struct {
	repo string
}

var gitBranchRe = regexp.MustCompile(`refs/(remotes/origin|heads)/(.*)\n`)

// LoadBranches will load the branches from a (remote) git repository
func (f git) LoadBranches() (branchNames []string, err error) {
	cmd := exec.Command("git", "ls-remote", "--refs", f.repo)
	output, err := cmd.Output()
	if err != nil {
		return
	}

	for _, match := range gitBranchRe.FindAllSubmatch(output, -1) {
		if string(match[2]) == "HEAD" {
			continue
		}
		branchNames = append(branchNames, string(match[2]))
	}

	return
}
