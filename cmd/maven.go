package cmd

import (
	"fmt"
	"github.com/spf13/viper"
	"strconv"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/lorislab/samo/internal"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type mavenFlags struct {
	Filename                    string `mapstructure:"maven-file"`
	PatchMsg                    string `mapstructure:"maven-patch-message"`
	DevMsg                      string `mapstructure:"maven-release-message"`
	PatchTag                    string `mapstructure:"maven-patch-tag"`
	ReleaseMajor                bool   `mapstructure:"maven-release-major"`
	HashLength                  string `mapstructure:"maven-hash-length"`
	Dockerfile                  string `mapstructure:"maven-dockerfile"`
	Context                     string `mapstructure:"maven-docker-context"`
	Branch                      bool   `mapstructure:"maven-docker-branch"`
	Latest                      bool   `mapstructure:"maven-docker-latest"`
	Repository                  string `mapstructure:"maven-docker-repo"`
	Image                       string `mapstructure:"maven-docker-image"`
	BuildTag                    string `mapstructure:"maven-docker-tag"`
	IgnoreLatestTag             bool   `mapstructure:"maven-docker-ignore-latest"`
	MavenSettingsFile           string `mapstructure:"maven-settings-file"`
	MavenSettingsServerID       string `mapstructure:"maven-settings-server-id"`
	MavenSettingsServerUsername string `mapstructure:"maven-settings-server-username"`
	MavenSettingsServerPassword string `mapstructure:"maven-settings-server-password"`
	ReleaseTagMessage           string `mapstructure:"release-tag-message"`
	BuildNumberPrefix           string `mapstructure:"build-number-prefix"`
}

func init() {
	mvnCmd.AddCommand(mvnVersionCmd)
	addMavenFlags(mvnVersionCmd)

	mvnCmd.AddCommand(mvnSetReleaseCmd)
	addMavenFlags(mvnSetReleaseCmd)

	mvnCmd.AddCommand(mvnSetBuildCmd)
	addMavenFlags(mvnSetBuildCmd)
	addFlag(mvnSetBuildCmd, "build-number-prefix", "b", "rc", "The build number prefix")

	mvnCmd.AddCommand(mvnCreateReleaseCmd)
	addMavenFlags(mvnCreateReleaseCmd)
	addFlag(mvnCreateReleaseCmd, "maven-release-message", "m", "Create new development version", "Commit message for new development version")
	addBoolFlag(mvnCreateReleaseCmd, "maven-release-major", "a", false, "Create a major release")
	addFlag(mvnCreateReleaseCmd, "release-tag-message", "s", "", "The release tag message")

	mvnCmd.AddCommand(mvnCreatePatchCmd)
	addMavenFlags(mvnCreatePatchCmd)
	addFlagRequired(mvnCreatePatchCmd, "maven-patch-tag", "t", "", "The tag version for the patch branch")
	addFlag(mvnCreatePatchCmd, "maven-patch-message", "m", "Create new patch version", "Commit message for new patch version")

	mvnCmd.AddCommand(dockerBuildCmd)
	addMavenFlags(dockerBuildCmd)
	addDockerImageFlags(dockerBuildCmd)
	addGitHashLength(dockerBuildCmd)
	addFlag(dockerBuildCmd, "maven-dockerfile", "d", "src/main/docker/Dockerfile", "The maven project dockerfile")
	addFlag(dockerBuildCmd, "maven-docker-repo", "r", "", "The docker repository")
	addFlag(dockerBuildCmd, "maven-docker-context", "c", ".", "The docker build context")
	addFlag(dockerBuildCmd, "maven-docker-tag", "t", "", "Add the extra tag to the build image")
	addBoolFlag(dockerBuildCmd, "maven-docker-branch", "h", true, "Tag the docker image with a branch name")
	addBoolFlag(dockerBuildCmd, "maven-docker-latest", "e", true, "Tag the docker image with a latest")

	mvnCmd.AddCommand(dockerPushCmd)
	addMavenFlags(dockerPushCmd)
	addDockerImageFlags(dockerPushCmd)
	addBoolFlag(dockerPushCmd, "maven-docker-ignore-latest", "p", true, "Ignore push latest tag to repository")

	mvnCmd.AddCommand(dockerReleaseCmd)
	addMavenFlags(dockerReleaseCmd)
	addGitHashLength(dockerReleaseCmd)
	addDockerImageFlags(dockerReleaseCmd)

	mvnCmd.AddCommand(settingsAddServer)
	addFlag(settingsAddServer, "maven-settings-file", "s", ".m2/settings.xml", "The maven settings.xml file")
	addFlag(settingsAddServer, "maven-settings-server-id", "", "github", "The maven repository server id")
	addFlag(settingsAddServer, "maven-settings-server-username", "", "x-access-token", "The maven repository server username")
	addFlag(settingsAddServer, "maven-settings-server-password", "", "${env.GITHUB_TOKEN}", "The maven repository server password")
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
	mvnSetBuildCmd = &cobra.Command{
		Use:   "set-build-version",
		Short: "Set the maven project version to build version",
		Long:  `Set the maven project version to build version`,
		Run: func(cmd *cobra.Command, args []string) {
			options := readMavenOptions()
			project := internal.LoadMavenProject(options.Filename)

			_, count, build := gitCommit()
			major, minor, patch, prerelease := createBuildVersionItems(versionWithoutSnapshot(project))
			ver := createBuildVersionFromItems(major, minor, patch, prerelease, options.BuildNumberPrefix, count, build)
			project.SetVersion(ver.String())
		},
		TraverseChildren: true,
	}
	mvnSetReleaseCmd = &cobra.Command{
		Use:   "set-release-version",
		Short: "Set the maven project version to release",
		Long:  `Set the maven project version to release`,
		Run: func(cmd *cobra.Command, args []string) {
			options := readMavenOptions()
			project := internal.LoadMavenProject(options.Filename)
			ver := versionWithoutSnapshot(project)
			project.SetVersion(ver.String())
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
			ver := versionWithoutSnapshot(project)

			releaseVersion := ver.String()
			msg := options.ReleaseTagMessage
			if len(msg) == 0 {
				msg = releaseVersion
			}
			execGitCmd("git", "tag", "-a", releaseVersion, "-m", msg)

			newVersion := snapshotPrerelease(nextReleaseVersion(ver, options.ReleaseMajor))
			project.SetVersion(newVersion.String())

			execGitCmd("git", "add", ".")
			execGitCmd("git", "commit", "-m", options.DevMsg+" ["+newVersion.String()+"]")
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

			// remove the prelrease
			ver := *tagVer
			if len(ver.Prerelease()) > 0 {
				ver, e = ver.SetPrerelease("")
				if e != nil {
					log.Panic(e)
				}
			}
			ver = snapshotPrerelease(ver.IncPatch())
			project := internal.LoadMavenProject(options.Filename)
			project.SetVersion(ver.String())

			execGitCmd("git", "add", ".")
			execGitCmd("git", "commit", "-m", options.PatchMsg+" ["+ver.String()+"]")
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

			_, count, hash := gitCommit()
			major, minor, patch, prerelease := createBuildVersionItems(versionWithoutSnapshot(project))
			ver := createBuildVersionFromItems(major, minor, patch, prerelease, options.BuildNumberPrefix, count, hash)
			releaseTag := createBuildVersionFromItems(major, minor, patch, prerelease, "", "", hash)

			var command []string
			command = append(command, "build", "-t", imageNameWithTag(image, ver.String()))

			// add only the build version <version>+<hash>
			command = append(command, "-t", imageNameWithTag(image, releaseTag.String()))

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

			if options.IgnoreLatestTag {
				tag := imageNameWithTag(image, "latest")
				output := execCmdOutput("docker", "images", "-q", tag)
				if len(output) > 0 {
					execCmd("docker", "rmi", tag)
				}
			}
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

			_, _, build := gitCommit()

			project := internal.LoadMavenProject(options.Filename)
			releaseVersion := project.Version()

			ver := versionWithoutSnapshot(project)
			pullVersion, err := ver.SetMetadata(build)
			if err != nil {
				log.Panic(err)
			}

			image := dockerImage(project, options)

			imagePull := imageNameWithTag(image, pullVersion.String())
			execCmd("docker", "pull", imagePull)

			imageRelease := imageNameWithTag(image, releaseVersion)
			execCmd("docker", "tag", imagePull, imageRelease)

			execCmd("docker", "push", imageRelease)
		},
		TraverseChildren: true,
	}
	settingsAddServer = &cobra.Command{
		Use:   "settings-add-server",
		Short: "Add the maven repository server to the settings",
		Long:  `Add the maven repository server configuration to the maven settings file`,
		Run: func(cmd *cobra.Command, args []string) {
			options := readMavenOptions()
			internal.CreateMavenSettingsServer(options.MavenSettingsFile, options.MavenSettingsServerID, options.MavenSettingsServerUsername, options.MavenSettingsServerPassword)
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

func snapshotPrerelease(ver semver.Version) semver.Version {
	tmp := "SNAPSHOT"
	if len(ver.Prerelease()) > 0 {
		tmp = ver.Prerelease() + "-" + tmp
	}
	return setPrerelease(ver, tmp)
}

func setPrerelease(ver semver.Version, prerelease string) semver.Version {
	result, err := ver.SetPrerelease(prerelease)
	if err != nil {
		log.Panic(err)
	}
	return result
}

func versionWithoutSnapshot(project *internal.MavenProject) *semver.Version {
	v, e := semver.NewVersion(strings.TrimSuffix(project.Version(), "-SNAPSHOT"))
	if e != nil {
		log.Panic(e)
	}
	return v
}
