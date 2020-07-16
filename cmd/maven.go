package cmd

import (
	"github.com/spf13/viper"

	"github.com/lorislab/samo/internal"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type mavenFlags struct {
	Filename                    string `mapstructure:"maven-file"`
	PatchBranchPrefix           string `mapstructure:"maven-patch-branch-prefix"`
	PatchSkipPush               bool   `mapstructure:"maven-patch-skip-push"`
	PatchMsg                    string `mapstructure:"maven-patch-message"`
	ReleaseSkipPush             bool   `mapstructure:"maven-release-skip-push"`
	DevMsg                      string `mapstructure:"maven-release-message"`
	PatchTag                    string `mapstructure:"maven-patch-tag"`
	ReleaseMajor                bool   `mapstructure:"maven-release-major"`
	HashLength                  int    `mapstructure:"maven-hash-length"`
	Dockerfile                  string `mapstructure:"maven-dockerfile"`
	DockerContext               string `mapstructure:"maven-docker-context"`
	DockerBranch                bool   `mapstructure:"maven-docker-branch"`
	DockerLatest                bool   `mapstructure:"maven-docker-latest"`
	DockerRepository            string `mapstructure:"maven-docker-repo"`
	DockerLib                   string `mapstructure:"maven-docker-lib"`
	DockerImage                 string `mapstructure:"maven-docker-image"`
	DockerBuildTag              string `mapstructure:"maven-docker-tag"`
	DockerIgnoreLatest          bool   `mapstructure:"maven-docker-ignore-latest"`
	DockerSkipPush              bool   `mapstructure:"maven-docker-skip-push"`
	MavenSettingsFile           string `mapstructure:"maven-settings-file"`
	MavenSettingsServerID       string `mapstructure:"maven-settings-server-id"`
	MavenSettingsServerUsername string `mapstructure:"maven-settings-server-username"`
	MavenSettingsServerPassword string `mapstructure:"maven-settings-server-password"`
	ReleaseTagMessage           string `mapstructure:"maven-release-tag-message"`
	BuildNumberPrefix           string `mapstructure:"maven-build-number-prefix"`
	BuildNumberLength           int    `mapstructure:"maven-build-number-length"`
	DockerDevTag                bool   `mapstructure:"maven-docker-dev"`
	DockerReleaseRepository     string `mapstructure:"maven-docker-release-repo"`
	DockerReleaseLib            string `mapstructure:"maven-docker-release-lib"`
	DockerReleaseImage          string `mapstructure:"maven-docker-release-image"`
	DockerReleaseSkipPush       bool   `mapstructure:"maven-docker-release-skip-push"`
}

func init() {
	mvnCmd.AddCommand(mvnVersionCmd)
	mavenFile := addFlag(mvnVersionCmd, "maven-file", "f", "pom.xml", "maven project file")

	mvnCmd.AddCommand(mvnSetReleaseVersionCmd)
	addFlagRef(mvnSetReleaseVersionCmd, mavenFile)

	mvnCmd.AddCommand(mvnReleaseVersionCmd)
	addFlagRef(mvnReleaseVersionCmd, mavenFile)

	mvnCmd.AddCommand(mvnSetBuildVersionCmd)
	addFlagRef(mvnSetBuildVersionCmd, mavenFile)
	buildNumberPrefix := addFlag(mvnSetBuildVersionCmd, "maven-build-number-prefix", "b", "rc", "the build number prefix")
	buildNumberLength := addIntFlag(mvnSetBuildVersionCmd, "maven-build-number-length", "e", 3, "the build number length")
	mavenHashLength := addGitHashLength(mvnSetBuildVersionCmd, "maven-hash-length", "")

	mvnCmd.AddCommand(mvnBuildVersionCmd)
	addFlagRef(mvnBuildVersionCmd, mavenFile)
	addFlagRef(mvnBuildVersionCmd, buildNumberPrefix)
	addFlagRef(mvnBuildVersionCmd, buildNumberLength)
	addFlagRef(mvnBuildVersionCmd, mavenHashLength)

	mvnCmd.AddCommand(mvnCreateReleaseCmd)
	addFlagRef(mvnCreateReleaseCmd, mavenFile)
	addFlag(mvnCreateReleaseCmd, "maven-release-message", "", "Create new development version", "commit message for new development version")
	addBoolFlag(mvnCreateReleaseCmd, "maven-release-major", "", false, "create a major release")
	addFlag(mvnCreateReleaseCmd, "maven-release-tag-message", "", "", "the release tag message")
	addBoolFlag(mvnCreateReleaseCmd, "maven-release-skip-push", "", false, "skip git push release")

	mvnCmd.AddCommand(mvnCreatePatchCmd)
	addFlagRef(mvnCreatePatchCmd, mavenFile)
	addFlagRequired(mvnCreatePatchCmd, "maven-patch-tag", "", "", "the tag version of the patch branch")
	addFlag(mvnCreatePatchCmd, "maven-patch-message", "", "Create new patch version", "commit message for new patch version")
	addFlag(mvnCreatePatchCmd, "maven-patch-branch-prefix", "", "", "patch branch prefix")
	addBoolFlag(mvnCreatePatchCmd, "maven-patch-skip-push", "", false, "skip git push patch branch.")

	mvnCmd.AddCommand(dockerBuildCmd)
	addFlagRef(dockerBuildCmd, mavenFile)
	addFlagRef(dockerBuildCmd, mavenHashLength)
	mavenDockerFile := addFlag(dockerBuildCmd, "maven-dockerfile", "d", "src/main/docker/Dockerfile", "maven project dockerfile")
	mavenDockerRepository := addFlag(dockerBuildCmd, "maven-docker-repo", "", "", "the docker repository")
	mavenDockerLib := addFlag(dockerBuildCmd, "maven-docker-lib", "", "", "the docker repository library")
	mavenDockerImage := addFlag(dockerBuildCmd, "maven-docker-image", "i", "", "the docker image. Default value maven project artifactId.")
	mavenDockerContext := addFlag(dockerBuildCmd, "maven-docker-context", "", ".", "the docker build context")
	addFlag(dockerBuildCmd, "maven-docker-tag", "", "", "add the extra tag to the build image")
	addBoolFlag(dockerBuildCmd, "maven-docker-branch", "", true, "tag the docker image with a branch name")
	addBoolFlag(dockerBuildCmd, "maven-docker-latest", "", true, "tag the docker image with a latest")
	addBoolFlag(dockerBuildCmd, "maven-docker-dev", "", true, "tag the docker image for local development")

	mvnCmd.AddCommand(dockerBuildDevCmd)
	addFlagRef(dockerBuildDevCmd, mavenFile)
	addFlagRef(dockerBuildDevCmd, mavenDockerImage)
	addFlagRef(dockerBuildDevCmd, mavenDockerFile)
	addFlagRef(dockerBuildDevCmd, mavenDockerContext)

	mvnCmd.AddCommand(dockerPushCmd)
	addFlagRef(dockerPushCmd, mavenFile)
	addFlagRef(dockerPushCmd, mavenDockerRepository)
	addFlagRef(dockerPushCmd, mavenDockerLib)
	addFlagRef(dockerPushCmd, mavenDockerImage)
	addBoolFlag(dockerPushCmd, "maven-docker-ignore-latest", "", true, "ignore push latest tag to repository")
	addBoolFlag(dockerPushCmd, "maven-docker-skip-push", "", false, "skip docker push")

	mvnCmd.AddCommand(dockerReleaseCmd)
	addFlagRef(dockerReleaseCmd, mavenFile)
	addFlagRef(dockerReleaseCmd, mavenHashLength)
	addFlagRef(dockerReleaseCmd, mavenDockerRepository)
	addFlagRef(dockerReleaseCmd, mavenDockerImage)
	addFlagRef(dockerReleaseCmd, mavenDockerLib)
	addFlag(dockerReleaseCmd, "maven-docker-release-repo", "", "", "the docker release repository")
	addFlag(dockerReleaseCmd, "maven-docker-release-lib", "", "", "the docker release repository library")
	addFlag(dockerReleaseCmd, "maven-docker-release-image", "", "", "the docker release image. Default value maven project artifactId.")
	addBoolFlag(dockerReleaseCmd, "maven-docker-release-skip-push", "", false, "skip docker push of release image")

	mvnCmd.AddCommand(settingsAddServer)
	addFlag(settingsAddServer, "maven-settings-file", "s", ".m2/settings.xml", "the maven settings.xml file")
	addFlag(settingsAddServer, "maven-settings-server-id", "", "github", "the maven repository server id")
	addFlag(settingsAddServer, "maven-settings-server-username", "", "x-access-token", "the maven repository server username")
	addFlag(settingsAddServer, "maven-settings-server-password", "", "${env.GITHUB_TOKEN}", "the maven repository server password")
}

var (

	// MvnCmd the maven command
	mvnCmd = &cobra.Command{
		Use:              "maven",
		Short:            "Maven operation",
		Long:             `Tasks for the maven project`,
		TraverseChildren: true,
	}
	mvnVersionCmd = &cobra.Command{
		Use:   "version",
		Short: "Show the maven project version",
		Long:  `Tasks to show the maven project version`,
		Run: func(cmd *cobra.Command, args []string) {
			_, project := readMavenOptions()
			projectVersion(project)
		},
		TraverseChildren: true,
	}
	mvnSetBuildVersionCmd = &cobra.Command{
		Use:   "set-build-version",
		Short: "Set the maven project version to build version",
		Long:  `Set the maven project version to build version`,
		Run: func(cmd *cobra.Command, args []string) {
			options, project := readMavenOptions()
			projectSetBuildVersion(project, options.HashLength, options.BuildNumberLength, options.BuildNumberPrefix)
		},
		TraverseChildren: true,
	}
	mvnBuildVersionCmd = &cobra.Command{
		Use:   "build-version",
		Short: "Show the maven project version to build version",
		Long:  `Show the maven project version to build version`,
		Run: func(cmd *cobra.Command, args []string) {
			options, project := readMavenOptions()
			projectBuildVersion(project, options.HashLength, options.BuildNumberLength, options.BuildNumberPrefix)
		},
		TraverseChildren: true,
	}
	mvnSetReleaseVersionCmd = &cobra.Command{
		Use:   "set-release-version",
		Short: "Set the maven project version to release",
		Long:  `Set the maven project version to release`,
		Run: func(cmd *cobra.Command, args []string) {
			_, project := readMavenOptions()
			projectSetReleaseVersion(project)
		},
		TraverseChildren: true,
	}
	mvnReleaseVersionCmd = &cobra.Command{
		Use:   "release-version",
		Short: "Show the maven project version to release",
		Long:  `Show the maven project version to release`,
		Run: func(cmd *cobra.Command, args []string) {
			_, project := readMavenOptions()
			projectReleaseVersion(project)
		},
		TraverseChildren: true,
	}
	mvnCreateReleaseCmd = &cobra.Command{
		Use:   "create-release",
		Short: "Create release of the current project and state",
		Long:  `Create release of the current project and state`,
		Run: func(cmd *cobra.Command, args []string) {
			options, project := readMavenOptions()
			projectCreateRelease(project, options.ReleaseTagMessage, options.DevMsg, options.ReleaseMajor, options.ReleaseSkipPush)
		},
		TraverseChildren: true,
	}
	mvnCreatePatchCmd = &cobra.Command{
		Use:   "create-patch",
		Short: "Create patch of the release",
		Long:  `Create patch of the release`,
		Run: func(cmd *cobra.Command, args []string) {
			options, project := readMavenOptions()
			projectCreatePatch(project, options.PatchMsg, options.PatchTag, options.PatchBranchPrefix, options.PatchSkipPush)
		},
		TraverseChildren: true,
	}
	dockerBuildCmd = &cobra.Command{
		Use:   "docker-build",
		Short: "Build the docker image of the maven project",
		Long:  `Build the docker image of the maven project`,
		Run: func(cmd *cobra.Command, args []string) {
			options, project := readMavenOptions()
			projectDockerBuild(
				project,
				options.DockerRepository,
				options.DockerLib,
				options.DockerImage,
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
	dockerBuildDevCmd = &cobra.Command{
		Use:   "docker-build-dev",
		Short: "Build the docker image of the maven project for local development",
		Long:  `Build the docker image of the maven project for local development`,
		Run: func(cmd *cobra.Command, args []string) {
			options, project := readMavenOptions()
			projectDockerBuildDev(project, options.DockerImage, options.Dockerfile, options.DockerContext)
		},
		TraverseChildren: true,
	}
	dockerPushCmd = &cobra.Command{
		Use:   "docker-push",
		Short: "Push the docker image of the maven project",
		Long:  `Push the docker image of the maven project`,
		Run: func(cmd *cobra.Command, args []string) {
			options, project := readMavenOptions()
			projectDockerPush(
				project,
				options.DockerRepository,
				options.DockerLib,
				options.DockerImage,
				options.DockerIgnoreLatest,
				options.DockerSkipPush)
		},
		TraverseChildren: true,
	}
	dockerReleaseCmd = &cobra.Command{
		Use:   "docker-release",
		Short: "Release the docker image",
		Long:  `Release the docker image`,
		Run: func(cmd *cobra.Command, args []string) {
			options, project := readMavenOptions()
			projectDockerRelease(
				project,
				options.DockerRepository,
				options.DockerLib,
				options.DockerImage,
				options.HashLength,
				options.DockerReleaseRepository,
				options.DockerReleaseLib,
				options.DockerReleaseImage,
				options.DockerReleaseSkipPush)
		},
		TraverseChildren: true,
	}
	settingsAddServer = &cobra.Command{
		Use:   "settings-add-server",
		Short: "Add the maven repository server to the settings",
		Long:  `Add the maven repository server configuration to the maven settings file`,
		Run: func(cmd *cobra.Command, args []string) {
			options, _ := readMavenOptions()
			internal.CreateMavenSettingsServer(options.MavenSettingsFile, options.MavenSettingsServerID, options.MavenSettingsServerUsername, options.MavenSettingsServerPassword)
		},
		TraverseChildren: true,
	}
)

func readMavenOptions() (mavenFlags, *internal.MavenProject) {
	options := mavenFlags{}
	err := viper.Unmarshal(&options)
	if err != nil {
		panic(err)
	}
	log.Debug(options)
	return options, internal.LoadMavenProject(options.Filename)
}
