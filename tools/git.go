package tools

import (
	"os"
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
func Git(arg ...string) {
	err := execCmdErr("git", arg...)
	if err != nil {
		ExecCmd("rm", "-f", ".git/index.lock")
	}
}

type GitDescribe struct {
	Tag, Count, Hash string
}

func GitDescribeInfo() GitDescribe {
	output, err := CmdOutputErr("git", "describe", "--long", "--abbrev=100")
	if err == nil {
		items := strings.Split(output, "-")
		return GitDescribe{
			Tag:   items[0],
			Count: items[1],
			Hash:  strings.TrimPrefix(items[2], "g"),
		}
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
	return GitDescribe{
		Tag:   "",
		Count: count,
		Hash:  hash,
	}
}

func GitDescribeExclude(tag string) GitDescribe {
	output, err := CmdOutputErr("git", "describe", "--long", "--abbrev=100", "--exclude", tag)
	if err != nil {
		log.WithField("tag", tag).Fatal("Error execute git discribe with exclude tag")
	}
	items := strings.Split(output, "-")
	return GitDescribe{
		Tag:   items[0],
		Count: items[1],
		Hash:  strings.TrimPrefix(items[2], "g"),
	}
}

func GitLogMessages(from, to string) []string {
	output, err := CmdOutputErr("git", "--no-pager", "log", `--pretty=format:"%s"`, from+"..."+to)
	if err != nil {
		log.WithFields(log.Fields{"from": from, "to": to}).Fatal("Error execute git log messages")
	}
	if len(output) < 1 {
		return []string{}
	}
	return strings.Split(output, "\n")
}
