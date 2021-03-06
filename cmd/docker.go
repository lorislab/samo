package cmd

import (
	"github.com/lorislab/samo/docker"
	"github.com/lorislab/samo/project"
	"github.com/spf13/cobra"
)

type dockerFlags struct {
	Project                 commonFlags `mapstructure:",squash"`
	DockerRegistry          string      `mapstructure:"docker-registry"`
	DockerRepoPrefix        string      `mapstructure:"docker-group"`
	DockerRepository        string      `mapstructure:"docker-repository"`
	Dockerfile              string      `mapstructure:"docker-file"`
	DockerfileProfile       string      `mapstructure:"docker-file-profile"`
	DockerContext           string      `mapstructure:"docker-context"`
	DockerSkipPull          bool        `mapstructure:"docker-pull-skip"`
	DockerSkipPush          bool        `mapstructure:"docker-push-skip"`
	DockerBuildpPush        bool        `mapstructure:"docker-build-push"`
	DockerSkipRemoveBuild   bool        `mapstructure:"docker-remove-intermediate-img-skip"`
	DockerReleaseRegistry   string      `mapstructure:"docker-release-registry"`
	DockerReleaseRepoPrefix string      `mapstructure:"docker-release-group"`
	DockerReleaseRepository string      `mapstructure:"docker-release-repository"`
	DockerReleaseSkipPush   bool        `mapstructure:"docker-release-push-skip"`
	DockerSkipRemoveTag     bool        `mapstructure:"docker-remove-tag-skip"`
}

var (
	dockerCmd = &cobra.Command{
		Use:              "docker",
		Short:            "Project docker operation",
		Long:             `Tasks docker tool for the project`,
		TraverseChildren: true,
	}
	dockerBuildCmd = &cobra.Command{
		Use:   "build",
		Short: "Build the docker image of the project",
		Long:  `Build the docker image of the project`,
		Run: func(cmd *cobra.Command, args []string) {
			op, p := readDockerOptions()
			docker := docker.DockerRequest{
				Project:            p,
				Registry:           op.DockerRegistry,
				RepositoryPrefix:   op.DockerRepoPrefix,
				Repository:         op.DockerRepository,
				Dockerfile:         op.Dockerfile,
				DockerfileProfile:  op.DockerfileProfile,
				Context:            op.DockerContext,
				SkipPull:           op.DockerSkipPull,
				SkipPush:           !op.DockerBuildpPush,
				SkipRemoveBuildImg: op.DockerSkipRemoveBuild,
				Versions:           createVersions(p, op.Project),
			}
			docker.Build()
		},
		TraverseChildren: true,
	}
	dockerPushCmd = &cobra.Command{
		Use:   "push",
		Short: "Push the docker image of the project",
		Long:  `Push the docker image of the project`,
		Run: func(cmd *cobra.Command, args []string) {
			op, p := readDockerOptions()
			docker := docker.DockerRequest{
				Project:          p,
				Registry:         op.DockerRegistry,
				RepositoryPrefix: op.DockerRepoPrefix,
				Repository:       op.DockerRepository,
				SkipPush:         op.DockerSkipPush,
				Versions:         createVersions(p, op.Project),
			}
			docker.Push()

		},
		TraverseChildren: true,
	}
	dockerReleaseCmd = &cobra.Command{
		Use:   "release",
		Short: "Release the docker image and push to release registry",
		Long:  `Release the docker image and push to release registry`,
		Run: func(cmd *cobra.Command, args []string) {
			op, p := readDockerOptions()

			// remove current git tag
			removeCurrentTag(op.DockerSkipRemoveTag)

			docker := docker.DockerRequest{
				Project:                 p,
				Registry:                op.DockerRegistry,
				RepositoryPrefix:        op.DockerRepoPrefix,
				Repository:              op.DockerRepository,
				Dockerfile:              op.Dockerfile,
				Context:                 op.DockerContext,
				ReleaseRegistry:         op.DockerReleaseRegistry,
				ReleaseRepositoryPrefix: op.DockerReleaseRepoPrefix,
				ReleaseRepository:       op.DockerReleaseRepository,
				SkipPush:                op.DockerReleaseSkipPush,
				Versions:                createVersionsFrom(p, op.Project, []string{project.VerBuild, project.VerRelease}),
			}
			docker.Release()
		},
		TraverseChildren: true,
	}
)

func initDocker() {
	addChildCmd(projectCmd, dockerCmd)
	addFlag(dockerCmd, "docker-registry", "", "", "the docker registry")
	addFlag(dockerCmd, "docker-group", "", "", "the docker repository group")
	addFlag(dockerCmd, "docker-repo", "", "", "the docker repository. Default value is the project name.")

	addChildCmd(dockerCmd, dockerBuildCmd)
	addFlag(dockerBuildCmd, "docker-file", "d", "src/main/docker/Dockerfile", "path of the project Dockerfile")
	addFlag(dockerBuildCmd, "docker-file-profile", "p", "", "profile of the Dockerfile.<profile>")
	addFlag(dockerBuildCmd, "docker-context", "", ".", "the docker build context")
	addBoolFlag(dockerBuildCmd, "docker-pull-skip", "", false, "skip docker pull for the build")
	addBoolFlag(dockerBuildCmd, "docker-build-push", "", false, "push docker images after build")
	addBoolFlag(dockerBuildCmd, "docker-remove-intermediate-img-skip", "", false, "skip remove build intermediate container")

	addChildCmd(dockerCmd, dockerPushCmd)
	addBoolFlag(dockerPushCmd, "docker-push-skip", "", false, "skip docker push of release image to registry")

	addChildCmd(dockerCmd, dockerReleaseCmd)
	addFlag(dockerReleaseCmd, "docker-release-registry", "", "", "the docker release registry")
	addFlag(dockerReleaseCmd, "docker-release-group", "", "", "the docker release repository group")
	addFlag(dockerReleaseCmd, "docker-release-repository", "", "", "the docker release repository. Default value project name.")
	addBoolFlag(dockerReleaseCmd, "docker-release-push-skip", "", false, "skip docker push of release image to registry")
	addBoolFlag(dockerReleaseCmd, "docker-remove-tag-skip", "", false, "remove tag from the git head")
}

func readDockerOptions() (dockerFlags, project.Project) {
	options := dockerFlags{}
	readOptions(&options)
	return options, loadProject(options.Project)
}
