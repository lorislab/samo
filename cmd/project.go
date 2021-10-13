package cmd

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/lorislab/samo/tools"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type projectFlags struct {
	FirstVersion       string `mapstructure:"first-version"`
	ReleaseMajor       bool   `mapstructure:"release-major"`
	ReleasePatch       bool   `mapstructure:"release-patch"`
	VersionTemplate    string `mapstructure:"version-template"`
	SkipPush           bool   `mapstructure:"skip-push"`
	ConvetionalCommits bool   `mapstructure:"conventional-commits"`
}

var versionTemplateInfo = `the version go temmplate string.
values: Tag,Hash,Count,Branch,Version
functions:  trunc <lenght>
For example: {{ .Tag }}-{{ trunc 10 .Hash }}
`

func createProjectCmd() *cobra.Command {

	cmd := &cobra.Command{
		Use:              "project",
		Short:            "Project operation",
		Long:             `Tasks for the project. To build, push or release docker and helm artifacts of the project.`,
		TraverseChildren: true,
	}

	addStringFlag(cmd, "first-version", "", "0.0.0", "the first version of the project")
	addBoolFlag(cmd, "release-major", "", false, "create a major release")
	addBoolFlag(cmd, "release-patch", "", false, "create a patch release")
	addStringFlag(cmd, "version-template", "t", "{{ .Version }}-rc.{{ .Count }}", versionTemplateInfo)
	addBoolFlag(cmd, "skip-push", "", false, "skip git push changes")
	addBoolFlag(cmd, "conventional-commits", "c", false, "determine the project version based on the conventional commits")

	addChildCmd(cmd, createProjectVersionCmd())
	addChildCmd(cmd, createProjectNameCmd())
	addChildCmd(cmd, createProjectReleaseCmd())
	addChildCmd(cmd, createProjectPatchCmd())
	addChildCmd(cmd, createDockerCmd())
	addChildCmd(cmd, createHelmCmd())

	return cmd
}

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

func loadProject(flags projectFlags) *Project {

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
	branch := tools.GitBranch()

	tag, count, hash := tools.GitDescribe()

	nextVersion := flags.FirstVersion

	// check for empty repository
	if len(tag) > 0 {
		// commit + tag
		if count == "0" {
			// next version is current tag
			nextVersion = tag
			// exdcute git describe withou current tag to get old tag + count + hash
			t, c, h := tools.GitDescribeExclude(tag)
			tag = t
			count = c
			hash = h
		} else {
			if flags.ConvetionalCommits {
				// TODO: patch branch
				nextVersion = createNextVersionConvetionalCommits(tag)
			} else {
				nextVersion = createNextVersion(tag, flags.ReleaseMajor, flags.ReleasePatch)
			}
		}
	}

	result := &Project{
		name:    name,
		tag:     tag,
		count:   count,
		hash:    hash,
		branch:  branch,
		version: createVersion(tag, nextVersion, count, hash, branch, flags.VersionTemplate),
		release: tools.CreateSemVer(nextVersion),
	}
	return result
}

func createVersion(tag, nextVersion, count, hash, branch, template string) *semver.Version {
	data := struct {
		Tag, Hash, Count, Branch, Version string
	}{
		Tag:     tag,
		Hash:    hash,
		Count:   count,
		Branch:  branch,
		Version: nextVersion,
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

func createNextVersionConvetionalCommits(tag string) string {

	return tag
}
