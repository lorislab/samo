package cmd

import (
	"strings"

	"github.com/lorislab/samo/log"
	"github.com/lorislab/samo/tools"
	"github.com/spf13/cobra"
)

type dockerFlags struct {
	Project         projectFlags `mapstructure:",squash"`
	Registry        string       `mapstructure:"docker-registry"`
	Group           string       `mapstructure:"docker-group"`
	Repo            string       `mapstructure:"docker-repository"`
	TagListTemplate string       `mapstructure:"docker-tag-template-list"`
}

func createDockerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:              "docker",
		Short:            "Docker operation for the project",
		Long:             `Docker task for the project. To build, push or release docker images of the project.`,
		TraverseChildren: true,
	}

	addStringFlag(cmd, "docker-registry", "", "", "the docker registry")
	addStringFlag(cmd, "docker-group", "", "", "the docker repository group")
	addStringFlag(cmd, "docker-repository", "", "", "the docker repository. Default value is the project name.")
	addStringFlag(cmd, "docker-tag-template-list", "", "{{ .Version }}", `docker tag list template. 
	Values: `+templateValues+`
	Example: {{ .Version }},latest,{{ .Hash }}
	`)

	addChildCmd(cmd, createDockerBuildCmd())
	addChildCmd(cmd, createDockerPushCmd())
	addChildCmd(cmd, createDockerReleaseCmd())

	return cmd
}

func dockerImage(project *Project, registry, group, repository string) string {
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

func dockerTags(dockerImage string, pro *Project, flags dockerFlags) []string {
	tagTemplate := tools.Template(pro, flags.TagListTemplate)
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
	log.Info("Push docker image tags", log.Fields{"image": image, "tags": tags})
	if skip {
		log.Info("Skip docker push", log.F("image", image))
	} else {
		for _, tag := range tags {
			tools.ExecCmd("docker", "push", tag)
		}
	}
	log.Info("Push docker image done!", log.Fields{"image": image, "tags": tags})
}
