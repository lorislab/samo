package cmd

import (
	"github.com/lorislab/samo/log"
	"github.com/lorislab/samo/tools"
	"github.com/spf13/cobra"
)

type dockerReleaseFlags struct {
	Docker          dockerFlags `mapstructure:",squash"`
	ReleaseRegistry string      `mapstructure:"docker-release-registry"`
	ReleaseGroup    string      `mapstructure:"docker-release-group"`
	ReleaseRepo     string      `mapstructure:"docker-release-repository"`
}

func createDockerReleaseCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "release",
		Short: "Release the docker image and push to release registry",
		Long:  `Release the docker image and push to release registry`,
		Run: func(cmd *cobra.Command, args []string) {
			flags := dockerReleaseFlags{}
			readOptions(&flags)
			project := loadProject(flags.Docker.Project)
			dockerRelease(project, flags)
		},
		TraverseChildren: true,
	}

	addStringFlag(cmd, "docker-release-registry", "", "", "the docker release registry")
	addStringFlag(cmd, "docker-release-group", "", "", "the docker release repository group")
	addStringFlag(cmd, "docker-release-repository", "", "", "the docker release repository. Default value project name.")

	return cmd
}

func dockerRelease(project *Project, flags dockerReleaseFlags) {

	dockerPullImage := dockerImage(project, flags.Docker.Registry, flags.Docker.Group, flags.Docker.Repo)
	imagePull := dockerImageTag(dockerPullImage, project.lastRC())
	log.Info("Pull docker image", log.F("image", imagePull))
	tools.ExecCmd("docker", "pull", imagePull)

	// check the release configuration
	if len(flags.ReleaseRegistry) == 0 {
		flags.ReleaseRegistry = flags.Docker.Registry
	}
	if len(flags.ReleaseGroup) == 0 {
		flags.ReleaseGroup = flags.Docker.Group
	}
	if len(flags.ReleaseRepo) == 0 {
		flags.ReleaseRepo = flags.Docker.Repo
	}

	// release docker registry
	dockerPushImage := dockerImage(project, flags.ReleaseRegistry, flags.ReleaseGroup, flags.ReleaseRepo)
	imagePush := dockerImageTag(dockerPushImage, project.Release())
	log.Info("Retag docker image", log.Fields{"build": imagePull, "release": imagePush})
	tools.ExecCmd("docker", "tag", imagePull, imagePush)

	if flags.Docker.Project.SkipPush {
		log.Info("Skip docker push for docker release image", log.F("image", imagePush))
	} else {
		tools.ExecCmd("docker", "push", imagePush)
		log.Info("Release docker image done!", log.F("image", imagePush))
	}
}
