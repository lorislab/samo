package cmd

import (
	"fmt"
	"github.com/spf13/viper"
	"strconv"

	"github.com/Masterminds/semver"
	"github.com/lorislab/samo/internal"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type mavenFlags struct {
	Filename     string `mapstructure:"maven-file"`
	PatchMsg     string `mapstructure:"maven-patch-message"`
	DevMsg       string `mapstructure:"maven-release-message"`
	PatchTag     string `mapstructure:"maven-patch-tag"`
	ReleaseMajor bool   `mapstructure:"maven-release-major"`
	HashLength   string `mapstructure:"maven-hash-length"`
	Dockerfile   string `mapstructure:"maven-dockerfile"`
	Context      string `mapstructure:"maven-docker-context"`
	Branch       bool   `mapstructure:"maven-docker-branch"`
	Latest       bool   `mapstructure:"maven-docker-latest"`
	Repository   string `mapstructure:"maven-docker-repo"`
	Image        string `mapstructure:"maven-docker-image"`
	BuildTag     string `mapstructure:"maven-docker-tag"`
}

func init() {
	mvnCmd.AddCommand(mvnVersionCmd)
	addMavenFlags(mvnVersionCmd)
	mvnCmd.AddCommand(mvnSetSnapshotCmd)
	addMavenFlags(mvnSetSnapshotCmd)
	mvnCmd.AddCommand(mvnSetReleaseCmd)
	addMavenFlags(mvnSetReleaseCmd)
	mvnCmd.AddCommand(mvnSetHashCmd)
	addMavenFlags(mvnSetHashCmd)
	addGitHashLength(mvnSetHashCmd)

	mvnCmd.AddCommand(mvnCreateReleaseCmd)
	addMavenFlags(mvnCreateReleaseCmd)
	addFlag(mvnCreateReleaseCmd, "maven-release-message", "m", "Create new development version", "Commit message for new development version")
	addBoolFlag(mvnCreateReleaseCmd, "maven-release-major", "a", false, "Create a major release")

	mvnCmd.AddCommand(mvnCreatePatchCmd)
	addMavenFlags(mvnCreatePatchCmd)
	addFlagRequired(mvnCreatePatchCmd, "maven-patch-tag", "t", "", "The tag version for the patch branch")
	addFlag(mvnCreatePatchCmd, "maven-patch-message", "m", "Create new patch version", "Commit message for new patch version")

	mvnCmd.AddCommand(dockerBuildCmd)
	addMavenFlags(dockerBuildCmd)
	addDockerImageFlags(dockerBuildCmd)
	addFlag(dockerBuildCmd, "maven-dockerfile", "d", "src/main/docker/Dockerfile", "The maven project dockerfile")
	addFlag(dockerBuildCmd, "maven-docker-repo", "r", "", "The docker repository")
	addFlag(dockerBuildCmd, "maven-docker-context", "c", ".", "The docker build context")
	addFlag(dockerBuildCmd, "maven-docker-tag", "t", "", "Add the extra tag to the build image")
	addBoolFlag(dockerBuildCmd, "maven-docker-branch", "b", true, "Tag the docker image with a branch name")
	addBoolFlag(dockerBuildCmd, "maven-docker-latest", "l", true, "Tag the docker image with a latest")

	mvnCmd.AddCommand(dockerPushCmd)
	addMavenFlags(dockerPushCmd)
	addDockerImageFlags(dockerPushCmd)

	mvnCmd.AddCommand(dockerReleaseCmd)
	addMavenFlags(dockerReleaseCmd)
	addGitHashLength(dockerReleaseCmd)
	addDockerImageFlags(dockerReleaseCmd)
}

func addGitHashLength(command *cobra.Command) {
	addFlag(command, "maven-hash-length", "l", "7", "The git hash length")
}
func addDockerImageFlags(command *cobra.Command) {
	addFlag(command, "maven-docker-image", "i", "", "the docker image. Default value maven project artifactId.")
}

func addMavenFlags(command *cobra.Command) {
	addFlag(command, "maven-file", "f", "pom.xml", "The maven project file")
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
			options := readMavenOptions()
			project := internal.LoadMavenProject(options.Filename)
			fmt.Printf("%s\n", project.Version())
		},
		TraverseChildren: true,
	}
	mvnSetSnapshotCmd = &cobra.Command{
		Use:   "set-snapshot",
		Short: "Set the maven project version to snapshot",
		Long:  `Set the maven project version to snapshot`,
		Run: func(cmd *cobra.Command, args []string) {
			options := readMavenOptions()
			project := internal.LoadMavenProject(options.Filename)
			version := project.SetPrerelease("SNAPSHOT")
			project.SetVersion(version)
		},
		TraverseChildren: true,
	}
	mvnSetReleaseCmd = &cobra.Command{
		Use:   "set-release",
		Short: "Set the maven project version to release",
		Long:  `Set the maven project version to release`,
		Run: func(cmd *cobra.Command, args []string) {
			options := readMavenOptions()
			project := internal.LoadMavenProject(options.Filename)
			version := project.ReleaseVersion()
			project.SetVersion(version)
		},
		TraverseChildren: true,
	}
	mvnSetHashCmd = &cobra.Command{
		Use:   "set-hash",
		Short: "Set the maven project version to git hash",
		Long:  `Set the maven project version to git hash`,
		Run: func(cmd *cobra.Command, args []string) {
			options := readMavenOptions()
			project := internal.LoadMavenProject(options.Filename)
			hash := gitHash(options.HashLength)
			version := project.SetPrerelease(hash)
			project.SetVersion(version)
		},
		TraverseChildren: true,
	}
	mvnCreateReleaseCmd = &cobra.Command{
		Use:   "create-release",
		Short: "Create release of the current project and state",
		Long:  `Create release of the current project and state`,
		Run: func(cmd *cobra.Command, args []string) {
			options := readMavenOptions()
			project := internal.LoadMavenProject(options.Filename)
			releaseVersion := project.ReleaseVersion()

			execGitCmd("git", "tag", releaseVersion)

			newVersion := project.NextReleaseVersion(options.ReleaseMajor)
			project.SetVersion(newVersion)

			execGitCmd("git", "add", ".")
			execGitCmd("git", "commit", "-m", options.DevMsg+" ["+newVersion+"]")
			execGitCmd("git", "push", "origin", "refs/heads/*:refs/heads/*", "refs/tags/*:refs/tags/*")
		},
		TraverseChildren: true,
	}
	mvnCreatePatchCmd = &cobra.Command{
		Use:   "create-patch",
		Short: "Create patch of the release",
		Long:  `Create patch of the release`,
		Run: func(cmd *cobra.Command, args []string) {
			options := readMavenOptions()
			tagVer, e := semver.NewVersion(options.PatchTag)
			if e != nil {
				log.Panic(e)
			}

			branchName := strconv.FormatInt(tagVer.Major(), 10) + "." + strconv.FormatInt(tagVer.Minor(), 10)
			execGitCmd("git", "checkout", "-b", branchName, options.PatchTag)

			project := internal.LoadMavenProject(options.Filename)
			patch := project.NextPatchVersion()
			project.SetVersion(patch)

			execGitCmd("git", "add", ".")
			execGitCmd("git", "commit", "-m", options.PatchMsg+" ["+patch+"]")
			execGitCmd("git", "push", "origin", "refs/heads/*:refs/heads/*")

		},
		TraverseChildren: true,
	}
	dockerBuildCmd = &cobra.Command{
		Use:   "docker-build",
		Short: "Build the docker image for the maven project",
		Long:  `Build the docker image for the maven project`,
		Run: func(cmd *cobra.Command, args []string) {
			options := readMavenOptions()
			project := internal.LoadMavenProject(options.Filename)
			image := dockerImage(project, options)

			var command []string
			command = append(command, "build", "-t", imageNameWithTag(image, project.Version()))

			if options.Branch {
				branch := gitBranch()
				command = append(command, "-t", imageNameWithTag(image, branch))
			}
			if options.Latest {
				command = append(command, "-t", imageNameWithTag(image, "latest"))
			}
			if len(options.BuildTag) > 0 {
				command = append(command, "-t", imageNameWithTag(image, options.BuildTag))
			}
			if len(options.Dockerfile) > 0 {
				command = append(command, "-f", options.Dockerfile)
			}
			command = append(command, options.Context)

			execCmd("docker", command...)
		},
		TraverseChildren: true,
	}
	dockerPushCmd = &cobra.Command{
		Use:   "docker-push",
		Short: "Push the docker image for the maven project",
		Long:  `Push the docker image for the maven project`,
		Run: func(cmd *cobra.Command, args []string) {
			options := readMavenOptions()
			project := internal.LoadMavenProject(options.Filename)
			image := dockerImage(project, options)

			execCmd("docker", "push", image)
		},
		TraverseChildren: true,
	}
	dockerReleaseCmd = &cobra.Command{
		Use:   "docker-release",
		Short: "Release the docker image",
		Long:  `Release the docker image`,
		Run: func(cmd *cobra.Command, args []string) {
			options := readMavenOptions()
			project := internal.LoadMavenProject(options.Filename)
			image := dockerImage(project, options)

			hash := gitHash(options.HashLength)
			pullVersion := project.SetPrerelease(hash)
			releaseVersion := project.ReleaseVersion()

			imagePull := imageNameWithTag(image, pullVersion)
			execCmd("docker", "pull", imagePull)

			imageRelease := imageNameWithTag(image, releaseVersion)
			execCmd("docker", "tag", imagePull, imageRelease)

			execCmd("docker", "push", imageRelease)
		},
		TraverseChildren: true,
	}
)

func dockerImage(project *internal.MavenProject, options mavenFlags) string {
	image := options.Image
	if len(image) == 0 {
		image = project.ArtifactID()
	}
	if len(options.Repository) > 0 {
		image = options.Repository + "/" + image
	}
	return image
}

func imageNameWithTag(name, tag string) string {
	return name + ":" + tag
}

func readMavenOptions() mavenFlags {
	mavenOptions := mavenFlags{}
	err := viper.Unmarshal(&mavenOptions)
	if err != nil {
		panic(err)
	}
	return mavenOptions
}
