package cmd

import (
	"github.com/lorislab/samo/docker"
	"github.com/lorislab/samo/git"
	"github.com/lorislab/samo/helm"
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
	ReleaseSkipPush         bool         `mapstructure:"release-skip-push"`
	ReleaseTagMessage       string       `mapstructure:"release-tag-message"`
	DevMsg                  string       `mapstructure:"release-message"`
	ReleaseMajor            bool         `mapstructure:"release-major"`
	PatchBranchPrefix       string       `mapstructure:"patch-branch-prefix"`
	PatchSkipPush           bool         `mapstructure:"patch-skip-push"`
	PatchMsg                string       `mapstructure:"patch-message"`
	PatchTag                string       `mapstructure:"patch-tag"`
	DockerBuildTags         []string     `mapstructure:"docker-build-tags"`
	DockerRegistry          string       `mapstructure:"docker-registry"`
	DockerRepoPrefix        string       `mapstructure:"docker-repo-prefix"`
	DockerRepository        string       `mapstructure:"docker-repository"`
	Dockerfile              string       `mapstructure:"dockerfile"`
	DockerContext           string       `mapstructure:"docker-context"`
	DockerSkipPull          bool         `mapstructure:"docker-skip-pull"`
	DockerSkipPush          bool         `mapstructure:"docker-skip-push"`
	DockerPushTags          []string     `mapstructure:"docker-push-tags"`
	HelmInputDir            string       `mapstructure:"helm-input"`
	HelmOutputDir           string       `mapstructure:"helm-output"`
	HelmClean               bool         `mapstructure:"helm-clean"`
	HelmFilterTemplate      string       `mapstructure:"helm-filter-template"`
	HelmUpdateChart         []string     `mapstructure:"helm-update-chart"`
	HelmUpdateValues        []string     `mapstructure:"helm-update-values"`
	HelmBuildTags           []string     `mapstructure:"helm-build-versions"`
	HelmPushTags            []string     `mapstructure:"helm-push-verions"`
	HelmSkipPush            bool         `mapstructure:"helm-skip-push"`
	HelmRepository          string       `mapstructure:"helm-repo"`
	HelmRepositoryURL       string       `mapstructure:"helm-repo-url"`
	HelmRepositoryAdd       bool         `mapstructure:"helm-repo-add"`
	DockerReleaseRegistry   string       `mapstructure:"docker-release-registry"`
	DockerReleaseRepoPrefix string       `mapstructure:"docker-release-repo-prefix"`
	DockerReleaseRepository string       `mapstructure:"docker-release-repository"`
	DockerReleaseSkipPush   bool         `mapstructure:"docker-release-skip-push"`
}

var (
	projectCmd = &cobra.Command{
		Use:              "project",
		Short:            "Project operation",
		Long:             `Tasks for the project`,
		TraverseChildren: true,
	}
	createReleaseCmd = &cobra.Command{
		Use:   "create-release",
		Short: "Create release of the current project and state",
		Long:  `Create release of the current project and state`,
		Run: func(cmd *cobra.Command, args []string) {
			options, p := readProjectOptions()
			version := project.CreateRelease(p, options.DevMsg, options.ReleaseTagMessage, options.ReleaseMajor, options.ReleaseSkipPush)
			log.WithField("version", version).Info("New release created.")
		},
		TraverseChildren: true,
	}
	createPatchCmd = &cobra.Command{
		Use:   "create-patch",
		Short: "Create patch of the project release",
		Long:  `Create patch of the project release`,
		Run: func(cmd *cobra.Command, args []string) {
			options, p := readProjectOptions()
			branch := project.CreatePatch(p, options.PatchMsg, options.PatchTag, options.PatchBranchPrefix, options.PatchSkipPush)
			log.WithField("branch", branch).Info("New patch branch.")
		},
		TraverseChildren: true,
	}
	dockerBuildCmd = &cobra.Command{
		Use:   "docker-build",
		Short: "Build the docker image of the project",
		Long:  `Build the docker image of the project`,
		Run: func(cmd *cobra.Command, args []string) {
			options, p := readProjectOptions()

			request := docker.DockerRequest{
				Project:          p,
				Registry:         options.DockerRegistry,
				RepositoryPrefix: options.DockerRepoPrefix,
				Repository:       options.DockerRepository,
				Dockerfile:       options.Dockerfile,
				Context:          options.DockerContext,
				SkipPull:         options.DockerSkipPull,
				Versions:         project.CreateVersions(p, options.DockerBuildTags, options.HashLength, options.BuildNumberLength, options.BuildNumberPrefix),
			}
			image, _ := request.DockerBuild()
			log.WithFields(log.Fields{
				"image": image,
			}).Info("Docker build done!")
		},
		TraverseChildren: true,
	}
	dockerPushCmd = &cobra.Command{
		Use:   "docker-push",
		Short: "Push the docker image of the project",
		Long:  `Push the docker image of the project`,
		Run: func(cmd *cobra.Command, args []string) {
			op, p := readProjectOptions()
			request := docker.DockerRequest{
				Project:    p,
				Registry:   op.DockerRegistry,
				Dockerfile: op.Dockerfile,
				Context:    op.DockerContext,
				SkipPush:   op.DockerSkipPush,
				Versions:   project.CreateVersions(p, op.DockerPushTags, op.HashLength, op.BuildNumberLength, op.BuildNumberPrefix),
			}
			image, _ := request.DockerPush()
			log.WithFields(log.Fields{
				"image": image,
			}).Info("Push docker image done!")
		},
		TraverseChildren: true,
	}
	dockerReleaseCmd = &cobra.Command{
		Use:   "docker-release",
		Short: "Release the docker image and push to release registry",
		Long:  `Release the docker image and push to release registry`,
		Run: func(cmd *cobra.Command, args []string) {
			op, p := readProjectOptions()

			request := docker.DockerRequest{
				Project:                 p,
				Registry:                op.DockerRegistry,
				Dockerfile:              op.Dockerfile,
				Context:                 op.DockerContext,
				ReleaseRegistry:         op.DockerReleaseRegistry,
				ReleaseRepositoryPrefix: op.DockerReleaseRepoPrefix,
				ReleaseRepository:       op.DockerReleaseRepository,
				SkipPush:                op.DockerReleaseSkipPush,
				Versions:                project.CreateVersions(p, []string{project.VerBuild, project.VerRelease}, op.HashLength, op.BuildNumberLength, op.BuildNumberPrefix),
			}
			image := request.DockerRelease()
			log.WithFields(log.Fields{
				"image": image,
			}).Info("Release docker image done!")
		},
		TraverseChildren: true,
	}
	helmFilterCmd = &cobra.Command{
		Use:   "helm-filter",
		Short: "Filter project helm chart",
		Long:  `Filter project helm chart`,
		Run: func(cmd *cobra.Command, args []string) {
			options, p := readProjectOptions()
			helm := helm.HelmRequest{
				Project:  p,
				Input:    options.HelmInputDir,
				Output:   options.HelmOutputDir,
				Clean:    options.HelmClean,
				Template: options.HelmFilterTemplate,
			}
			helm.Filter()
		},
		TraverseChildren: true,
	}
	helmBuildCmd = &cobra.Command{
		Use:   "helm-build",
		Short: "Build helm chart",
		Long:  `Helm build helm chart`,
		Run: func(cmd *cobra.Command, args []string) {
			op, p := readProjectOptions()
			helm := helm.HelmRequest{
				Project:       p,
				Input:         op.HelmInputDir,
				Output:        op.HelmOutputDir,
				Clean:         op.HelmClean,
				SkipPush:      op.HelmSkipPush,
				Versions:      project.CreateVersions(p, op.HelmBuildTags, op.HashLength, op.BuildNumberLength, op.BuildNumberPrefix),
				PushVersions:  project.CreateVersions(p, op.HelmPushTags, op.HashLength, op.BuildNumberLength, op.BuildNumberPrefix),
				Template:      op.HelmFilterTemplate,
				Repository:    op.HelmRepository,
				RepositoryURL: op.HelmRepositoryURL,
				AddRepo:       op.HelmRepositoryAdd,
			}
			helm.Build()
		},
		TraverseChildren: true,
	}
	helmReleaseCmd = &cobra.Command{
		Use:   "helm-release",
		Short: "Release helm chart",
		Long:  `Download build version of the helm chart and create final version`,
		Run: func(cmd *cobra.Command, args []string) {
			options, p := readProjectOptions()
			helm := helm.HelmRequest{
				Project:           p,
				Input:             options.HelmInputDir,
				Output:            options.HelmOutputDir,
				Clean:             options.HelmClean,
				ChartUpdate:       options.HelmUpdateChart,
				ValuesUpdate:      options.HelmUpdateValues,
				SkipPush:          options.HelmSkipPush,
				HashLength:        options.HashLength,
				BuildNumberLength: options.BuildNumberLength,
				BuildNumberPrefix: options.BuildNumberPrefix,
			}
			helm.Release()
		},
		TraverseChildren: true,
	}
)

func init() {
	verList := project.VersionsText()

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

	addChildCmd(projectCmd, dockerBuildCmd)
	addSliceFlag(dockerBuildCmd, "docker-build-tags", "", []string{"build", "branch", "latest", "dev", "hash"}, "the list of docker image tags, custom or "+verList)
	addFlag(dockerBuildCmd, "dockerfile", "d", "src/main/docker/Dockerfile", "project dockerfile")
	addFlag(dockerBuildCmd, "docker-context", "", ".", "the docker build context")
	addBoolFlag(dockerBuildCmd, "docker-skip-pull", "", false, "skip docker pull for the build")
	fDockerRepository := addFlag(dockerBuildCmd, "docker-registry", "", "", "the docker registry")
	fDockerRepoPrefix := addFlag(dockerBuildCmd, "docker-repo-prefix", "", "", "the docker repository prefix")
	fDockerRepo := addFlag(dockerBuildCmd, "docker-repo", "i", "", "the docker repository. Default value project name.")

	addChildCmd(projectCmd, dockerPushCmd)
	addFlagRef(dockerPushCmd, fDockerRepository)
	addFlagRef(dockerPushCmd, fDockerRepoPrefix)
	addFlagRef(dockerPushCmd, fDockerRepo)
	addSliceFlag(dockerPushCmd, "docker-push-tags", "", []string{"build", "hash"}, "the list of docker image tags to be push, custom or "+verList)

	addChildCmd(projectCmd, dockerReleaseCmd)
	addFlagRef(dockerReleaseCmd, fDockerRepository)
	addFlagRef(dockerReleaseCmd, fDockerRepo)
	addFlagRef(dockerReleaseCmd, fDockerRepoPrefix)
	addFlag(dockerReleaseCmd, "docker-release-registry", "", "", "the docker release registry")
	addFlag(dockerReleaseCmd, "docker-release-repo-prefix", "", "", "the docker release repository prefix")
	addFlag(dockerReleaseCmd, "docker-release-repository", "", "", "the docker release repository. Default value project name.")
	addBoolFlag(dockerReleaseCmd, "docker-release-skip-push", "", false, "skip docker push of release image to registry")

	addChildCmd(projectCmd, helmFilterCmd)
	fHelmInput := addFlag(helmFilterCmd, "helm-input", "", "helm", "filter project helm chart input directory")
	fHelmOutput := addFlag(helmFilterCmd, "helm-output", "", "target/helm", "filter project helm chart output directory")
	fHelmClean := addBoolFlag(helmFilterCmd, "helm-clean", "", false, "clean output directory before filter")
	addFlagRequired(helmFilterCmd, "helm-filter-template", "", "maven", "Use the maven template for filter")

	addChildCmd(projectCmd, helmBuildCmd)
	addFlagRef(helmBuildCmd, fHelmInput)
	addFlagRef(helmBuildCmd, fHelmOutput)
	addFlagRef(helmBuildCmd, fHelmClean)
	addSliceFlag(helmBuildCmd, "helm-build-versions", "", []string{"build"}, "the list of build helm chart versions, custom or "+verList)
	addSliceFlag(helmBuildCmd, "helm-push-versions", "", []string{"build"}, "helm list of push helm chart versions, custom or "+verList)
	fHelmSkipPush := addBoolFlag(helmBuildCmd, "helm-skip-push", "", false, "skip helm push")
	addFlag(helmBuildCmd, "helm-repo", "", "", "helm repository name")
	addFlag(helmBuildCmd, "helm-repo-url", "", "", "helm repository name")

	addChildCmd(projectCmd, dockerReleaseCmd)
	addFlagRef(helmReleaseCmd, fHelmInput)
	addFlagRef(helmReleaseCmd, fHelmOutput)
	addFlagRef(helmReleaseCmd, fHelmClean)
	addFlagRef(helmReleaseCmd, fHelmSkipPush)
	addFlag(helmReleaseCmd, "helm-update-chart", "", "version={{ .Version }},appVersion={{ .Version }}", "list of key value to be replaced in the Chart.yaml")
	addFlag(helmReleaseCmd, "helm-update-values", "", "image.tag={{ .Version }}", "list of key value to be replaced in the values.yaml")

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
