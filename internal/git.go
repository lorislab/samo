package internal

import (
	"bytes"
	"os"
	"os/exec"
	"strings"
)

// GitBranch get the git branch
func GitBranch() string {
	if isGitHub() {
		tmp, exists := os.LookupEnv("GITHUB_REF")
		if exists && len(tmp) > 0 {
			return strings.TrimPrefix(tmp, "refs/heads/")
		}
	}
	if isGitLab() {
		tmp, exists := os.LookupEnv("CI_COMMIT_REF_NAME")
		if exists {
			return tmp
		}
	}
	out, err := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD").Output()
	if err != nil {
		panic(err)
	}
	return string(bytes.TrimRight(out, "\n"))
}

// GitHash get the git hash
func GitHash(length string) string {
	out, err := exec.Command("git", "rev-parse", "--short="+length, "HEAD").Output()
	if err != nil {
		panic(err)
	}
	return string(bytes.TrimRight(out, "\n"))
}

func isGitHub() bool {
	tmp, exists := os.LookupEnv("GITHUB_REF")
	return exists && len(tmp) > 0
}

func isGitLab() bool {
	tmp, exists := os.LookupEnv("GITLAB_CI")
	return exists && len(tmp) > 0
}
