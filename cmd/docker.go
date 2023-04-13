package cmd

import (
	"regexp"
	"strings"
	"time"

	"github.com/lorislab/samo/log"
	"github.com/lorislab/samo/tools"
	"github.com/spf13/cobra"
)

type dockerFlags struct {
	Project                  projectFlags `mapstructure:",squash"`
	Registry                 string       `mapstructure:"docker-registry"`
	Group                    string       `mapstructure:"docker-group"`
	Repo                     string       `mapstructure:"docker-repository"`
	TagListTemplate          string       `mapstructure:"docker-tag-template-list"`
	SkipOpenContainersLabels bool         `mapstructure:"docker-skip-opencontainers-labels"`
}

func createDockerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:              "docker",
		Short:            "Docker operation for the project",
		Long:             `Docker task for the project. To build, push or release docker images of the project.`,
		TraverseChildren: true,
	}

	addBoolFlag(cmd, "docker-skip-open-containers-labels", "", false, "skip open containers labels ")
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
	addChildCmd(cmd, createDockerTagsCmd())
	addChildCmd(cmd, createDockerLabelsCmd())

	return cmd
}

func dockerImage(project *Project, registry, group, repository string) string {
	dockerImage := repository
	if len(dockerImage) == 0 {
		dockerImage = project.Name()
	}
	if len(group) > 0 {
		if !strings.HasSuffix(group, `/`) {
			group = group + "/"
		}
		dockerImage = group + dockerImage
	}
	if len(registry) > 0 {
		dockerImage = registry + "/" + dockerImage
	}
	return dockerImage
}

func dockerTags(dockerImage string, pro *Project, dockerTags string) []string {
	tagTemplate := tools.Template(pro, dockerTags)
	items := strings.Split(tagTemplate, ",")

	var tags []string
	for _, tag := range items {
		var t = dockerReplaceTag(tag)
		if len(dockerImage) > 0 {
			tags = append(tags, dockerImageTag(dockerImage, t))
		} else {
			tags = append(tags, t)
		}
	}
	return tags
}

// A tag name must be valid ASCII and may contain lowercase and uppercase letters, digits, underscores,
// periods and hyphens. A tag name may not start with a period or a hyphen and may contain a maximum
// of 128 characters.
// [a-z][A-Z][0-9]_.-
var dockerTagRegex = regexp.MustCompile(`[^a-zA-Z0-9_.-]+`)

func dockerReplaceTag(tag string) string {
	return dockerTagRegex.ReplaceAllString(tag, "_")
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

func dockerLabels(project *Project, skipLabels bool, skipOpenContainersLabels bool, customLabels string) map[string]string {

	result := map[string]string{}

	created := time.Now().Format(time.RFC3339)

	// add labels
	if !skipLabels {
		result["samo.project.revision"] = project.Hash()
		result["samo.project.version"] = project.Version()
		result["samo.project.created"] = created
	}

	// add open-containers labels
	if !skipOpenContainersLabels {
		result["org.opencontainers.image.created"] = created
		result["org.opencontainers.image.title"] = project.Name()
		result["org.opencontainers.image.revision"] = project.Hash()
		result["org.opencontainers.image.version"] = project.Version()
		result["org.opencontainers.image.source"] = project.Source()
	}

	// add custom labels
	if len(customLabels) > 0 {
		labelTemplate := tools.Template(project, customLabels)
		labels := strings.Split(labelTemplate, ",")
		for _, label := range labels {
			kv := strings.Split(label, "=")
			if len(kv) > 1 {
				result[kv[0]] = kv[1]
			}
		}
	}

	return result
}
