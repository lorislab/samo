package tools

import (
	"os"
	"strings"
)

func GitBranch() string {
	tmp, exists := os.LookupEnv("GITHUB_REF")
	if exists && len(tmp) > 0 {
		return strings.TrimPrefix(tmp, "refs/heads/")
	}
	tmp, exists = os.LookupEnv("CI_COMMIT_REF_SLUG")
	if exists && len(tmp) > 0 {
		return tmp
	}
	tmp = ExecCmdOutput("git", "rev-parse", "--abbrev-ref", "HEAD")
	return strings.TrimPrefix(tmp, "heads/")
}

// Git execute git command
func Git(arg ...string) {
	err := execCmdErr("git", arg...)
	if err != nil {
		ExecCmd("rm", "-f", ".git/index.lock")
	}
}

func GitDescribe(firstVer string) (string, string, string) {
	output, err := CmdOutputErr("git", "describe", "--long", "--abbrev=100")
	if err == nil {
		items := strings.Split(output, "-")
		return items[0], items[1], strings.TrimPrefix(items[2], "g")
	}

	count := "0"
	hash := ""
	tmp, err := CmdOutputErr("git", "rev-list", "--max-count=1", "HEAD")
	if err == nil {
		hash = tmp
		c, err := CmdOutputErr("git", "rev-list", "--count", "HEAD")
		if err == nil {
			count = c
		}
	}
	return firstVer, count, hash
}
