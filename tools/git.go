package tools

import (
	"os"
	"strings"

	log "github.com/sirupsen/logrus"
)

func GitRemoveAllTagsForCurrentCommit() {
	// find all tags for the current commit
	list := ExecCmdOutput("git", "--no-pager", "tag", "--points-at", "HEAD")
	if len(list) <= 0 {
		log.Debug("No tag found on current commit")
		return
	}
	// could be multiple tags
	tags := strings.Split(list, "\n")
	log.WithField("tags", tags).Info("Remove git tags for current commit")

	// delete the local tags
	var cmd []string
	cmd = append(cmd, "tag", "-d")
	cmd = append(cmd, tags...)
	Git(cmd...)
}

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

func GitDescribe() (string, string, string) {
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
	return "", count, hash
}

func GitDescribeExclude(tag string) (string, string, string) {
	output, err := CmdOutputErr("git", "describe", "--long", "--abbrev=100", "--exclude", tag)
	if err != nil {
		log.WithField("tag", tag).Info("Error execute git discribe with exclude tag")
	}
	items := strings.Split(output, "-")
	return items[0], items[1], strings.TrimPrefix(items[2], "g")
}
