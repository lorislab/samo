package cmd

import (
	"github.com/lorislab/samo/docker"
	"github.com/lorislab/samo/project"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type dockerFlags struct {
	Project                 commonFlags `mapstructure:",squash"`
	DockerRegistry          string      `mapstructure:"docker-registry"`
	DockerRepoPrefix        string      `mapstructure:"docker-repo-prefix"`
	DockerRepository        string      `mapstructure:"docker-repository"`
	Dockerfile              string      `mapstructure:"dockerfile"`
	DockerContext           string      `mapstructure:"docker-context"`
	DockerSkipPull          bool        `mapstructure:"docker-pull-skip"`
	DockerSkipPush          bool        `mapstructure:"docker-push-skip"`
	DockerReleaseRegistry   string      `mapstructure:"docker-release-registry"`
	DockerReleaseRepoPrefix string      `mapstructure:"docker-release-repo-prefix"`
	DockerReleaseRepository string      `mapstructure:"docker-release-repository"`
	DockerReleaseSkipPush   bool        `mapstructure:"docker-release-push-skip"`
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
				Project:          p,
				Registry:         op.DockerRegistry,
				RepositoryPrefix: op.DockerRepoPrefix,
				Repository:       op.DockerRepository,
				Dockerfile:       op.Dockerfile,
				Context:          op.DockerContext,
				SkipPull:         op.DockerSkipPull,
				Versions:         project.CreateVersions(p, op.Project.Versions, op.Project.HashLength, op.Project.BuildNumberLength, op.Project.BuildNumber),
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
				Versions:         project.CreateVersions(p, op.Project.Versions, op.Project.HashLength, op.Project.BuildNumberLength, op.Project.BuildNumber),
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
			docker := docker.DockerRequest{
				Project:                 p,
				Registry:                op.DockerRegistry,
				Dockerfile:              op.Dockerfile,
				Context:                 op.DockerContext,
				ReleaseRegistry:         op.DockerReleaseRegistry,
				ReleaseRepositoryPrefix: op.DockerReleaseRepoPrefix,
				ReleaseRepository:       op.DockerReleaseRepository,
				SkipPush:                op.DockerReleaseSkipPush,
				Versions:                project.CreateVersions(p, []string{project.VerBuild, project.VerRelease}, op.Project.HashLength, op.Project.BuildNumberLength, op.Project.BuildNumber),
			}
			docker.Release()
		},
		TraverseChildren: true,
	}
)

func init() {
	addChildCmd(projectCmd, dockerCmd)
	addFlag(dockerCmd, "docker-registry", "", "", "the docker registry")
	addFlag(dockerCmd, "docker-repo-prefix", "", "", "the docker repository prefix")
	addFlag(dockerCmd, "docker-repo", "", "", "the docker repository. Default value is the project name.")

	addChildCmd(dockerCmd, dockerBuildCmd)
	addFlag(dockerBuildCmd, "docker-file", "d", "src/main/docker/Dockerfile", "path of the project Dockerfile")
	addFlag(dockerBuildCmd, "docker-context", "", ".", "the docker build context")
	addBoolFlag(dockerBuildCmd, "docker-skip-pull", "", false, "skip docker pull for the build")

	addChildCmd(dockerCmd, dockerPushCmd)
	addSliceFlag(dockerPushCmd, "docker-push-tags", "", []string{"build", "hash"}, "the list of docker image tags to be push, custom or "+verList)
	addBoolFlag(dockerPushCmd, "docker-push-skip", "", false, "skip docker push of release image to registry")

	addChildCmd(dockerCmd, dockerReleaseCmd)
	addFlag(dockerReleaseCmd, "docker-release-registry", "", "", "the docker release registry")
	addFlag(dockerReleaseCmd, "docker-release-repo-prefix", "", "", "the docker release repository prefix")
	addFlag(dockerReleaseCmd, "docker-release-repository", "", "", "the docker release repository. Default value project name.")
	addBoolFlag(dockerReleaseCmd, "docker-release-push-skip", "", false, "skip docker push of release image to registry")
}

func readDockerOptions() (dockerFlags, project.Project) {
	options := dockerFlags{}
	err := viper.Unmarshal(&options)
	if err != nil {
		panic(err)
	}
	log.WithField("options", options).Debug("Load project options")
	return options, loadProject(options.Project.File, project.Type(options.Project.Type))
}
