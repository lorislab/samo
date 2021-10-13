package cmd

import (
	"github.com/lorislab/samo/project"
	"github.com/lorislab/samo/tools"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type dockerReleaseFlags struct {
	Docker          dockerFlags `mapstructure:",squash"`
	ReleaseRegistry string      `mapstructure:"release-registry"`
	ReleaseGroup    string      `mapstructure:"release-group"`
	ReleaseRepo     string      `mapstructure:"release-repository"`
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

	addStringFlag(cmd, "release-registry", "", "", "the docker release registry")
	addStringFlag(cmd, "release-group", "", "", "the docker release repository group")
	addStringFlag(cmd, "release-repository", "", "", "the docker release repository. Default value project name.")

	return cmd
}

func dockerRelease(project *project.Project, flags dockerReleaseFlags) {

	dockerPullImage := dockerImage(project, flags.Docker.Registry, flags.Docker.Group, flags.Docker.Repo)
	imagePull := dockerImageTag(dockerPullImage, project.Version())
	log.WithField("image", dockerPullImage).Info("Pull docker image")
	tools.ExecCmd("docker", "pull", dockerPullImage)

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
	imagePush := dockerImageTag(dockerPushImage, project.ReleaseVersion())
	log.WithFields(log.Fields{"build": imagePull, "release": imagePush}).Info("Retag docker image")
	tools.ExecCmd("docker", "tag", imagePull, imagePush)

	if flags.Docker.Project.SkipPush {
		log.WithField("image", imagePush).Info("Skip docker push for docker release image")
	} else {
		tools.ExecCmd("docker", "push", imagePush)
		log.WithField("image", imagePush).Info("Release docker image done!")
	}
}
