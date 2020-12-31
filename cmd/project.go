package cmd

import (
	"github.com/lorislab/samo/git"
	"github.com/lorislab/samo/maven"
	"github.com/lorislab/samo/npm"
	"github.com/lorislab/samo/project"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type projectFlags struct {
	File                    string       `mapstructure:"file"`
	Type                    project.Type `mapstructure:"type"`
	Versions                []string     `mapstructure:"version"`
	OutputValue             bool         `mapstructure:"value-only"`
	HashLength              int          `mapstructure:"build-hash"`
	BuildNumberPrefix       string       `mapstructure:"build-prefix"`
	BuildNumberLength       int          `mapstructure:"build-length"`
	ReleaseSkipPush         bool         `mapstructure:"release-push-skip"`
	ReleaseTagMessage       string       `mapstructure:"release-tag-message"`
	DevMsg                  string       `mapstructure:"release-message"`
	ReleaseMajor            bool         `mapstructure:"release-major"`
	PatchBranchPrefix       string       `mapstructure:"patch-branch-prefix"`
	PatchSkipPush           bool         `mapstructure:"patch-push-skip"`
	PatchMsg                string       `mapstructure:"patch-message"`
	PatchTag                string       `mapstructure:"patch-tag"`
	DockerRegistry          string       `mapstructure:"docker-registry"`
	DockerRepoPrefix        string       `mapstructure:"docker-repo-prefix"`
	DockerRepository        string       `mapstructure:"docker-repository"`
	Dockerfile              string       `mapstructure:"dockerfile"`
	DockerContext           string       `mapstructure:"docker-context"`
	DockerSkipPull          bool         `mapstructure:"docker-pull-skip"`
	DockerSkipPush          bool         `mapstructure:"docker-push-skip"`
	HelmInputDir            string       `mapstructure:"helm-input"`
	HelmOutputDir           string       `mapstructure:"helm-output"`
	HelmClean               bool         `mapstructure:"helm-clean"`
	HelmFilterTemplate      string       `mapstructure:"helm-filter-template"`
	HelmBuildFilter         bool         `mapstructure:"helm-filter"`
	HelmUpdateChart         []string     `mapstructure:"helm-update-chart"`
	HelmUpdateValues        []string     `mapstructure:"helm-update-values"`
	HelmSkipPush            bool         `mapstructure:"helm-push-skip"`
	HelmRepository          string       `mapstructure:"helm-repo"`
	HelmRepoUsername        string       `mapstructure:"helm-repo-username"`
	HelmRepoPassword        string       `mapstructure:"helm-repo-password"`
	HelmRepositoryURL       string       `mapstructure:"helm-repo-url"`
	HelmRepositoryAdd       bool         `mapstructure:"helm-repo-add"`
	DockerReleaseRegistry   string       `mapstructure:"docker-release-registry"`
	DockerReleaseRepoPrefix string       `mapstructure:"docker-release-repo-prefix"`
	DockerReleaseRepository string       `mapstructure:"docker-release-repository"`
	DockerReleaseSkipPush   bool         `mapstructure:"docker-release-push-skip"`
}

var (
	verList = project.VersionsText()

	projectCmd = &cobra.Command{
		Use:              "project",
		Short:            "Project operation",
		Long:             `Tasks for the project`,
		TraverseChildren: true,
	}
	createReleaseCmd = &cobra.Command{
		Use:   "release",
		Short: "Create release of the current project and state",
		Long:  `Create release of the current project and state`,
		Run: func(cmd *cobra.Command, args []string) {
			op, p := readProjectOptions()
			project.CreateRelease(p, op.DevMsg, op.ReleaseTagMessage, op.ReleaseMajor, op.ReleaseSkipPush)
		},
		TraverseChildren: true,
	}
	createPatchCmd = &cobra.Command{
		Use:   "patch",
		Short: "Create patch of the project release",
		Long:  `Create patch of the project release`,
		Run: func(cmd *cobra.Command, args []string) {
			op, p := readProjectOptions()
			project.CreatePatch(p, op.PatchMsg, op.PatchTag, op.PatchBranchPrefix, op.PatchSkipPush)
		},
		TraverseChildren: true,
	}
)

func init() {
	addChildCmd(rootCmd, projectCmd)
	addSliceFlag(projectCmd, "version", "", []string{project.VerVersion}, "project version type, custom or "+verList)
	addFlag(projectCmd, "build-prefix", "b", "rc", "the build number prefix")
	addIntFlag(projectCmd, "build-length", "e", 3, "the build number length")
	addIntFlag(projectCmd, "build-hash", "", 12, "the git hash length")

	addChildCmd(projectCmd, createReleaseCmd)
	addFlag(createReleaseCmd, "release-message", "", "Create new development version", "commit message for new development version")
	addBoolFlag(createReleaseCmd, "release-major", "", false, "create a major release")
	addFlag(createReleaseCmd, "release-tag-message", "", "", "the release tag message")
	addBoolFlag(createReleaseCmd, "release-skip-push", "", false, "skip git push release")

	addChildCmd(projectCmd, createPatchCmd)
	addFlagRequired(createPatchCmd, "patch-tag", "", "", "the tag version of the patch branch")
	addFlag(createPatchCmd, "patch-message", "", "Create new patch version", "commit message for new patch version")
	addFlag(createPatchCmd, "patch-branch-prefix", "", "", "patch branch prefix")
	addBoolFlag(createPatchCmd, "patch-skip-push", "", false, "skip git push patch branch.")
}

func readProjectOptions() (projectFlags, project.Project) {
	options := projectFlags{}
	err := viper.Unmarshal(&options)
	if err != nil {
		panic(err)
	}
	log.WithField("options", options).Debug("Load project options")
	return options, loadProject(options.File, project.Type(options.Type))
}

func loadProject(file string, projectType project.Type) project.Project {

	// find the project type
	if len(projectType) > 0 {
		switch projectType {
		case project.Maven:
			return maven.Load(file)
		case project.Npm:
			return npm.Load(file)
		case project.Git:
			return git.Load(file)
		}
	}

	// priority 1 maven
	project := maven.Load("")
	if project != nil {
		return project
	}
	// priority 2 npm
	project = npm.Load("")
	if project != nil {
		return project
	}
	// priority 3 git
	project = git.Load("")
	if project != nil {
		return project
	}

	// failed loading the poject
	log.WithFields(log.Fields{
		"type": projectType,
		"file": file,
	}).Fatal("Could to find project file. Please specified the type --type.")
	return nil
}
