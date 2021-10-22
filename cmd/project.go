package cmd

import (
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/lorislab/samo/log"
	"github.com/lorislab/samo/tools"
	"github.com/spf13/cobra"
	cc "gitlab.com/digitalxero/go-conventional-commit"
)

type projectFlags struct {
	FirstVersion       string `mapstructure:"first-version"`
	ReleaseMajor       bool   `mapstructure:"release-major"`
	ReleasePatch       bool   `mapstructure:"release-patch"`
	VersionTemplate    string `mapstructure:"version-template"`
	SkipPush           bool   `mapstructure:"skip-push"`
	ConvetionalCommits bool   `mapstructure:"conventional-commits"`
	BranchTemplate     string `mapstructure:"branch-template"`
	SkipLabels         bool   `mapstructure:"skip-samo-labels"`
	LabelTemplate      string `mapstructure:"labels-template-list"`
}

var sourceLinkRegex = `\/\/.*@`

var templateValues = `Name,Tag,Hash,Count,Branch,Version,Release,Major,Minor,Patch,Prerelease`

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
	addStringFlag(cmd, "version-template", "t", "{{ .Version }}-rc.{{ .Count }}", `the version go template string.
	values: `+templateValues+`
	functions:  trunc <lenght>
	For example: {{ .Tag }}-{{ trunc 10 .Hash }}`)
	addBoolFlag(cmd, "skip-push", "", false, "skip push changes")
	addBoolFlag(cmd, "conventional-commits", "c", false, "determine the project version based on the conventional commits")
	addStringFlag(cmd, "branch-template", "", "fix/{{ .Major }}.{{ .Minor }}.x", "patch-branch name template. Values: Major,Minor,Patch")

	addBoolFlag(cmd, "skip-samo-labels", "", false, "skip samo labels/annotations samo.project.revision,samo.project.version,samo.project.created")
	addStringFlag(cmd, "labels-template-list", "", "", `custom labels template list. 
	Values: `+templateValues+`
	Example: my-labe={{ .Branch }},my-const=123,my-count={{ .Count }}`)

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
	name       string
	describe   tools.GitDescribe
	rc         tools.GitDescribe
	branch     string
	source     string
	patchBuild bool
	version    *semver.Version
	rcVersion  *semver.Version
	release    *semver.Version
}

// Name project name
func (g Project) Name() string {
	return g.name
}

func (g Project) Source() string {
	return g.source
}

func (g Project) Major() int64 {
	return g.version.Major()
}

func (g Project) Minor() int64 {
	return g.version.Minor()
}

func (g Project) Patch() int64 {
	return g.version.Patch()
}

func (g Project) Prerelease() string {
	return g.version.Prerelease()
}

func (g Project) Version() string {
	return g.version.String()
}

func (g Project) lastRC() string {
	return g.rcVersion.String()
}

func (g Project) Release() string {
	return g.release.String()
}

func (g Project) Hash() string {
	return g.describe.Hash
}

func (g Project) Branch() string {
	return g.branch
}

func (g Project) Count() string {
	return g.describe.Count
}

func (g Project) Tag() string {
	return g.describe.Tag
}

func (g Project) IsPatchBuild() bool {
	return g.patchBuild
}

func loadProject(flags projectFlags) *Project {

	if _, err := os.Stat(".git"); os.IsNotExist(err) {
		log.Fatal("Missing git directory!", log.F("directory", ".git"))
	}

	// read repot git url or directory name
	tmp, err := tools.CmdOutputErr("git", "config", "remote.origin.url")
	if err != nil {
		tmp = tools.ExecCmdOutput("git", "rev-parse", "--show-toplevel")
	}

	// create project source
	source := tmp
	reg, _ := regexp.Compile(sourceLinkRegex)
	source = reg.ReplaceAllString(source, `//`)
	log.Debug("Project", log.F("source", source))

	// create project name
	name := "no-name"
	tmp = strings.TrimSuffix(tmp, ".git")
	tmp = filepath.Base(tmp)
	if len(tmp) > 0 && tmp != "." && tmp != "/" {
		name = tmp
	}

	describe := tools.GitDescribeInfo()
	rc := describe

	branch := tools.GitBranch()
	patchBuild := false

	version := flags.FirstVersion
	lastRC := version

	// check for empty repository
	if len(describe.Tag) > 0 {
		ver := tools.CreateSemVer(describe.Tag)
		patchBranch := createPatchBranchName(ver, flags)

		// branch name is patch branch or version is patch
		patchBuild = (branch == patchBranch) || ver.Patch() > 0

		log.Debug("Branch", log.Fields{"branch": branch, "patchBranch": patchBranch, "patchBuild": patchBuild, "count": describe.Count})

		// check last rc version
		if describe.Count == "0" {

			// find last tag before release
			rc = tools.GitDescribeExclude(describe.Tag)
			if len(rc.Tag) > 0 {
				rcver := tools.CreateSemVer(rc.Tag)
				if flags.ConvetionalCommits {
					lastRC = createNextVersionConvetionalCommits(rcver, patchBuild, rc)
				} else {
					lastRC = createNextVersion(rcver, false, false, patchBuild)
				}
			}
		}

		if flags.ConvetionalCommits {
			version = createNextVersionConvetionalCommits(ver, patchBuild, describe)
		} else {
			version = createNextVersion(ver, flags.ReleaseMajor, flags.ReleasePatch, patchBuild)
		}
	}

	p := &Project{
		name:       name,
		describe:   describe,
		branch:     branch,
		source:     source,
		patchBuild: patchBuild,
		rc:         rc,
		rcVersion:  createVersion(lastRC, branch, flags.VersionTemplate, rc),
		version:    createVersion(version, branch, flags.VersionTemplate, describe),
		release:    tools.CreateSemVer(version),
	}
	log.Debug("Versions", log.Fields{"version": p.Version(), "release": p.Release(), "rcVersion": p.rcVersion})
	return p
}

func createPatchBranchName(version *semver.Version, flags projectFlags) string {
	return tools.Template(version, flags.BranchTemplate)
}

func createVersion(version, branch, template string, describe tools.GitDescribe) *semver.Version {
	data := struct {
		Tag, Hash, Count, Branch, Version string
	}{
		Tag:     describe.Tag,
		Hash:    describe.Hash,
		Count:   describe.Count,
		Branch:  branch,
		Version: version,
	}

	tmp := tools.Template(data, template)
	return tools.CreateSemVer(tmp)
}

func createNextVersion(ver *semver.Version, major, patch, patchBranch bool) string {

	if patchBranch || patch || ver.Patch() != 0 {
		tmp := ver.IncPatch()
		return tmp.String()
	}
	if major {
		if ver.Patch() != 0 {
			log.Fatal("Can not created major release from the patch version!", log.F("version", ver.String()))
		}
		tmp := ver.IncMajor()
		return tmp.String()
	}
	tmp := ver.IncMinor()
	return tmp.String()
}

func createNextVersionConvetionalCommits(ver *semver.Version, patchBranch bool, describe tools.GitDescribe) string {

	// for patch branch we can ignore conventional commits
	if patchBranch {
		tmp := ver.IncPatch()
		return tmp.String()
	}

	if describe.Count == "0" {
		tmp := ver.IncMinor()
		return tmp.String()
	}

	commits := tools.GitLogMessages(ver.String(), "HEAD")
	commit := findConvCommit(commits)
	if commit.Major {
		tmp := ver.IncMajor()
		return tmp.String()
	}
	tmp := ver.IncMinor()
	return tmp.String()
}

func findConvCommit(commits []string) *cc.ConventionalCommit {
	var result *cc.ConventionalCommit
	for _, commit := range commits {
		item := cc.ParseConventionalCommit(strings.TrimPrefix(strings.TrimSuffix(commit, `"`), `"`))
		if item.Major {
			log.Debug("Major", log.F("commit", item))
			return item
		}
		if result == nil {
			result = item
		} else {
			if !result.Minor {
				result = item
			}
		}
	}
	return result
}
