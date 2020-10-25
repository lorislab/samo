package internal

import (
	"strconv"
	"strings"

	"github.com/Masterminds/semver/v3"
	log "github.com/sirupsen/logrus"
)

// GitProject git project
type GitProject struct {
	name           string
	version        string
	hashLength     int
	releaseVersion semver.Version
	count          string
	hash           string
}

// Name project name
func (r GitProject) Name() string {
	return r.name
}

// Version project version
func (r GitProject) Version() string {
	return r.version
}

// Filename project file name
func (r GitProject) Filename() string {
	return ""
}

// ReleaseSemVersion project release version
func (r GitProject) ReleaseSemVersion() *semver.Version {
	return &r.releaseVersion
}

// SetVersion set project version
func (r GitProject) SetVersion(version string) {
	// ignore
}

// NextReleaseSuffix next release suffix
func (r GitProject) NextReleaseSuffix() string {
	return ""
}

// Hash git hash
func (r GitProject) Hash() string {
	return r.hash
}

// Count git count
func (r GitProject) Count() string {
	return r.count
}

// LoadGitProject load git project
func LoadGitProject(hashLength int) *GitProject {
	name := ExecCmdOutput("basename", "$(git remote get-url origin)")
	name = strings.TrimSuffix(name, ".git")

	version, count, hash := GitCommit(hashLength)
	releaseVersion := NextReleaseVersion(CreateVersion(version), false)
	project := &GitProject{name: name, version: version, releaseVersion: releaseVersion, hashLength: hashLength, hash: hash, count: count}
	return project
}

// GitCommit get the git commit
func GitCommit(length int) (string, string, string) {
	// search for latest annotated tag
	lastTag, err := execCmdOutputErr("git", "describe", "--abbrev=0")
	log.Debugf("Last tag %s", lastTag)
	if err == nil {
		// get metadata from the git describe
		describe, err := execCmdOutputErr("git", "describe", "--long", "--abbrev="+strconv.Itoa(length))
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
	hash, err := execCmdOutputErr("git", "rev-parse", "--short="+strconv.Itoa(length), "HEAD")
	if err != nil {
		hash = lpad("", "0", length)
	} else {
		// git commit count in the branch
		tmp, err := execCmdOutputErr("git", "rev-list", "HEAD", "--count")
		if err == nil {
			count = tmp
		}
	}
	// git describe add 'g' prefix for the commit hash
	hash = "g" + hash
	return lastTag, count, hash
}
