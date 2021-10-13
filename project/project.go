package project

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/lorislab/samo/tools"
	log "github.com/sirupsen/logrus"
)

// Project common project interface
type Project struct {
	name    string
	tag     string
	count   string
	hash    string
	branch  string
	version *semver.Version
	release *semver.Version
}

// Name project name
func (g Project) Name() string {
	return g.name
}

func (g Project) Version() string {
	return g.version.String()
}

func (g Project) ReleaseVersion() string {
	return g.release.String()
}

func (g Project) Hash() string {
	return g.hash
}

func (g Project) Branch() string {
	return g.branch
}

func (g Project) Count() string {
	return g.count
}

func (g Project) Tag() string {
	return g.tag
}

func LoadProject(firstVer, template string, major, patch bool) *Project {

	if _, err := os.Stat(".git"); os.IsNotExist(err) {
		log.WithField("directory", ".git").Fatal("Missing git directory!")
	}

	name := "no-name"
	tmp, err := tools.CmdOutputErr("git", "config", "remote.origin.url")
	if err != nil {
		tmp = tools.ExecCmdOutput("git", "rev-parse", "--show-toplevel")
	}
	tmp = strings.TrimSuffix(tmp, ".git")
	tmp = filepath.Base(tmp)
	if len(tmp) > 0 && tmp != "." && tmp != "/" {
		name = tmp
	}

	tag, count, hash := tools.GitDescribe(firstVer)

	branch := tools.GitBranch()

	nextVersion := createNextVersion(tag, major, patch)

	result := &Project{
		name:    name,
		tag:     tag,
		count:   count,
		hash:    hash,
		branch:  branch,
		version: createVersion(tag, nextVersion, count, hash, branch, template),
		release: tools.CreateSemVer(nextVersion),
	}
	return result
}

func createVersion(tag, nextVersion, count, hash, branch, template string) *semver.Version {

	if count == "0" {
		return tools.CreateSemVer(tag)
	}

	data := struct {
		Tag    string
		Hash   string
		Count  string
		Branch string
	}{
		Tag:    nextVersion,
		Hash:   hash,
		Count:  count,
		Branch: branch,
	}

	tmp := tools.Template(data, template)
	return tools.CreateSemVer(tmp)
}

func createNextVersion(tag string, major, patch bool) string {
	ver := tools.CreateSemVer(tag)
	if patch || ver.Patch() != 0 {
		tmp := ver.IncPatch()
		return tmp.String()
	}
	if major {
		if ver.Patch() != 0 {
			log.WithField("version", ver.String()).Fatal("Can not created major release from the patch version!")
		}
		tmp := ver.IncMajor()
		return tmp.String()
	}
	tmp := ver.IncMinor()
	return tmp.String()
}
