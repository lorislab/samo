package cmd

import (
	"strings"

	"github.com/lorislab/samo/tools"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type dockerBuildFlags struct {
	Docker          dockerFlags `mapstructure:",squash"`
	File            string      `mapstructure:"file"`
	Profile         string      `mapstructure:"profile"`
	Context         string      `mapstructure:"context"`
	SkipPull        bool        `mapstructure:"pull-skip"`
	BuildPush       bool        `mapstructure:"build-push"`
	SkipRemoveBuild bool        `mapstructure:"remove-build-skip"`
}

func createDockerBuildCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "build",
		Short: "Build the docker image of the project",
		Long:  `Build the docker image of the project`,
		Run: func(cmd *cobra.Command, args []string) {

			flags := dockerBuildFlags{}
			readOptions(&flags)
			project := loadProject(flags.Docker.Project)
			dockerBuild(project, flags)
		},
		TraverseChildren: true,
	}

	addStringFlag(cmd, "file", "d", "src/main/docker/Dockerfile", "path of the project Dockerfile")
	addStringFlag(cmd, "profile", "p", "", "profile of the Dockerfile.<profile>")
	addStringFlag(cmd, "context", "", ".", "the docker build context")
	addBoolFlag(cmd, "pull-skip", "", false, "skip docker pull new images for the build")
	addBoolFlag(cmd, "build-push", "", false, "push docker image after build")
	addBoolFlag(cmd, "remove-intermediate-img-skip", "", false, "skip remove build intermediate containers")

	return cmd
}

// DockerBuild build docker image of the project
func dockerBuild(project *Project, flags dockerBuildFlags) {

	dockerfile := flags.File
	if len(dockerfile) <= 0 {
		dockerfile = "Dockerfile"
	}
	if len(flags.Profile) > 0 {
		dockerfile = dockerfile + "." + flags.Profile
	}

	if !tools.Exists(flags.File) {
		log.WithFields(log.Fields{"file": dockerfile}).Fatal("Dockerfile does not exists!")
	}

	dockerImage := dockerImage(project, flags.Docker.Registry, flags.Docker.Group, flags.Docker.Repo)
	tags := dockerTags(dockerImage, project, flags.Docker)

	log.WithFields(log.Fields{"image": flags.Docker.Registry, "tags": tags}).Info("Build docker image")

	var command []string
	command = append(command, "build")
	if !flags.SkipPull {
		command = append(command, "--pull")
	}
	// Removing intermediate container
	if !flags.SkipRemoveBuild {
		command = append(command, "--rm")
	}

	// add labels
	if !flags.Docker.Project.SkipLabels {
		command = append(command, "--label", "samo.project.hash="+project.Hash())
		command = append(command, "--label", "samo.project.version="+project.Version())
	}

	// add custom labels
	if len(flags.Docker.Project.LabelTemplate) > 0 {
		labelTemplate := tools.Template(project, flags.Docker.Project.LabelTemplate)
		labels := strings.Split(labelTemplate, ",")
		for _, label := range labels {
			command = append(command, "--label", label)
		}
	}

	// add tags
	for _, tag := range tags {
		command = append(command, "-t", tag)
	}

	// add dockerfile and dockerfile profile
	command = append(command, "-f", dockerfile)

	// set docker context
	command = append(command, flags.Context)
	// execute command
	tools.ExecCmd("docker", command...)

	log.WithFields(log.Fields{"image": dockerImage, "tags": tags}).Info("Docker build done!")

	if flags.BuildPush {
		dockerImagePush(dockerImage, tags, flags.Docker.Project.SkipPush)
	}
}
