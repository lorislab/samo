package docker

import (
	"os"

	"github.com/lorislab/samo/git"
	"github.com/lorislab/samo/project"
	"github.com/lorislab/samo/tools"
	log "github.com/sirupsen/logrus"
)

// DockerRequest the docker command request
type DockerRequest struct {
	Registry                string
	RepositoryPrefix        string
	Repository              string
	HashLength              int
	Tags                    map[string]bool
	CustomTags              []string
	Dockerfile              string
	Context                 string
	SkipPull                bool
	SkipPush                bool
	SkipPushLatest          bool
	ReleaseRegistry         string
	ReleaseRepositoryPrefix string
	ReleaseRepository       string
	BuildNumberLength       int
	BuildNumberPrefix       string
}

func (request DockerRequest) dockerProjectRepositoryImage(p project.Project) string {
	return dockerProjectRepositoryImage(request.Registry, request.RepositoryPrefix, request.Repository, p.Name())
}

func dockerProjectRepositoryImage(registry, repositoryPrefix, repository, name string) string {
	tmp := dockerProjectImage(repository, name)
	if len(repositoryPrefix) > 0 {
		tmp = repositoryPrefix + tmp
	}
	if len(registry) > 0 {
		tmp = registry + "/" + tmp
	}
	return tmp
}

func (request DockerRequest) dockerProjectImage(p project.Project) string {
	return dockerProjectImage(request.Repository, p.Name())
}

func dockerProjectImage(repository, name string) string {
	if len(repository) == 0 {
		return name
	}
	return repository
}

// DockerRelease docker release existing docker image
func (request DockerRequest) DockerRelease(p project.Project) string {

	buildVersion := project.BuildVersion(p, request.HashLength, request.BuildNumberLength, request.BuildNumberPrefix)
	imagePull := dockerImageTag(request.dockerProjectRepositoryImage(p), buildVersion.String())
	log.WithField("image", imagePull).Info("Pull docker image")
	tools.ExecCmd("docker", "pull", imagePull)

	// check the release configuration

	releaseRegistry := request.ReleaseRegistry
	if len(releaseRegistry) == 0 {
		releaseRegistry = request.Registry
	}
	releaseRepositoryPrefix := request.ReleaseRepositoryPrefix
	if len(releaseRepositoryPrefix) == 0 {
		releaseRepositoryPrefix = request.RepositoryPrefix
	}
	releaseRepository := request.ReleaseRepository
	if len(releaseRepository) == 0 {
		releaseRepository = request.Repository
	}

	// release docker registry
	releaseVersion := project.ReleaseVersion(p)
	dockerReleaseImageRegistry := dockerProjectRepositoryImage(releaseRegistry, releaseRepositoryPrefix, releaseRepository, p.Name())
	imageRelease := dockerImageTag(dockerReleaseImageRegistry, releaseVersion.String())
	log.WithFields(log.Fields{
		"build":   imagePull,
		"release": imageRelease,
	}).Info("Retag docker image")
	tools.ExecCmd("docker", "tag", imagePull, imageRelease)

	if !request.SkipPush {
		tools.ExecCmd("docker", "push", imageRelease)
	} else {
		log.Info("Skip docker push for docker release image: " + imageRelease)
	}
	return imageRelease
}

// DockerBuild build docker image of the project
func (request DockerRequest) DockerBuild(p project.Project) (string, []string) {

	if _, err := os.Stat(request.Dockerfile); os.IsNotExist(err) {
		log.WithField("Dockerfile", request.Dockerfile).Fatal("Dockerfile does not exists!")
	}

	dockerImage, tags := request.dockerTags(p)
	request.dockerBuild(tags)
	return dockerImage, tags
}

// DockerBuildDev build development version
func (request DockerRequest) DockerBuildDev(p project.Project) (string, []string) {

	if _, err := os.Stat(request.Dockerfile); os.IsNotExist(err) {
		log.WithField("Dockerfile", request.Dockerfile).Fatal("Dockerfile does not exists!")
	}

	dockerImage := request.dockerProjectImage(p)
	tags := []string{dockerImageTag(dockerImage, p.Version()), dockerImageTag(dockerImage, "latest")}
	request.dockerBuild(tags)
	return dockerImage, tags
}

func (request DockerRequest) dockerBuild(tags []string) {
	log.WithFields(log.Fields{
		"image": request.Registry,
		"tags":  tags,
	}).Info("Build docker image")

	var command []string
	command = append(command, "build")
	if !request.SkipPull {
		command = append(command, "--pull")
	}
	// add tags
	for _, tag := range tags {
		command = append(command, "-t", tag)
	}
	// add dockerfile
	if len(request.Dockerfile) > 0 {
		command = append(command, "-f", request.Dockerfile)
	}
	// set docker context
	command = append(command, request.Context)
	// execute command
	tools.ExecCmd("docker", command...)
}

// DockerPush push docker image of the project
func (request DockerRequest) DockerPush(p project.Project) (string, []string) {

	dockerImage, tags := request.dockerTags(p)

	log.WithFields(log.Fields{
		"image": dockerImage,
		"tags":  tags,
	}).Info("Push docker image tags")
	if !request.SkipPush {
		for _, tag := range tags {
			tools.ExecCmd("docker", "push", tag)
		}
	} else {
		log.WithField("image", dockerImage).Info("Skip docker push")
	}
	return dockerImage, tags
}

func (request DockerRequest) dockerTags(p project.Project) (string, []string) {
	dockerImage := request.dockerProjectRepositoryImage(p)
	var tags []string
	// project version tag
	if request.Tags["version"] {
		tags = append(tags, dockerImageTag(dockerImage, p.Version()))
	}
	// project build-version tag
	if request.Tags["build-version"] {
		buildVersion := project.BuildVersion(p, request.HashLength, request.BuildNumberLength, request.BuildNumberPrefix)
		tags = append(tags, dockerImageTag(dockerImage, buildVersion.String()))
	}
	// latest tag
	if request.Tags["latest"] {
		tags = append(tags, dockerImageTag(dockerImage, "latest"))
	}
	// hash tag
	if request.Tags["hash"] {
		ver := project.HashVersion(p, request.HashLength)
		tags = append(tags, dockerImageTag(dockerImage, ver.String()))
	}
	// branch tag
	if request.Tags["branch"] {
		branch := git.GitBranch()
		tags = append(tags, dockerImageTag(dockerImage, branch))
	}
	// developer latest tag
	if request.Tags["dev"] {
		tmp := request.dockerProjectImage(p)
		tags = append(tags, dockerImageTag(tmp, "latest"))
	}
	// custom tags
	if len(request.CustomTags) > 0 {
		for _, tag := range request.CustomTags {
			tags = append(tags, dockerImageTag(dockerImage, tag))
		}
	}
	return dockerImage, tags
}

func dockerImageTag(name, tag string) string {
	return name + ":" + tag
}
