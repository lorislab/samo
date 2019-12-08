package cmd

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/lorislab/samo/internal"
	"github.com/spf13/cobra"
)

func init() {

	dockerCmd.AddCommand(dockerBuildCmd)
	addMavenFlags(dockerBuildCmd)
	addDockerImageFlags(dockerBuildCmd)
	dockerBuildCmd.Flags().StringVarP(&(dockerOptions.dockerfile), "dockerfile", "d", "src/main/docker/Dockerfile", "The maven project dockerfile")
	dockerBuildCmd.Flags().StringVarP(&(dockerOptions.repository), "repository", "r", "", "The docker repository")
	dockerBuildCmd.Flags().StringVarP(&(dockerOptions.context), "context", "c", ".", "The docker build context")
	dockerBuildCmd.Flags().StringVarP(&(dockerOptions.tag), "tag", "t", "", "Add the extra tag to the build image")
	dockerBuildCmd.Flags().BoolVarP(&(dockerOptions.branch), "branch", "b", true, "Tag the docker image with a branch name")
	dockerBuildCmd.Flags().BoolVarP(&(dockerOptions.latest), "latest", "l", true, "Tag the docker image with a latest")

	dockerCmd.AddCommand(dockerPushCmd)
	addMavenFlags(dockerPushCmd)

	dockerCmd.AddCommand(dockerReleaseCmd)
	addMavenFlags(dockerReleaseCmd)
	addGitFlags(dockerReleaseCmd)
	addDockerImageFlags(dockerReleaseCmd)

	dockerCmd.AddCommand(dockerConfigCmd)
	dockerConfigCmd.Flags().StringVarP(&(dockerOptions.config), "variable", "v", "SAMO_DOCKER_CONFIG", "The docker env variable")
	dockerConfigCmd.Flags().StringVarP(&(dockerOptions.configFile), "config-file", "j", "~/.docker/config.json", "Docker client configuration client")
}

type dockerFlags struct {
	hashLength string
	dockerfile string
	context    string
	branch     bool
	latest     bool
	tag        string
	repository string
	image      string
	config     string
	configFile string
}

func addDockerImageFlags(command *cobra.Command) {
	command.Flags().StringVarP(&(dockerOptions.image), "image", "i", "", "the docker image. Default value maven project artifactId.")
}

var (
	dockerOptions = dockerFlags{}

	dockerCmd = &cobra.Command{
		Use:              "docker",
		Short:            "Docker operation",
		Long:             `Docker operation`,
		TraverseChildren: true,
	}
	dockerBuildCmd = &cobra.Command{
		Use:   "build",
		Short: "Build the docker image",
		Long:  `Build the docker image`,
		Run: func(cmd *cobra.Command, args []string) {
			project := internal.LoadMavenProject(mavenOptions.filename)
			image := dockerOptions.image
			if len(image) == 0 {
				image = project.ArtifactID()
			}

			var command []string

			if len(dockerOptions.repository) > 0 {
				image = dockerOptions.repository + "/" + image
			}

			command = append(command, "build", "-t", imageNameWithTag(image, project.Version()))

			if dockerOptions.branch {
				branch := internal.GitBranch()
				command = append(command, "-t", imageNameWithTag(image, branch))
			}
			if dockerOptions.latest {
				command = append(command, "-t", imageNameWithTag(image, "latest"))
			}
			if len(dockerOptions.tag) > 0 {
				command = append(command, "-t", imageNameWithTag(image, dockerOptions.tag))
			}
			if len(dockerOptions.dockerfile) > 0 {
				command = append(command, "-f", dockerOptions.dockerfile)
			}
			command = append(command, dockerOptions.context)

			err := exec.Command("docker", command...).Run()
			if err != nil {
				panic(err)
			}
		},
		TraverseChildren: true,
	}
	dockerPushCmd = &cobra.Command{
		Use:   "push",
		Short: "Push the docker image",
		Long:  `Push the docker image`,
		Run: func(cmd *cobra.Command, args []string) {
			project := internal.LoadMavenProject(mavenOptions.filename)
			image := dockerOptions.image
			if len(image) == 0 {
				image = project.ArtifactID()
			}

			err := exec.Command("docker", "push", image).Run()
			if err != nil {
				panic(err)
			}

		},
		TraverseChildren: true,
	}
	dockerReleaseCmd = &cobra.Command{
		Use:   "release",
		Short: "Release the docker image",
		Long:  `Release the docker image`,
		Run: func(cmd *cobra.Command, args []string) {
			project := internal.LoadMavenProject(mavenOptions.filename)
			image := dockerOptions.image
			if len(image) == 0 {
				image = project.ArtifactID()
			}
			hash := internal.GitHash(dockerOptions.hashLength)
			pullVersion := project.SetPrerelease(hash)
			releaseVersion := project.ReleaseVersion()

			imagePull := imageNameWithTag(image, pullVersion)
			err := exec.Command("docker", "pull", imagePull).Run()
			if err != nil {
				panic(err)
			}
			imageRelease := imageNameWithTag(image, releaseVersion)
			err = exec.Command("docker", "tag", imagePull, imageRelease).Run()
			if err != nil {
				panic(err)
			}
			err = exec.Command("docker", "push", imageRelease).Run()
			if err != nil {
				panic(err)
			}

		},
		TraverseChildren: true,
	}
	dockerConfigCmd = &cobra.Command{
		Use:   "config",
		Short: "Config the docker client",
		Long:  `Config the docker client`,
		Run: func(cmd *cobra.Command, args []string) {

			value, exists := os.LookupEnv(dockerOptions.config)
			if exists && len(value) > 0 {

				dir := filepath.Dir(dockerOptions.configFile)
				err := os.MkdirAll(dir, os.ModeDir)
				if err != nil {
					panic(err)
				}

				file, err := os.Create(dockerOptions.configFile)
				if err != nil {
					panic(err)
				}
				w := bufio.NewWriter(file)
				w.WriteString(value)
				w.Flush()
			}
		},
		TraverseChildren: true,
	}
)

func dockerImage(project *internal.MavenProject, options dockerFlags) string {
	if len(options.image) == 0 {
		return project.ArtifactID()
	}
	return options.image
}

func imageName(options dockerFlags) string {
	if len(options.repository) > 0 {
		return options.repository
	}
	return options.repository + "/" + options.image
}

func imageNameWithTag(name, tag string) string {
	return name + ":" + tag
}
