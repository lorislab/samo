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

// IsFile is project base on the project file
func (g GitProject) IsFile() bool {
	return false
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
	log.WithField("type", g.Type()).Warn("This project does not support project file changes")
}

func Load(filename, firstVer string) project.Project {

	if len(filename) == 0 {
		filename = ".git"
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			log.WithField("directory", filename).Debug("Missing default git directory .git")
			return nil
		}
	}

	name := tools.ExecCmdOutput("git", "remote", "get-url", "origin")
	name = name[strings.LastIndex(name, "/")+1:]
	name = strings.TrimSuffix(name, ".git")

	version, _, _ := tools.GitCommit(6, firstVer)
	result := GitProject{
		name:    name,
		version: version,
	}

	versions := project.CreateVersions(result, nil, 0, 0, "", firstVer)
	tmp := versions.NextReleaseVersion(false)
	result.version = tmp.String()

	return &result
}
