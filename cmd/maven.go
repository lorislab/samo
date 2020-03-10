package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"

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
	BuildNumberLength           int    `mapstructure:"build-number-length"`
}

func init() {
	mvnCmd.AddCommand(mvnVersionCmd)
	addMavenFlags(mvnVersionCmd)

	mvnCmd.AddCommand(mvnSetReleaseVersionCmd)
	addMavenFlags(mvnSetReleaseVersionCmd)

	mvnCmd.AddCommand(mvnReleaseVersionCmd)
	addMavenFlags(mvnReleaseVersionCmd)

	mvnCmd.AddCommand(mvnSetBuildVersionCmd)
	addMavenFlags(mvnSetBuildVersionCmd)
	addFlag(mvnSetBuildVersionCmd, "build-number-prefix", "b", "rc", "The build number prefix")
	addIntFlag(mvnSetBuildVersionCmd, "build-number-length", "e", 3, "The build number length")
	addGitHashLength(mvnSetBuildVersionCmd, "maven-hash-length")

	mvnCmd.AddCommand(mvnBuildVersionCmd)
	addMavenFlags(mvnBuildVersionCmd)
	addFlag(mvnBuildVersionCmd, "build-number-prefix", "b", "rc", "The build number prefix")
	addIntFlag(mvnBuildVersionCmd, "build-number-length", "e", 3, "The build number length")
	addGitHashLength(mvnBuildVersionCmd, "maven-hash-length")

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
	addGitHashLength(dockerBuildCmd, "maven-hash-length")
	addFlag(dockerBuildCmd, "maven-dockerfile", "d", "src/main/docker/Dockerfile", "The maven project dockerfile")
	addFlag(dockerBuildCmd, "maven-docker-repo", "r", "", "The docker repository")
	addFlag(dockerBuildCmd, "maven-docker-context", "c", ".", "The docker build context")
	addFlag(dockerBuildCmd, "maven-docker-tag", "t", "", "Add the extra tag to the build image")
	addBoolFlag(dockerBuildCmd, "maven-docker-branch", "k", true, "Tag the docker image with a branch name")
	addBoolFlag(dockerBuildCmd, "maven-docker-latest", "e", true, "Tag the docker image with a latest")

	mvnCmd.AddCommand(dockerPushCmd)
	addMavenFlags(dockerPushCmd)
	addDockerImageFlags(dockerPushCmd)
	addBoolFlag(dockerPushCmd, "maven-docker-ignore-latest", "p", true, "Ignore push latest tag to repository")

	mvnCmd.AddCommand(dockerReleaseCmd)
	addMavenFlags(dockerReleaseCmd)
	addGitHashLength(dockerReleaseCmd, "maven-hash-length")
	addDockerImageFlags(dockerReleaseCmd)

	mvnCmd.AddCommand(settingsAddServer)
	addFlag(settingsAddServer, "maven-settings-file", "s", ".m2/settings.xml", "The maven settings.xml file")
	addFlag(settingsAddServer, "maven-settings-server-id", "", "github", "The maven repository server id")
	addFlag(settingsAddServer, "maven-settings-server-username", "", "x-access-token", "The maven repository server username")
	addFlag(settingsAddServer, "maven-settings-server-password", "", "${env.GITHUB_TOKEN}", "The maven repository server password")
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
	mvnSetBuildVersionCmd = &cobra.Command{
		Use:   "set-build-version",
		Short: "Set the maven project version to build version",
		Long:  `Set the maven project version to build version`,
		Run: func(cmd *cobra.Command, args []string) {
			options := readMavenOptions()
			project := internal.LoadMavenProject(options.Filename)
			ver := mavenBuildVersion(project, options)
			project.SetVersion(ver.String())
		},
		TraverseChildren: true,
	}
	mvnBuildVersionCmd = &cobra.Command{
		Use:   "build-version",
		Short: "Show the maven project version to build version",
		Long:  `Show the maven project version to build version`,
		Run: func(cmd *cobra.Command, args []string) {
			options := readMavenOptions()
			project := internal.LoadMavenProject(options.Filename)
			ver := mavenBuildVersion(project, options)
			fmt.Printf("%s\n", ver.String())
		},
		TraverseChildren: true,
	}
	mvnSetReleaseVersionCmd = &cobra.Command{
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
	mvnReleaseVersionCmd = &cobra.Command{
		Use:   "release-version",
		Short: "Show the maven project version to release",
		Long:  `Show the maven project version to release`,
		Run: func(cmd *cobra.Command, args []string) {
			options := readMavenOptions()
			project := internal.LoadMavenProject(options.Filename)
			ver := versionWithoutSnapshot(project)
			fmt.Printf("%s\n", ver.String())
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

			newVersion := addPrerelease(nextReleaseVersion(ver, options.ReleaseMajor), "SNAPSHOT")
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

			branchName := createPatchBranchName(tagVer)
			execGitCmd("git", "checkout", "-b", branchName, options.PatchTag)

			// remove the prerelease
			ver := *tagVer
			if len(ver.Prerelease()) > 0 {
				ver = setPrerelease(ver, "")
			}
			ver = addPrerelease(ver.IncPatch(), "SNAPSHOT")
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
			ver := versionWithoutSnapshot(project)

			pre := updatePrereleaseToHashVersion(ver.Prerelease(), options.HashLength)
			gitHashVer := setPrerelease(*ver, pre)

			var command []string
			command = append(command, "build", "-t", imageNameWithTag(image, project.Version()))

			command = append(command, "-t", imageNameWithTag(image, gitHashVer.String()))
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

			// x.x.x
			project := internal.LoadMavenProject(options.Filename)
			releaseVersion := versionWithoutSnapshot(project)

			// x.x.x-hash
			_, _, hash := gitCommit(options.HashLength)
			ver := versionWithoutSnapshot(project)
			pullVersion := addPrerelease(*ver, hash)

			image := dockerImage(project, options)
			imagePull := imageNameWithTag(image, pullVersion.String())
			execCmd("docker", "pull", imagePull)

			imageRelease := imageNameWithTag(image, releaseVersion.String())
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

func versionWithoutSnapshot(project *internal.MavenProject) *semver.Version {
	return createVersion(strings.TrimSuffix(project.Version(), "-SNAPSHOT"))
}

func mavenBuildVersion(project *internal.MavenProject, options mavenFlags) semver.Version {
	cr := versionWithoutSnapshot(project)
	_, count, hash := gitCommit(options.HashLength)
	return createMavenBuildVersion(cr.String(), count, hash, options.BuildNumberPrefix, options.BuildNumberLength)
}

// x.x.x-<pre>.rc0.hash -> x.x.x-<pre>.hash
func updatePrereleaseToHashVersion(ver, length string) string {
	pre := ver
	if len(pre) > 0 {
		hi := strings.LastIndex(pre, ".")
		if hi != -1 {
			pre = pre[0:hi]
			ri := strings.LastIndex(pre, ".")
			if ri != -1 {
				pre = pre[0:ri]
			} else {
				pre = ""
			}
		}
	}
	_, _, hash := gitCommit(length)
	if len(pre) > 0 {
		pre = pre + "."
	}
	return pre + hash
}
