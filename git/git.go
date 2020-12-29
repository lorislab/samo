package git

import (
	"os"
	"strings"

	"github.com/lorislab/samo/project"
	"github.com/lorislab/samo/tools"
	log "github.com/sirupsen/logrus"
)

// GitProject git project
type GitProject struct {
	name    string
	version string
}

// Type the type of the project
func (g GitProject) Type() project.Type {
	return project.Git
}

// Name project name
func (g GitProject) Name() string {
	return g.name
}

// Version project version
func (g GitProject) Version() string {
	return g.version
}

// Filename project file name
func (g GitProject) Filename() string {
	return ""
}

// SetVersion set project version
func (g GitProject) SetVersion(version string) {
	log.WithField("type", g.Type()).Fatal("This project does not support project file changes")
}

func Load(filename string) project.Project {

	if len(filename) == 0 {
		filename = ".git"
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			log.WithField("directory", filename).Debug("Missing default git directory .git")
			return nil
		}
	}

	name := tools.ExecCmdOutput("basename", "$(git remote get-url origin)")
	name = strings.TrimSuffix(name, ".git")

	version, _, _ := tools.GitCommit(6)
	result := GitProject{
		name:    name,
		version: version,
	}

	tmp := project.NextReleaseVersion(result, false)
	result.version = tmp.String()

	return &result
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
	return tools.ExecCmdOutput("git", "rev-parse", "--abbrev-ref", "HEAD")
}
