package cmd

import (
	"fmt"
	"strings"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"

	"github.com/Masterminds/semver"

	"github.com/lorislab/samo/internal"
	"github.com/spf13/cobra"
)

type npmFlags struct {
	Filename          string `mapstructure:"npm-file"`
	BuildNumberPrefix string `mapstructure:"npm-build-number-prefix"`
	BuildNumberLength int    `mapstructure:"npm-build-number-length"`
	HashLength        int    `mapstructure:"npm-hash-length"`
	Image             string `mapstructure:"npm-docker-image"`
	Dockerfile        string `mapstructure:"npm-dockerfile"`
	Context           string `mapstructure:"npm-docker-context"`
	Repository        string `mapstructure:"npm-docker-repo"`
	Branch            bool   `mapstructure:"npm-docker-branch"`
	Latest            bool   `mapstructure:"npm-docker-latest"`
	BuildTag          string `mapstructure:"npm-docker-tag"`
	IgnoreLatestTag   bool   `mapstructure:"npm-docker-ignore-latest"`
	DevTag            bool   `mapstructure:"npm-docker-dev"`
	ReleaseTagMessage string `mapstructure:"npm-release-tag-message"`
	DevMsg            string `mapstructure:"npm-release-message"`
	ReleaseMajor      bool   `mapstructure:"npm-release-major"`
	PatchMsg          string `mapstructure:"npm-patch-message"`
	PatchTag          string `mapstructure:"npm-patch-tag"`
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
	npmDockerImage := addFlag(npmDockerBuildCmd, "npm-docker-image", "", "", "the docker image. Default value maven project name.")
	addFlagRef(npmDockerBuildCmd, npmHashLength)
	npmDockerFile := addFlag(npmDockerBuildCmd, "npm-dockerfile", "", "Dockerfile", "the maven project dockerfile")
	addFlag(npmDockerBuildCmd, "npm-docker-repo", "", "", "the docker repository")
	npmDockerContext := addFlag(npmDockerBuildCmd, "npm-docker-context", "", ".", "the docker build context")
	addFlag(npmDockerBuildCmd, "npm-docker-tag", "", "", "add the extra tag to the build image")
	addBoolFlag(npmDockerBuildCmd, "npm-docker-branch", "", true, "tag the docker image with a branch name")
	addBoolFlag(npmDockerBuildCmd, "npm-docker-latest", "", true, "tag the docker image with a latest")
	addBoolFlag(npmDockerBuildCmd, "npm-docker-dev", "", true, "tag the docker image for local development")

	npmCmd.AddCommand(npmDockerBuildDevCmd)
	addFlagRef(npmDockerBuildDevCmd, npmFile)
	addFlagRef(npmDockerBuildDevCmd, npmDockerImage)
	addFlagRef(npmDockerBuildDevCmd, npmDockerFile)
	addFlagRef(npmDockerBuildDevCmd, npmDockerContext)

	npmCmd.AddCommand(npmDockerPushCmd)
	addFlagRef(npmDockerPushCmd, npmFile)
	addFlagRef(npmDockerPushCmd, npmDockerImage)
	addBoolFlag(npmDockerPushCmd, "npm-docker-ignore-latest", "", true, "ignore push latest tag to repository")

	npmCmd.AddCommand(npmCreateReleaseCmd)
	addFlagRef(npmCreateReleaseCmd, npmFile)
	addFlag(npmCreateReleaseCmd, "npm-release-message", "", "Create new development version", "commit message for new development version")
	addBoolFlag(npmCreateReleaseCmd, "npm-release-major", "", false, "create a major release")
	addFlag(npmCreateReleaseCmd, "npm-release-tag-message", "", "", "the release tag message")

	npmCmd.AddCommand(npmCreatePatchCmd)
	addFlagRef(npmCreatePatchCmd, npmFile)
	addFlagRequired(npmCreatePatchCmd, "npm-patch-tag", "", "", "the tag version of the patch branch")
	addFlag(npmCreatePatchCmd, "npm-patch-message", "", "Create new patch version", "commit message for new patch version")

	npmCmd.AddCommand(npmDockerReleaseCmd)
	addFlagRef(npmDockerReleaseCmd, npmFile)
	addFlagRef(npmDockerReleaseCmd, npmHashLength)
	addFlagRef(npmDockerReleaseCmd, npmDockerImage)
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
			options := readNpmOptions()
			project := internal.LoadNpmProject(options.Filename)
			fmt.Printf("%s\n", project.Version())
		},
		TraverseChildren: true,
	}
	npmSetBuildVersionCmd = &cobra.Command{
		Use:   "set-build-version",
		Short: "Set the npm project version to build version",
		Long:  `Set the npm project version to build version`,
		Run: func(cmd *cobra.Command, args []string) {
			options := readNpmOptions()
			project := internal.LoadNpmProject(options.Filename)
			ver := npmBuildVersion(project, options)
			project.SetVersion(ver.String())
		},
		TraverseChildren: true,
	}
	npmBuildVersionCmd = &cobra.Command{
		Use:   "build-version",
		Short: "Show the npm project version to build version",
		Long:  `Show the npm project version to build version`,
		Run: func(cmd *cobra.Command, args []string) {
			options := readNpmOptions()
			project := internal.LoadNpmProject(options.Filename)
			ver := npmBuildVersion(project, options)
			fmt.Printf("%s\n", ver.String())
		},
		TraverseChildren: true,
	}
	npmSetReleaseVersionCmd = &cobra.Command{
		Use:   "set-release-version",
		Short: "Set the npm project version to release",
		Long:  `Set the npm project version to release`,
		Run: func(cmd *cobra.Command, args []string) {
			options := readNpmOptions()
			project := internal.LoadNpmProject(options.Filename)
			ver := npmVersionWithoutSnapshot(project)
			project.SetVersion(ver.String())
		},
		TraverseChildren: true,
	}
	npmReleaseVersionCmd = &cobra.Command{
		Use:   "release-version",
		Short: "Show the npm project version to release",
		Long:  `Show the npm project version to release`,
		Run: func(cmd *cobra.Command, args []string) {
			options := readNpmOptions()
			project := internal.LoadNpmProject(options.Filename)
			ver := npmVersionWithoutSnapshot(project)
			fmt.Printf("%s\n", ver.String())
		},
		TraverseChildren: true,
	}
	npmDockerBuildCmd = &cobra.Command{
		Use:   "docker-build",
		Short: "Build the docker image of the maven project",
		Long:  `Build the docker image of the maven project`,
		Run: func(cmd *cobra.Command, args []string) {
			options := readNpmOptions()
			project := internal.LoadNpmProject(options.Filename)
			image := npmDockerImage(project, options)
			ver := npmVersionWithoutSnapshot(project)

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
				tmp := npmDockerDevImage(project, options)
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
	npmDockerBuildDevCmd = &cobra.Command{
		Use:   "docker-build-dev",
		Short: "Build the docker image of the npm project for local development",
		Long:  `Build the docker image of the npm project for local development`,
		Run: func(cmd *cobra.Command, args []string) {
			options := readNpmOptions()
			project := internal.LoadNpmProject(options.Filename)
			image := npmDockerDevImage(project, options)

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
	npmDockerPushCmd = &cobra.Command{
		Use:   "docker-push",
		Short: "Push the docker image of the npm project",
		Long:  `Push the docker image of the npm project`,
		Run: func(cmd *cobra.Command, args []string) {
			options := readNpmOptions()
			project := internal.LoadNpmProject(options.Filename)
			image := npmDockerImage(project, options)

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
	npmCreateReleaseCmd = &cobra.Command{
		Use:   "create-release",
		Short: "Create release of the current project and state",
		Long:  `Create release of the current project and state`,
		Run: func(cmd *cobra.Command, args []string) {
			options := readNpmOptions()
			project := internal.LoadNpmProject(options.Filename)
			ver := npmVersionWithoutSnapshot(project)

			releaseVersion := ver.String()
			msg := options.ReleaseTagMessage
			if len(msg) == 0 {
				msg = releaseVersion
			}
			execGitCmd("git", "tag", "-a", releaseVersion, "-m", msg)

			newVersion := addPrerelease(nextReleaseVersion(ver, options.ReleaseMajor), "")
			project.SetVersion(newVersion.String())

			execGitCmd("git", "add", ".")
			execGitCmd("git", "commit", "-m", options.DevMsg+" ["+newVersion.String()+"]")
			execGitCmd("git", "push", "origin", "refs/heads/*:refs/heads/*", "refs/tags/*:refs/tags/*")
		},
		TraverseChildren: true,
	}
	npmCreatePatchCmd = &cobra.Command{
		Use:   "create-patch",
		Short: "Create patch of the release",
		Long:  `Create patch of the release`,
		Run: func(cmd *cobra.Command, args []string) {
			options := readNpmOptions()
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
			ver = addPrerelease(ver.IncPatch(), "")
			project := internal.LoadNpmProject(options.Filename)
			project.SetVersion(ver.String())

			execGitCmd("git", "add", ".")
			execGitCmd("git", "commit", "-m", options.PatchMsg+" ["+ver.String()+"]")
			execGitCmd("git", "push", "origin", "refs/heads/*:refs/heads/*")

		},
		TraverseChildren: true,
	}
	npmDockerReleaseCmd = &cobra.Command{
		Use:   "docker-release",
		Short: "Release the docker image",
		Long:  `Release the docker image`,
		Run: func(cmd *cobra.Command, args []string) {
			options := readNpmOptions()

			// x.x.x
			project := internal.LoadNpmProject(options.Filename)
			releaseVersion := npmVersionWithoutSnapshot(project)

			// x.x.x-hash
			_, _, hash := gitCommit(options.HashLength)
			ver := npmVersionWithoutSnapshot(project)
			pullVersion := addPrerelease(*ver, hash)

			image := npmDockerImage(project, options)
			imagePull := imageNameWithTag(image, pullVersion.String())
			execCmd("docker", "pull", imagePull)

			imageRelease := imageNameWithTag(image, releaseVersion.String())
			execCmd("docker", "tag", imagePull, imageRelease)
			execCmd("docker", "push", imageRelease)
		},
		TraverseChildren: true,
	}
)

func npmDockerImage(project *internal.NpmProject, options npmFlags) string {
	image := npmDockerDevImage(project, options)
	if len(options.Repository) > 0 {
		image = options.Repository + "/" + image
	}
	return image
}

func npmDockerDevImage(project *internal.NpmProject, options npmFlags) string {
	image := options.Image
	if len(image) == 0 {
		image = project.Name()
	}
	return image
}

func readNpmOptions() npmFlags {
	npmOptions := npmFlags{}
	err := viper.Unmarshal(&npmOptions)
	if err != nil {
		panic(err)
	}
	log.Debug(npmOptions)
	return npmOptions
}

func npmVersionWithoutSnapshot(project *internal.NpmProject) *semver.Version {
	tmp := project.Version()
	index := strings.Index(tmp, "-")
	if index != -1 {
		tmp = tmp[0:index]
	}
	return createVersion(tmp)
}

func npmBuildVersion(project *internal.NpmProject, options npmFlags) semver.Version {
	cr := npmVersionWithoutSnapshot(project)
	_, count, hash := gitCommit(options.HashLength)
	return createProjectBuildVersion(cr.String(), count, hash, options.BuildNumberPrefix, options.BuildNumberLength)
}
