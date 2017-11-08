package vcs

import (
	"os/exec"
	"regexp"
)

type git struct {
	repo string
}

// LoadBranches will load the branches from a (remote) git repository
func (f git) LoadBranches() ([]string, error) {
	var branchNames []string

	cmd := exec.Command("git", "ls-remote", "--refs", f.repo)
	output, err := cmd.Output()
	if err != nil {
		return branchNames, err
	}

	re := regexp.MustCompile(`refs\/remotes\/origin\/(.*)\n`)
	for _, match := range re.FindAllStringSubmatch(string(output), -1) {
		if match[1] == "HEAD" {
			continue
		}
		branchNames = append(branchNames, match[1])
	}

	return branchNames, err
}
