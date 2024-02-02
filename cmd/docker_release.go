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
	ReleaseTags     string      `mapstructure:"docker-release-tags"`
	ImageTools      bool        `mapstructure:"docker-release-image-tools"`
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
	addStringFlag(cmd, "docker-release-tags", "", "{{ .Release }}", "the docker release tags. Default value release version.")
	addBoolFlag(cmd, "docker-release-image-tools", "", false, "buildx imagetools create a new image based on source images")

	return cmd
}

func dockerRelease(project *Project, flags dockerReleaseFlags) {

	if project.Count() != "0" || len(project.Tag()) == 0 {
		log.Fatal("Can not created docker release. Missing tag on current commit",
			log.Fields{"version": project.Version(), "hash": project.Hash(), "count": project.Count(), "tag": project.Tag()})
	}

	// switch back to rc version
	project.switchBackToReleaseCandidate()
	log.Info("Create docker release", log.Fields{"version": project.Version(), "release": project.Release()})

	dockerPullImage := dockerImage(project, flags.Docker.Registry, flags.Docker.Group, flags.Docker.Repo)
	imagePull := dockerImageTag(dockerPullImage, project.Version())

	log.Info("Docker image", log.F("image", imagePull))

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
	dockerPushImageTags := dockerTags(dockerPushImage, project, flags.ReleaseTags)

	if flags.ImageTools {
		dockerReleaseImageTools(flags.Docker.Project.SkipPush, imagePull, dockerPushImageTags)
	} else {
		dockerReleasePullPush(flags.Docker.Project.SkipPush, imagePull, dockerPushImage, dockerPushImageTags)
	}
}

func dockerReleaseImageTools(skip bool, imagePull string, dockerPushImageTags []string) {

	var command []string

	// buildx imagetools create
	command = append(command, "buildx", "imagetools", "create")

	// dry run
	if skip {
		command = append(command, "--dry-run")
	}

	// add new tags
	for _, imagePush := range dockerPushImageTags {
		command = append(command, "-t", imagePush)
	}

	// add remote repository image
	command = append(command, imagePull)

	// execute command
	tools.ExecCmd("docker", command...)
}

// deprecated
func dockerReleasePullPush(skip bool, imagePull string, dockerPushImage string, dockerPushImageTags []string) {

	// pull docker image
	tools.ExecCmd("docker", "pull", imagePull)

	for _, imagePush := range dockerPushImageTags {
		log.Info("Re-tag docker image", log.Fields{"build": imagePull, "release": imagePush})
		tools.ExecCmd("docker", "tag", imagePull, imagePush)
	}

	if skip {
		log.Info("Skip docker push for docker release image", log.Fields{"image": dockerPushImage, "tags": dockerPushImageTags})
	} else {
		dockerImagePush(dockerPushImage, dockerPushImageTags, skip)
		log.Info("Release docker image done!", log.F("image", dockerPushImage))
	}
}
