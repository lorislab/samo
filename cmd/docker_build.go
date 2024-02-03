package cmd

import (
	"github.com/lorislab/samo/log"
	"github.com/lorislab/samo/tools"
	"github.com/spf13/cobra"
)

type dockerBuildFlags struct {
	Docker          dockerFlags `mapstructure:",squash"`
	File            string      `mapstructure:"docker-file"`
	Profile         string      `mapstructure:"docker-profile"`
	Context         string      `mapstructure:"docker-context"`
	Platform        string      `mapstructure:"docker-platform"`
	Provenance      string      `mapstructure:"docker-provenance"`
	BuildX          bool        `mapstructure:"docker-buildx"`
	SkipDevBuild    bool        `mapstructure:"docker-skip-dev"`
	SkipPull        bool        `mapstructure:"docker-skip-pull"`
	BuildPush       bool        `mapstructure:"docker-build-push"`
	SkipRemoveBuild bool        `mapstructure:"docker-remove-build-skip"`
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

	addStringFlag(cmd, "docker-file", "d", "src/main/docker/Dockerfile", "path of the project Dockerfile")
	addStringFlag(cmd, "docker-profile", "", "", "profile of the Dockerfile.<profile>")
	addStringFlag(cmd, "docker-context", "", ".", "the docker build context")
	addStringFlag(cmd, "docker-platform", "", "", "the docker build platform")
	addStringFlag(cmd, "docker-provenance", "", "", "the provenance attestations include facts about the build process, including details")
	addBoolFlag(cmd, "docker-skip-pull", "", false, "skip docker pull new images for the build")
	addBoolFlag(cmd, "docker-build-push", "", false, "push docker image after build")
	addBoolFlag(cmd, "docker-buildx", "", false, "extended build capabilities with BuildKit")
	addBoolFlag(cmd, "docker-skip-dev", "", false, "skip build image {{ .Name }}:latest")
	addBoolFlag(cmd, "docker-remove-intermediate-img-skip", "", false, "skip remove build intermediate containers")

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
		log.Fatal("Dockerfile does not exists!", log.F("file", dockerfile))
	}

	dockerImage := dockerImage(project, flags.Docker.Registry, flags.Docker.Group, flags.Docker.Repo)
	tags := dockerTags(dockerImage, project, flags.Docker.TagListTemplate)

	if !flags.SkipDevBuild {
		tags = append(tags, project.Name()+":latest")
	}

	log.Info("Build docker image", log.Fields{"image": dockerImage, "tags": tags})

	var command []string
	if flags.BuildX {
		command = append(command, "buildx")
	}
	command = append(command, "build")
	if !flags.SkipPull {
		command = append(command, "--pull")
	}

	// Removing intermediate container
	if !flags.SkipRemoveBuild {
		command = append(command, "--rm")
	}
	if len(flags.Provenance) > 0 {
		command = append(command, "--provenance", flags.Provenance)
	}
	if len(flags.Platform) > 0 {
		command = append(command, "--platform", flags.Platform)
	}
	// create labels
	labels := dockerLabels(project, flags.Docker.Project.SkipLabels, flags.Docker.SkipOpenContainersLabels, flags.Docker.Project.LabelTemplate)
	for key, value := range labels {
		command = append(command, "--label", key+"="+value)
	}

	// add tags
	for _, tag := range tags {
		command = append(command, "-t", tag)
	}

	// add dockerfile and dockerfile profile
	command = append(command, "-f", dockerfile)

	// push images for buildx
	if flags.BuildX && flags.BuildPush {
		command = append(command, "--push")
	}

	// set docker context
	command = append(command, flags.Context)
	// execute command
	tools.ExecCmd("docker", command...)

	log.Info("Docker build done!", log.Fields{"image": dockerImage, "tags": tags})

	// for none buildx we need to push it manually
	if !flags.BuildX && flags.BuildPush {
		dockerImagePush(dockerImage, tags, flags.Docker.Project.SkipPush)
	}
}
