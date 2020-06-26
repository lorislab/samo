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
	PatchBranchPrefix           string `mapstructure:"maven-patch-branch-prefix"`
	PatchSkipPush               bool   `mapstructure:"maven-patch-skip-push"`
	PatchMsg                    string `mapstructure:"maven-patch-message"`
	ReleaseSkipPush             bool   `mapstructure:"maven-release-skip-push"`
	DevMsg                      string `mapstructure:"maven-release-message"`
	PatchTag                    string `mapstructure:"maven-patch-tag"`
	ReleaseMajor                bool   `mapstructure:"maven-release-major"`
	HashLength                  int    `mapstructure:"maven-hash-length"`
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
	ReleaseTagMessage           string `mapstructure:"maven-release-tag-message"`
	BuildNumberPrefix           string `mapstructure:"maven-build-number-prefix"`
	BuildNumberLength           int    `mapstructure:"maven-build-number-length"`
	DevTag                      bool   `mapstructure:"maven-docker-dev"`
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
	mavenDockerImage := addFlag(dockerBuildCmd, "maven-docker-image", "i", "", "the docker image. Default value maven project artifactId.")
	addFlagRef(dockerBuildCmd, mavenHashLength)
	mavenDockerFile := addFlag(dockerBuildCmd, "maven-dockerfile", "d", "src/main/docker/Dockerfile", "maven project dockerfile")
	addFlag(dockerBuildCmd, "maven-docker-repo", "", "", "the docker repository")
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
	addFlagRef(dockerPushCmd, mavenDockerImage)
	addBoolFlag(dockerPushCmd, "maven-docker-ignore-latest", "", true, "ignore push latest tag to repository")

	mvnCmd.AddCommand(dockerReleaseCmd)
	addFlagRef(dockerReleaseCmd, mavenFile)
	addFlagRef(dockerReleaseCmd, mavenHashLength)
	addFlagRef(dockerReleaseCmd, mavenDockerImage)

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
			if !options.ReleaseSkipPush {
				execGitCmd("git", "push", "origin", "refs/heads/*:refs/heads/*", "refs/tags/*:refs/tags/*")
			} else {
				log.Info("Skip git push for release: " + releaseVersion)
			}
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

			branchName := createPatchBranchName(tagVer, options.PatchBranchPrefix)
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
			if !options.PatchSkipPush {
				execGitCmd("git", "push", "origin", "refs/heads/*:refs/heads/*")
			} else {
				log.Info("Skip git push for patch branch: " + branchName)
			}
		},
		TraverseChildren: true,
	}
	dockerBuildCmd = &cobra.Command{
		Use:   "docker-build",
		Short: "Build the docker image of the maven project",
		Long:  `Build the docker image of the maven project`,
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
			if options.DevTag {
				tmp := dockerDevImage(project, options)
				command = append(command, "-t", imageNameWithTag(tmp, "latest"))
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
	dockerBuildDevCmd = &cobra.Command{
		Use:   "docker-build-dev",
		Short: "Build the docker image of the maven project for local development",
		Long:  `Build the docker image of the maven project for local development`,
		Run: func(cmd *cobra.Command, args []string) {
			options := readMavenOptions()
			project := internal.LoadMavenProject(options.Filename)
			image := dockerDevImage(project, options)

			var command []string
			command = append(command, "build", "-t", imageNameWithTag(image, project.Version()))
			command = append(command, "-t", imageNameWithTag(image, "latest"))

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
		Short: "Push the docker image of the maven project",
		Long:  `Push the docker image of the maven project`,
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

func dockerDevImage(project *internal.MavenProject, options mavenFlags) string {
	image := options.Image
	if len(image) == 0 {
		image = project.ArtifactID()
	}
	return image
}

func dockerImage(project *internal.MavenProject, options mavenFlags) string {
	image := dockerDevImage(project, options)
	if len(options.Repository) > 0 {
		image = options.Repository + "/" + image
	}
	return image
}

func readMavenOptions() mavenFlags {
	mavenOptions := mavenFlags{}
	err := viper.Unmarshal(&mavenOptions)
	if err != nil {
		panic(err)
	}
	log.Debug(mavenOptions)
	return mavenOptions
}

func versionWithoutSnapshot(project *internal.MavenProject) *semver.Version {
	return createVersion(strings.TrimSuffix(project.Version(), "-SNAPSHOT"))
}

func mavenBuildVersion(project *internal.MavenProject, options mavenFlags) semver.Version {
	cr := versionWithoutSnapshot(project)
	_, count, hash := gitCommit(options.HashLength)
	return createProjectBuildVersion(cr.String(), count, hash, options.BuildNumberPrefix, options.BuildNumberLength)
}
