package docker

import (
	"os"

	"github.com/lorislab/samo/project"
	"github.com/lorislab/samo/tools"
	log "github.com/sirupsen/logrus"
)

// DockerRequest the docker command request
type DockerRequest struct {
	Project                 project.Project
	Versions                project.Versions
	PushVersions            project.Versions
	Registry                string
	RepositoryPrefix        string
	Repository              string
	Dockerfile              string
	Context                 string
	SkipPull                bool
	SkipPush                bool
	ReleaseRegistry         string
	ReleaseRepositoryPrefix string
	ReleaseRepository       string
}

func (request DockerRequest) dockerProjectRepositoryImage() string {
	return dockerProjectRepositoryImage(request.Registry, request.RepositoryPrefix, request.Repository, request.Project.Name())
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

func (request DockerRequest) dockerProjectImage() string {
	return dockerProjectImage(request.Repository, request.Project.Name())
}

func dockerProjectImage(repository, name string) string {
	if len(repository) == 0 {
		return name
	}
	return repository
}

// DockerRelease docker release existing docker image
func (request DockerRequest) DockerRelease() string {

	imagePull := dockerImageTag(request.dockerProjectRepositoryImage(), request.Versions.BuildVersion())
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
	dockerReleaseImageRegistry := dockerProjectRepositoryImage(releaseRegistry, releaseRepositoryPrefix, releaseRepository, request.Project.Name())
	imageRelease := dockerImageTag(dockerReleaseImageRegistry, request.Versions.ReleaseVersion())
	log.WithFields(log.Fields{
		"build":   imagePull,
		"release": imageRelease,
	}).Info("Retag docker image")
	tools.ExecCmd("docker", "tag", imagePull, imageRelease)

	if !request.SkipPush {
		tools.ExecCmd("docker", "push", imageRelease)
	} else {
		log.WithField("image", imageRelease).Info("Skip docker push for docker release image")
	}
	return imageRelease
}

// DockerBuild build docker image of the project
func (request DockerRequest) DockerBuild() (string, []string) {

	if _, err := os.Stat(request.Dockerfile); os.IsNotExist(err) {
		log.WithField("Dockerfile", request.Dockerfile).Fatal("Dockerfile does not exists!")
	}

	dockerImage, tags := request.dockerTags()
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
func (request DockerRequest) DockerPush() (string, []string) {

	dockerImage, tags := request.dockerTags()

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

func (request DockerRequest) dockerTags() (string, []string) {
	dockerImage := request.dockerProjectRepositoryImage()
	var tags []string
	// project version tag
	if request.Versions.IsVersion() {
		tags = append(tags, dockerImageTag(dockerImage, request.Versions.Version()))
	}
	// project build-version tag
	if request.Versions.IsBuildVersion() {
		tags = append(tags, dockerImageTag(dockerImage, request.Versions.BuildVersion()))
	}
	// latest tag
	if request.Versions.IsLatestVersion() {
		tags = append(tags, dockerImageTag(dockerImage, request.Versions.LatestVersion()))
	}
	// hash tag
	if request.Versions.IsHashVersion() {
		tags = append(tags, dockerImageTag(dockerImage, request.Versions.HashVersion()))
	}
	// branch tag
	if request.Versions.IsBranchVersion() {
		tags = append(tags, dockerImageTag(dockerImage, request.Versions.BranchVersion()))
	}
	// developer latest tag
	if request.Versions.IsDevVersion() {
		tmp := request.dockerProjectImage()
		tags = append(tags, dockerImageTag(tmp, request.Versions.DevVersion()))
	}
	// custom tags
	if request.Versions.IsCustom() {
		for _, tag := range request.Versions.Custom() {
			tags = append(tags, dockerImageTag(dockerImage, tag))
		}
	}
	return dockerImage, tags
}

func dockerImageTag(name, tag string) string {
	return name + ":" + tag
}
