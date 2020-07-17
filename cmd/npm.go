package cmd

import (
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/lorislab/samo/internal"
	"github.com/spf13/cobra"
)

type npmFlags struct {
	Filename                string `mapstructure:"npm-file"`
	BuildNumberPrefix       string `mapstructure:"npm-build-number-prefix"`
	BuildNumberLength       int    `mapstructure:"npm-build-number-length"`
	HashLength              int    `mapstructure:"npm-hash-length"`
	DockerRepository        string `mapstructure:"npm-docker-repository"`
	Dockerfile              string `mapstructure:"npm-dockerfile"`
	DockerContext           string `mapstructure:"npm-docker-context"`
	DockerRegistry          string `mapstructure:"npm-docker-registry"`
	DockerBranch            bool   `mapstructure:"npm-docker-branch"`
	DockerLatest            bool   `mapstructure:"npm-docker-latest"`
	DockerBuildTag          string `mapstructure:"npm-docker-tag"`
	DockerIgnoreLatest      bool   `mapstructure:"npm-docker-ignore-latest"`
	DockerSkipPush          bool   `mapstructure:"npm-docker-skip-push"`
	DockerDevTag            bool   `mapstructure:"npm-docker-dev"`
	DockerRepoPrefix        string `mapstructure:"npm-docker-repo-prefix"`
	ReleaseSkipPush         bool   `mapstructure:"npm-release-skip-push"`
	ReleaseTagMessage       string `mapstructure:"npm-release-tag-message"`
	DevMsg                  string `mapstructure:"npm-release-message"`
	ReleaseMajor            bool   `mapstructure:"npm-release-major"`
	PatchMsg                string `mapstructure:"npm-patch-message"`
	PatchBranchPrefix       string `mapstructure:"npm-patch-branch-prefix"`
	PatchSkipPush           bool   `mapstructure:"npm-patch-skip-push"`
	PatchTag                string `mapstructure:"npm-patch-tag"`
	DockerReleaseRegistry   string `mapstructure:"npm-docker-release-registry"`
	DockerReleaseRepoPrefix string `mapstructure:"npm-docker-release-repo-prefix"`
	DockerReleaseRepository string `mapstructure:"npm-docker-release-repository"`
	DockerReleaseSkipPush   bool   `mapstructure:"npm-docker-release-skip-push"`
}

func init() {
	npmCmd.AddCommand(npmVersionCmd)
	npmFile := addFlag(npmVersionCmd, "npm-file", "", "package.json", "npm project file")

	npmCmd.AddCommand(npmSetReleaseVersionCmd)
	addFlagRef(npmSetReleaseVersionCmd, npmFile)

	npmCmd.AddCommand(npmReleaseVersionCmd)
	addFlagRef(npmReleaseVersionCmd, npmFile)

	npmCmd.AddCommand(npmSetBuildVersionCmd)
	addFlagRef(npmSetBuildVersionCmd, npmFile)
	buildNumberPrefix := addFlag(npmSetBuildVersionCmd, "npm-build-number-prefix", "", "rc", "the build number prefix")
	buildNumberLength := addIntFlag(npmSetBuildVersionCmd, "npm-build-number-length", "", 3, "the build number length")
	npmHashLength := addGitHashLength(npmSetBuildVersionCmd, "npm-hash-length", "")

	npmCmd.AddCommand(npmBuildVersionCmd)
	addFlagRef(npmBuildVersionCmd, npmFile)
	addFlagRef(npmBuildVersionCmd, buildNumberPrefix)
	addFlagRef(npmBuildVersionCmd, buildNumberLength)
	addFlagRef(npmBuildVersionCmd, npmHashLength)

	npmCmd.AddCommand(npmDockerBuildCmd)
	addFlagRef(npmDockerBuildCmd, npmFile)
	npmDockerImage := addFlag(npmDockerBuildCmd, "npm-docker-repository", "", "", "the docker repository. Default value maven project name.")
	addFlagRef(npmDockerBuildCmd, npmHashLength)
	npmDockerFile := addFlag(npmDockerBuildCmd, "npm-dockerfile", "", "Dockerfile", "the maven project dockerfile")
	npmDockerRepository := addFlag(npmDockerBuildCmd, "npm-docker-registry", "", "", "the docker registry")
	npmDockerContext := addFlag(npmDockerBuildCmd, "npm-docker-context", "", ".", "the docker build context")
	addFlag(npmDockerBuildCmd, "npm-docker-tag", "", "", "add the extra tag to the build image")
	addBoolFlag(npmDockerBuildCmd, "npm-docker-branch", "", true, "tag the docker image with a branch name")
	addBoolFlag(npmDockerBuildCmd, "npm-docker-latest", "", true, "tag the docker image with a latest")
	addBoolFlag(npmDockerBuildCmd, "npm-docker-dev", "", true, "tag the docker image for local development")
	npmDockerLib := addFlag(npmDockerBuildCmd, "npm-docker-repo-prefix", "", "", "the docker repository prefix")

	npmCmd.AddCommand(npmDockerBuildDevCmd)
	addFlagRef(npmDockerBuildDevCmd, npmFile)
	addFlagRef(npmDockerBuildDevCmd, npmDockerImage)
	addFlagRef(npmDockerBuildDevCmd, npmDockerFile)
	addFlagRef(npmDockerBuildDevCmd, npmDockerContext)

	npmCmd.AddCommand(npmDockerPushCmd)
	addFlagRef(npmDockerPushCmd, npmFile)
	addFlagRef(npmDockerPushCmd, npmDockerRepository)
	addFlagRef(npmDockerPushCmd, npmDockerLib)
	addFlagRef(npmDockerPushCmd, npmDockerImage)
	addBoolFlag(npmDockerPushCmd, "npm-docker-ignore-latest", "", true, "ignore push latest tag to repository")
	addBoolFlag(npmDockerPushCmd, "npm-docker-skip-push", "", false, "skip docker push")

	npmCmd.AddCommand(npmCreateReleaseCmd)
	addFlagRef(npmCreateReleaseCmd, npmFile)
	addFlag(npmCreateReleaseCmd, "npm-release-message", "", "Create new development version", "commit message for new development version")
	addBoolFlag(npmCreateReleaseCmd, "npm-release-major", "", false, "create a major release")
	addFlag(npmCreateReleaseCmd, "npm-release-tag-message", "", "", "the release tag message")
	addBoolFlag(npmCreateReleaseCmd, "npm-release-skip-push", "", false, "skip git push release")

	npmCmd.AddCommand(npmCreatePatchCmd)
	addFlagRef(npmCreatePatchCmd, npmFile)
	addFlagRequired(npmCreatePatchCmd, "npm-patch-tag", "", "", "the tag version of the patch branch")
	addFlag(npmCreatePatchCmd, "npm-patch-message", "", "Create new patch version", "commit message for new patch version")
	addFlag(npmCreatePatchCmd, "npm-patch-branch-prefix", "", "", "patch branch prefix")
	addBoolFlag(npmCreatePatchCmd, "npm-patch-skip-push", "", false, "skip git push patch branch")

	npmCmd.AddCommand(npmDockerReleaseCmd)
	addFlagRef(npmDockerReleaseCmd, npmFile)
	addFlagRef(npmDockerReleaseCmd, npmHashLength)
	addFlagRef(npmDockerReleaseCmd, npmDockerRepository)
	addFlagRef(npmDockerReleaseCmd, npmDockerLib)
	addFlagRef(npmDockerReleaseCmd, npmDockerImage)
	addFlag(npmDockerReleaseCmd, "npm-docker-release-registry", "", "", "the docker release registry")
	addFlag(npmDockerReleaseCmd, "npm-docker-release-repo-prefix", "", "", "the docker release repository prefix")
	addFlag(npmDockerReleaseCmd, "npm-docker-release-repository", "", "", "the docker release repository. Default value maven project artifactId.")
	addBoolFlag(npmDockerReleaseCmd, "npm-docker-release-skip-push", "", false, "skip docker push of release image to registry")
}

var (
	// NpmCmd the npm command
	npmCmd = &cobra.Command{
		Use:              "npm",
		Short:            "Npm operation",
		Long:             `Tasks for the npm project`,
		TraverseChildren: true,
	}
	npmVersionCmd = &cobra.Command{
		Use:   "version",
		Short: "Show the npm project version",
		Long:  `Tasks to show the npm project version`,
		Run: func(cmd *cobra.Command, args []string) {
			_, project := readNpmOptions()
			projectVersion(project)
		},
		TraverseChildren: true,
	}
	npmSetBuildVersionCmd = &cobra.Command{
		Use:   "set-build-version",
		Short: "Set the npm project version to build version",
		Long:  `Set the npm project version to build version`,
		Run: func(cmd *cobra.Command, args []string) {
			options, project := readNpmOptions()
			projectSetBuildVersion(project, options.HashLength, options.BuildNumberLength, options.BuildNumberPrefix)
		},
		TraverseChildren: true,
	}
	npmBuildVersionCmd = &cobra.Command{
		Use:   "build-version",
		Short: "Show the npm project version to build version",
		Long:  `Show the npm project version to build version`,
		Run: func(cmd *cobra.Command, args []string) {
			options, project := readNpmOptions()
			projectBuildVersion(project, options.HashLength, options.BuildNumberLength, options.BuildNumberPrefix)
		},
		TraverseChildren: true,
	}
	npmSetReleaseVersionCmd = &cobra.Command{
		Use:   "set-release-version",
		Short: "Set the npm project version to release",
		Long:  `Set the npm project version to release`,
		Run: func(cmd *cobra.Command, args []string) {
			_, project := readNpmOptions()
			projectSetReleaseVersion(project)
		},
		TraverseChildren: true,
	}
	npmReleaseVersionCmd = &cobra.Command{
		Use:   "release-version",
		Short: "Show the npm project version to release",
		Long:  `Show the npm project version to release`,
		Run: func(cmd *cobra.Command, args []string) {
			_, project := readNpmOptions()
			projectReleaseVersion(project)
		},
		TraverseChildren: true,
	}
	npmCreateReleaseCmd = &cobra.Command{
		Use:   "create-release",
		Short: "Create release of the current project and state",
		Long:  `Create release of the current project and state`,
		Run: func(cmd *cobra.Command, args []string) {
			options, project := readNpmOptions()
			projectCreateRelease(project, options.ReleaseTagMessage, options.DevMsg, options.ReleaseMajor, options.ReleaseSkipPush)
		},
		TraverseChildren: true,
	}
	npmCreatePatchCmd = &cobra.Command{
		Use:   "create-patch",
		Short: "Create patch of the release",
		Long:  `Create patch of the release`,
		Run: func(cmd *cobra.Command, args []string) {
			options, project := readNpmOptions()
			projectCreatePatch(project, options.PatchMsg, options.PatchTag, options.PatchBranchPrefix, options.PatchSkipPush)
		},
		TraverseChildren: true,
	}
	npmDockerBuildCmd = &cobra.Command{
		Use:   "docker-build",
		Short: "Build the docker image of the maven project",
		Long:  `Build the docker image of the maven project`,
		Run: func(cmd *cobra.Command, args []string) {
			options, project := readNpmOptions()
			projectDockerBuild(
				project,
				options.DockerRegistry,
				options.DockerRepoPrefix,
				options.DockerRepository,
				options.HashLength,
				options.DockerBranch,
				options.DockerLatest,
				options.DockerDevTag,
				options.DockerBuildTag,
				options.Dockerfile,
				options.DockerContext)
		},
		TraverseChildren: true,
	}
	npmDockerBuildDevCmd = &cobra.Command{
		Use:   "docker-build-dev",
		Short: "Build the docker image of the npm project for local development",
		Long:  `Build the docker image of the npm project for local development`,
		Run: func(cmd *cobra.Command, args []string) {
			options, project := readNpmOptions()
			projectDockerBuildDev(project, options.DockerRepository, options.Dockerfile, options.DockerContext)
		},
		TraverseChildren: true,
	}
	npmDockerPushCmd = &cobra.Command{
		Use:   "docker-push",
		Short: "Push the docker image of the npm project",
		Long:  `Push the docker image of the npm project`,
		Run: func(cmd *cobra.Command, args []string) {
			options, project := readNpmOptions()
			projectDockerPush(
				project,
				options.DockerRegistry,
				options.DockerRepoPrefix,
				options.DockerRepository,
				options.DockerIgnoreLatest,
				options.DockerSkipPush)
		},
		TraverseChildren: true,
	}

	npmDockerReleaseCmd = &cobra.Command{
		Use:   "docker-release",
		Short: "Release the docker image and push to release registry",
		Long:  `Release the docker image and push to release registry`,
		Run: func(cmd *cobra.Command, args []string) {
			options, project := readNpmOptions()
			projectDockerRelease(
				project,
				options.DockerRegistry,
				options.DockerRepoPrefix,
				options.DockerRepository,
				options.HashLength,
				options.DockerReleaseRegistry,
				options.DockerReleaseRepoPrefix,
				options.DockerReleaseRepository,
				options.DockerReleaseSkipPush)
		},
		TraverseChildren: true,
	}
)

func readNpmOptions() (npmFlags, *internal.NpmProject) {
	options := npmFlags{}
	err := viper.Unmarshal(&options)
	if err != nil {
		panic(err)
	}
	log.Debug(options)
	return options, internal.LoadNpmProject(options.Filename)
}
