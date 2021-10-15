package cmd

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/lorislab/samo/tools"
	log "github.com/sirupsen/logrus"
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
	SkipLabels         bool   `mapstructure:"skip-default-labels"`
	LabelTemplate      string `mapstructure:"labels-template-list"`
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
	addBoolFlag(cmd, "skip-push", "", false, "skip push changes")
	addBoolFlag(cmd, "conventional-commits", "c", false, "determine the project version based on the conventional commits")
	addStringFlag(cmd, "branch-template", "", "fix/{{ .Major }}.{{ .Minor }}.x", "patch-branch name template. Values: Major,Minor,Patch")

	addBoolFlag(cmd, "skip-default-labels", "", false, "skip default labels/annotations samo.git.hash,samo.project.version,samo.project.name,samo.project.release")
	addStringFlag(cmd, "labels-template-list", "", "", `custom labels template list. 
	Values: Hash,Branch,Tag,Count,Version,Release.
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
	name        string
	tag         string
	count       string
	hash        string
	branch      string
	patchBranch bool
	version     *semver.Version
	release     *semver.Version
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

func (g Project) IsPatchBranch() bool {
	return g.patchBranch
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

	describe := tools.GitDescribeInfo()

	branch := tools.GitBranch()
	isPatchBranch := false

	version := flags.FirstVersion

	// check for empty repository
	if len(describe.Tag) > 0 {

		ver := tools.CreateSemVer(describe.Tag)
		patchBranch := createPatchBranchName(ver, flags)
		isPatchBranch = branch == patchBranch
		log.WithFields(log.Fields{"branch": branch, "patchBranch": patchBranch, "isPatchBranch": isPatchBranch}).Debug("Branch")

		// commit + tag
		if describe.Count == "0" {
			// next version is current tag
			version = describe.Tag
			// exdcute git describe without current tag to get old tag + count + hash
			describe = tools.GitDescribeExclude(describe.Tag)
		} else {
			if flags.ConvetionalCommits {
				version = createNextVersionConvetionalCommits(ver, isPatchBranch)
			} else {
				version = createNextVersion(ver, flags.ReleaseMajor, flags.ReleasePatch, isPatchBranch)
			}
		}
	}

	return &Project{
		name:        name,
		tag:         describe.Tag,
		count:       describe.Count,
		hash:        describe.Hash,
		branch:      branch,
		patchBranch: isPatchBranch,
		version:     createVersion(version, branch, flags.VersionTemplate, describe),
		release:     tools.CreateSemVer(version),
	}
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
			log.WithField("version", ver.String()).Fatal("Can not created major release from the patch version!")
		}
		tmp := ver.IncMajor()
		return tmp.String()
	}
	tmp := ver.IncMinor()
	return tmp.String()
}

func createNextVersionConvetionalCommits(ver *semver.Version, patchBranch bool) string {

	// for patch branch we can ignore conventional commits
	if patchBranch {
		tmp := ver.IncPatch()
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
		log.WithField("commit", item).Debug("Commit")
		if item.Major {
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
