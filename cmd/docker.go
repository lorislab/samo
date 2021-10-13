package cmd

import (
	"strings"

	"github.com/lorislab/samo/project"
	"github.com/lorislab/samo/tools"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type dockerFlags struct {
	Project         projectFlags `mapstructure:",squash"`
	Registry        string       `mapstructure:"registry"`
	Group           string       `mapstructure:"group"`
	Repo            string       `mapstructure:"repository"`
	TagListTemplate string       `mapstructure:"tag-list-template"`
}

func createDockerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:              "docker",
		Short:            "Docker operation for the project",
		Long:             `Docker task for the project. To build, push or release docker images of the project.`,
		TraverseChildren: true,
	}

	addStringFlag(cmd, "registry", "", "", "the docker registry")
	addStringFlag(cmd, "group", "", "", "the docker repository group")
	addStringFlag(cmd, "repository", "", "", "the docker repository. Default value is the project name.")
	addStringFlag(cmd, "tag-list-template", "", "{{ .Project.Version }}", `docker tag list template. 
	Values: Project.Hash,Project.Branch,Project.Tag,Project.Count,Project.Version,Project.Release. 
	Example: {{ .Project.Version }},latest,{{ .Project.Hash }}`)

	addChildCmd(cmd, createDockerBuildCmd())
	addChildCmd(cmd, createDockerPushCmd())
	addChildCmd(cmd, createDockerReleaseCmd())

	return cmd
}

func dockerImage(project *project.Project, registry, group, repository string) string {
	dockerImage := repository
	if len(dockerImage) == 0 {
		dockerImage = project.Name()
	}
	if len(group) > 0 {
		dockerImage = group + dockerImage
	}
	if len(registry) > 0 {
		dockerImage = registry + "/" + dockerImage
	}
	return dockerImage
}

func dockerTags(dockerImage string, pro *project.Project, flags dockerFlags) []string {

	data := struct {
		Project *project.Project
	}{
		Project: pro,
	}
	tagTemplate := tools.Template(data, flags.TagListTemplate)
	items := strings.Split(tagTemplate, ",")

	var tags []string
	for _, tag := range items {
		tags = append(tags, dockerImageTag(dockerImage, tag))
	}
	return tags
}

func dockerImageTag(dockerImage, tag string) string {
	return dockerImage + ":" + tag
}

func dockerImagePush(image string, tags []string, skip bool) {
	log.WithFields(log.Fields{"image": image, "tags": tags}).Info("Push docker image tags")
	if skip {
		log.WithField("image", image).Info("Skip docker push")
	} else {
		for _, tag := range tags {
			tools.ExecCmd("docker", "push", tag)
		}
	}
	log.WithFields(log.Fields{"image": image, "tags": tags}).Info("Push docker image done!")
}
