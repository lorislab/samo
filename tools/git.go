package tools

import (
	"os"
	"strconv"
	"strings"

	log "github.com/sirupsen/logrus"
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
func Git(name string, arg ...string) {
	err := execCmdErr("git", arg...)
	if err != nil {
		ExecCmd("rm", "-f", ".git/index.lock")
	}
}

// GitCommit get the git commit
func GitCommit(length int) (string, string, string) {
	// search for latest annotated tag
	lastTag, err := cmdOutputErr("git", "describe", "--abbrev=0")
	log.WithField("tag", lastTag).Debug("Last tag")
	if err == nil {
		// get metadata from the git describe
		describe, err := cmdOutputErr("git", "describe", "--long", "--abbrev="+strconv.Itoa(length))
		if err == nil {
			describe = strings.TrimPrefix(describe, lastTag+"-")
			items := strings.Split(describe, "-")
			return lastTag, items[0], items[1]
		}
	}
	// not tag found in the git repository
	lastTag = "0.0.0"
	count := "0"
	// git commit hash
	hash, err := cmdOutputErr("git", "rev-parse", "--short="+strconv.Itoa(length), "HEAD")
	if err != nil {
		hash = Lpad("", "0", length)
	} else {
		// git commit count in the branch
		tmp, err := cmdOutputErr("git", "rev-list", "HEAD", "--count")
		if err == nil {
			count = tmp
		}
	}
	// git describe add 'g' prefix for the commit hash
	hash = "g" + hash
	return lastTag, count, hash
}
