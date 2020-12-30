package cmd

import (
	"fmt"

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
	ReleaseSkipPush         bool         `mapstructure:"release-skip-push"`
	ReleaseTagMessage       string       `mapstructure:"release-tag-message"`
	DevMsg                  string       `mapstructure:"release-message"`
	ReleaseMajor            bool         `mapstructure:"release-major"`
	HashLength              int          `mapstructure:"hash-length"`
	BuildNumberPrefix       string       `mapstructure:"build-number-prefix"`
	BuildNumberLength       int          `mapstructure:"build-number-length"`
	PatchBranchPrefix       string       `mapstructure:"patch-branch-prefix"`
	PatchSkipPush           bool         `mapstructure:"patch-skip-push"`
	PatchMsg                string       `mapstructure:"patch-message"`
	PatchTag                string       `mapstructure:"patch-tag"`
	DockerBuildTags         []string     `mapstructure:"docker-build-tags"`
	DockerBuildCustomTags   []string     `mapstructure:"docker-build-custom-tags"`
	DockerDevTags           []string     `mapstructure:"docker-dev-tags"`
	DockerDevCustomTags     []string     `mapstructure:"docker-dev-custom-tags"`
	DockerRegistry          string       `mapstructure:"docker-registry"`
	DockerRepoPrefix        string       `mapstructure:"docker-repo-prefix"`
	DockerRepository        string       `mapstructure:"docker-repository"`
	Dockerfile              string       `mapstructure:"dockerfile"`
	DockerContext           string       `mapstructure:"docker-context"`
	DockerSkipPull          bool         `mapstructure:"docker-skip-pull"`
	DockerSkipPush          bool         `mapstructure:"docker-skip-push"`
	DockerPushTags          []string     `mapstructure:"docker-push-tags"`
	DockerPushCustomTags    []string     `mapstructure:"docker-push-custom-tags"`
	HelmInputDir            string       `mapstructure:"helm-input"`
	HelmOutputDir           string       `mapstructure:"helm-output"`
	HelmClean               bool         `mapstructure:"helm-clean"`
	HelmFilterTemplate      string       `mapstructure:"helm-filter-template"`
	HelmUpdateChart         []string     `mapstructure:"helm-update-chart"`
	HelmUpdateValues        []string     `mapstructure:"helm-update-values"`
	HelmBuildTags           []string     `mapstructure:"helm-build-tags"`
	HelmPushTags            []string     `mapstructure:"helm-push-tags"`
	HelmSkipPush            bool         `mapstructure:"helm-skip-push"`
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
	projectVersionCmd = &cobra.Command{
		Use:   "version",
		Short: "Show the project version",
		Long:  `Tasks to show the project version`,
		Run: func(cmd *cobra.Command, args []string) {
			_, p := readProjectOptions()
			version := project.CreateVersion(p)
			fmt.Printf("%s\n", version.String())
		},
		TraverseChildren: true,
	}
	nameCmd = &cobra.Command{
		Use:   "name",
		Short: "Show the project name",
		Long:  `Tasks to show the maven project name`,
		Run: func(cmd *cobra.Command, args []string) {
			_, p := readProjectOptions()
			fmt.Printf("%s\n", p.Name())
		},
		TraverseChildren: true,
	}
	buildVersionCmd = &cobra.Command{
		Use:   "build-version",
		Short: "Show the project version to build version",
		Long:  `Show the project version to build version`,
		Run: func(cmd *cobra.Command, args []string) {
			options, p := readProjectOptions()
			version := project.BuildVersion(p, options.HashLength, options.BuildNumberLength, options.BuildNumberPrefix)
			fmt.Printf("%s\n", version.String())
		},
		TraverseChildren: true,
	}
	releaseVersionCmd = &cobra.Command{
		Use:   "release-version",
		Short: "Show the project next release version",
		Long:  `Show the project next release version`,
		Run: func(cmd *cobra.Command, args []string) {
			_, p := readProjectOptions()
			version := project.ReleaseVersion(p)
			fmt.Printf("%s\n", version.String())
		},
		TraverseChildren: true,
	}
	setBuildVersionCmd = &cobra.Command{
		Use:   "set-build-version",
		Short: "Set the build version to the project",
		Long:  `Change the version of the project to the build version`,
		Run: func(cmd *cobra.Command, args []string) {
			options, p := readProjectOptions()
			buildVersion := project.BuildVersion(p, options.HashLength, options.BuildNumberLength, options.BuildNumberPrefix)

			version := p.Version()
			p.SetVersion(buildVersion.String())

			log.WithFields(log.Fields{
				"file":          p.Filename(),
				"version":       version,
				"build-version": buildVersion,
			}).Info("Change the version of the project to the build version")
		},
		TraverseChildren: true,
	}
	setReleaseVersionCmd = &cobra.Command{
		Use:   "set-release-version",
		Short: "Set the release version to the project",
		Long:  `Change the version of the project to the release version`,
		Run: func(cmd *cobra.Command, args []string) {

			_, p := readProjectOptions()
			releaseVersion := project.ReleaseVersion(p)
			version := p.Version()
			p.SetVersion(releaseVersion.String())

			log.WithFields(log.Fields{
				"file":            p.Filename(),
				"version":         version,
				"release-version": releaseVersion,
			}).Info("Change the version of the project to the release version")
		},
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
				Project:           p,
				Registry:          options.DockerRegistry,
				RepositoryPrefix:  options.DockerRepoPrefix,
				Repository:        options.DockerRepository,
				HashLength:        options.HashLength,
				Tags:              project.CreateSet(options.DockerBuildTags),
				CustomTags:        options.DockerBuildCustomTags,
				Dockerfile:        options.Dockerfile,
				Context:           options.DockerContext,
				SkipPull:          options.DockerSkipPull,
				BuildNumberLength: options.BuildNumberLength,
				BuildNumberPrefix: options.BuildNumberPrefix,
			}
			image, _ := request.DockerBuild()
			log.WithFields(log.Fields{
				"image": image,
			}).Info("Docker build done!")
		},
		TraverseChildren: true,
	}
	dockerBuildDevCmd = &cobra.Command{
		Use:   "docker-build-dev",
		Short: "Build the docker image of the project for local development",
		Long:  `Build the docker image of the project for local development`,
		Run: func(cmd *cobra.Command, args []string) {
			options, p := readProjectOptions()
			request := docker.DockerRequest{
				Project:    p,
				Registry:   options.DockerRegistry,
				Dockerfile: options.Dockerfile,
				Context:    options.DockerContext,
				SkipPull:   options.DockerSkipPull,
				Tags:       project.CreateSet(options.DockerDevTags),
				CustomTags: options.DockerDevCustomTags,
			}
			image, _ := request.DockerBuildDev()
			log.WithFields(log.Fields{
				"image": image,
			}).Info("Docker dev build done!")
		},
		TraverseChildren: true,
	}
	dockerPushCmd = &cobra.Command{
		Use:   "docker-push",
		Short: "Push the docker image of the project",
		Long:  `Push the docker image of the project`,
		Run: func(cmd *cobra.Command, args []string) {
			options, p := readProjectOptions()
			request := docker.DockerRequest{
				Project:    p,
				Registry:   options.DockerRegistry,
				Dockerfile: options.Dockerfile,
				Context:    options.DockerContext,
				SkipPush:   options.DockerSkipPush,
				HashLength: options.HashLength,
				Tags:       project.CreateSet(options.DockerPushTags),
				CustomTags: options.DockerPushCustomTags,
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
			options, p := readProjectOptions()

			request := docker.DockerRequest{
				Project:                 p,
				Registry:                options.DockerRegistry,
				Dockerfile:              options.Dockerfile,
				Context:                 options.DockerContext,
				HashLength:              options.HashLength,
				BuildNumberLength:       options.BuildNumberLength,
				BuildNumberPrefix:       options.BuildNumberPrefix,
				ReleaseRegistry:         options.DockerReleaseRegistry,
				ReleaseRepositoryPrefix: options.DockerReleaseRepoPrefix,
				ReleaseRepository:       options.DockerReleaseRepository,
				SkipPush:                options.DockerReleaseSkipPush,
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
			options, p := readProjectOptions()
			helm := helm.HelmRequest{
				Project:           p,
				Input:             options.HelmInputDir,
				Output:            options.HelmOutputDir,
				Clean:             options.HelmClean,
				SkipPush:          options.HelmSkipPush,
				BuildTags:         project.CreateSet(options.HelmBuildTags),
				PushTags:          project.CreateSet(options.HelmPushTags),
				Template:          options.HelmFilterTemplate,
				HashLength:        options.HashLength,
				BuildNumberLength: options.BuildNumberLength,
				BuildNumberPrefix: options.BuildNumberPrefix,
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
	rootCmd.AddCommand(projectCmd)

	projectCmd.AddCommand(projectVersionCmd)
	fFile := addFlag(projectVersionCmd, "file", "f", "", "project file pom.xml, project.json or .git")
	fType := addFlag(projectVersionCmd, "type", "t", "", "project type maven, npm or git")

	projectCmd.AddCommand(nameCmd)
	addFlagRef(nameCmd, fFile)
	addFlagRef(nameCmd, fType)

	projectCmd.AddCommand(buildVersionCmd)
	addFlagRef(buildVersionCmd, fFile)
	addFlagRef(buildVersionCmd, fType)
	fBuildNumberPrefix := addFlag(buildVersionCmd, "build-number-prefix", "b", "rc", "the build number prefix")
	fBuildNumberLength := addIntFlag(buildVersionCmd, "build-number-length", "e", 3, "the build number length")
	fHashLength := addIntFlag(buildVersionCmd, "hash-length", "", 12, "the git hash length")

	projectCmd.AddCommand(releaseVersionCmd)
	addFlagRef(releaseVersionCmd, fFile)
	addFlagRef(releaseVersionCmd, fType)

	projectCmd.AddCommand(setBuildVersionCmd)
	addFlagRef(setBuildVersionCmd, fFile)
	addFlagRef(setBuildVersionCmd, fType)
	addFlagRef(setBuildVersionCmd, fBuildNumberPrefix)
	addFlagRef(setBuildVersionCmd, fBuildNumberLength)
	addFlagRef(setBuildVersionCmd, fHashLength)

	projectCmd.AddCommand(setReleaseVersionCmd)
	addFlagRef(setReleaseVersionCmd, fFile)
	addFlagRef(setReleaseVersionCmd, fType)

	projectCmd.AddCommand(createReleaseCmd)
	addFlagRef(createReleaseCmd, fFile)
	addFlagRef(createReleaseCmd, fType)
	addFlag(createReleaseCmd, "release-message", "", "Create new development version", "commit message for new development version")
	addBoolFlag(createReleaseCmd, "release-major", "", false, "create a major release")
	addFlag(createReleaseCmd, "release-tag-message", "", "", "the release tag message")
	addBoolFlag(createReleaseCmd, "release-skip-push", "", false, "skip git push release")

	projectCmd.AddCommand(createPatchCmd)
	addFlagRef(createPatchCmd, fFile)
	addFlagRef(createPatchCmd, fType)
	addFlagRequired(createPatchCmd, "patch-tag", "", "", "the tag version of the patch branch")
	addFlag(createPatchCmd, "patch-message", "", "Create new patch version", "commit message for new patch version")
	addFlag(createPatchCmd, "patch-branch-prefix", "", "", "patch branch prefix")
	addBoolFlag(createPatchCmd, "patch-skip-push", "", false, "skip git push patch branch.")

	projectCmd.AddCommand(dockerBuildCmd)
	addFlagRef(dockerBuildCmd, fFile)
	addFlagRef(dockerBuildCmd, fType)
	addFlagRef(dockerBuildCmd, fHashLength)
	addFlagRef(dockerBuildCmd, fBuildNumberPrefix)
	addFlagRef(dockerBuildCmd, fBuildNumberLength)
	addSliceFlag(dockerBuildCmd, "docker-build-tags", "", []string{"build-version", "branch", "latest", "dev", "hash"}, "the list of docker image tags, values:{version,build-version,branch,latest,dev,hash}")
	addSliceFlag(dockerBuildCmd, "docker-build-custom-tags", "", []string{}, "add your custom image tags")
	fDockerFile := addFlag(dockerBuildCmd, "dockerfile", "d", "src/main/docker/Dockerfile", "project dockerfile")
	fDockerRepository := addFlag(dockerBuildCmd, "docker-registry", "", "", "the docker registry")
	fDockerRepoPrefix := addFlag(dockerBuildCmd, "docker-repo-prefix", "", "", "the docker repository prefix")
	fDockerRepo := addFlag(dockerBuildCmd, "docker-repo", "i", "", "the docker repository. Default value project name.")
	fDockerContext := addFlag(dockerBuildCmd, "docker-context", "", ".", "the docker build context")
	fockerSkipPull := addBoolFlag(dockerBuildCmd, "docker-skip-pull", "", false, "skip docker pull for the build")

	projectCmd.AddCommand(dockerBuildDevCmd)
	addFlagRef(dockerBuildDevCmd, fFile)
	addFlagRef(dockerBuildDevCmd, fType)
	addFlagRef(dockerBuildDevCmd, fDockerRepo)
	addFlagRef(dockerBuildDevCmd, fDockerFile)
	addFlagRef(dockerBuildDevCmd, fDockerContext)
	addFlagRef(dockerBuildDevCmd, fockerSkipPull)
	addSliceFlag(dockerBuildDevCmd, "docker-dev-tags", "", []string{"latest"}, "the list of docker image tags, values:{version,latest}")
	addSliceFlag(dockerBuildDevCmd, "docker-dev-custom-tags", "", []string{}, "add your custom image tags")

	projectCmd.AddCommand(dockerPushCmd)
	addFlagRef(dockerPushCmd, fFile)
	addFlagRef(dockerPushCmd, fType)
	addFlagRef(dockerPushCmd, fDockerRepository)
	addFlagRef(dockerPushCmd, fDockerRepoPrefix)
	addFlagRef(dockerPushCmd, fDockerRepo)
	addFlagRef(dockerPushCmd, fHashLength)
	addFlagRef(dockerPushCmd, fBuildNumberPrefix)
	addFlagRef(dockerPushCmd, fBuildNumberLength)
	addSliceFlag(dockerPushCmd, "docker-push-tags", "", []string{"build-version", "hash"}, "the list of docker image tags to be push, values:{version,build-version,branch,latest,dev,hash}")
	addSliceFlag(dockerPushCmd, "docker-push-custom-tags", "", []string{}, "add your custom image tags to be push")

	projectCmd.AddCommand(dockerReleaseCmd)
	addFlagRef(dockerReleaseCmd, fFile)
	addFlagRef(dockerReleaseCmd, fType)
	addFlagRef(dockerReleaseCmd, fHashLength)
	addFlagRef(dockerReleaseCmd, fBuildNumberPrefix)
	addFlagRef(dockerReleaseCmd, fBuildNumberLength)
	addFlagRef(dockerReleaseCmd, fDockerRepository)
	addFlagRef(dockerReleaseCmd, fDockerRepo)
	addFlagRef(dockerReleaseCmd, fDockerRepoPrefix)
	addFlag(dockerReleaseCmd, "docker-release-registry", "", "", "the docker release registry")
	addFlag(dockerReleaseCmd, "docker-release-repo-prefix", "", "", "the docker release repository prefix")
	addFlag(dockerReleaseCmd, "docker-release-repository", "", "", "the docker release repository. Default value project name.")
	addBoolFlag(dockerReleaseCmd, "docker-release-skip-push", "", false, "skip docker push of release image to registry")

	projectCmd.AddCommand(helmFilterCmd)
	fHelmInput := addFlag(helmFilterCmd, "helm-input", "", "helm", "filter project helm chart input directory")
	fHelmOutput := addFlag(helmFilterCmd, "helm-output", "", "target/helm", "filter project helm chart output directory")
	fHelmClean := addBoolFlag(helmFilterCmd, "helm-clean", "", false, "clean output directory before filter")
	addFlagRequired(helmFilterCmd, "helm-filter-template", "", "maven", "Use the maven template for filter")

	projectCmd.AddCommand(helmBuildCmd)
	addFlagRef(helmBuildCmd, fHelmInput)
	addFlagRef(helmBuildCmd, fHelmOutput)
	addFlagRef(helmBuildCmd, fHelmClean)
	addSliceFlag(helmBuildCmd, "helm-build-tags", "", []string{"build-version"}, "the list of docker image tags, values:{version,build-version,branch,latest,dev,hash}")
	addSliceFlag(helmBuildCmd, "helm-push-tags", "", []string{"build-version"}, "helm push tag version, values:{version,build-version,branch,latest,dev,hash}")
	fHelmSkipPush := addBoolFlag(helmBuildCmd, "helm-skip-push", "", false, "skip helm push")

	projectCmd.AddCommand(helmReleaseCmd)
	addFlagRef(dockerReleaseCmd, fHelmInput)
	addFlagRef(dockerReleaseCmd, fHelmOutput)
	addFlagRef(dockerReleaseCmd, fHelmClean)
	addFlagRef(dockerReleaseCmd, fHelmSkipPush)
	addFlag(dockerReleaseCmd, "helm-update-chart", "", "version={{ .Version }},appVersion={{ .Version }}", "list of key value to be replaced in the Chart.yaml")
	addFlag(dockerReleaseCmd, "helm-update-values", "", "image.tag={{ .Version }}", "list of key value to be replaced in the values.yaml")

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
