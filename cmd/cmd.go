package cmd

import (
	"fmt"
	"github.com/lorislab/samo/internal"
	"github.com/spf13/pflag"
	"os"
	"strconv"
	"strings"

	"github.com/Masterminds/semver"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

// Main the commands method
func Main(rootCmd *cobra.Command) {
	rootCmd.AddCommand(mvnCmd)
	rootCmd.AddCommand(gitCmd)
	rootCmd.AddCommand(dockerCmd)
	rootCmd.AddCommand(npmCmd)
}

func addFlagRequired(command *cobra.Command, name, shorthand string, value string, usage string) *pflag.Flag {
	f := addFlag(command, name, shorthand, value, usage)
	err := command.MarkFlagRequired(name)
	if err != nil {
		log.Panic(err)
	}
	return f
}

func addFlagRef(command *cobra.Command, flag *pflag.Flag) {
	command.Flags().AddFlag(flag)
}

func addFlag(command *cobra.Command, name, shorthand string, value string, usage string) *pflag.Flag {
	command.Flags().StringP(name, shorthand, value, usage)
	return addViper(command, name)
}

func addIntFlag(command *cobra.Command, name, shorthand string, value int, usage string) *pflag.Flag {
	command.Flags().IntP(name, shorthand, value, usage)
	return addViper(command, name)
}

func addBoolFlag(command *cobra.Command, name, shorthand string, value bool, usage string) *pflag.Flag {
	command.Flags().BoolP(name, shorthand, value, usage)
	return addViper(command, name)
}

func addGitHashLength(command *cobra.Command, name, shorthand string) *pflag.Flag {
	return addIntFlag(command, name, shorthand, 12, "the git hash length")
}

func addViper(command *cobra.Command, name string) *pflag.Flag {
	f := command.Flags().Lookup(name)
	err := viper.BindPFlag(name, f)
	if err != nil {
		panic(err)
	}
	return f
}

// Commands execution

func projectVersion(project internal.Project) {
	fmt.Printf("%s\n", project.Version())
}

func projectSetBuildVersion(project internal.Project, hashLength, length int, prefix string) {
	buildVersion := buildVersion(project, hashLength, length, prefix)
	version := project.Version()
	project.SetVersion(buildVersion)
	log.Infof("Set project '%s' build version from [%s] to [%s]\n", project.Filename(), version, buildVersion)
}

func projectBuildVersion(project internal.Project, hashLength, length int, prefix string) {
	buildVersion := buildVersion(project, hashLength, length, prefix)
	fmt.Printf("%s\n", buildVersion)
}

func buildVersion(project internal.Project, hashLength, length int, prefix string) string {
	releaseVersion := project.ReleaseSemVersion().String()
	_, count, hash := internal.GitCommit(hashLength)
	tmp := internal.CreateVersion(releaseVersion)
	ver := internal.AddBuildInfo(*tmp, count, hash, prefix, length)
	return ver.String()
}

func projectReleaseVersion(project internal.Project) {
	releaseVersion := project.ReleaseSemVersion().String()
	fmt.Printf("%s\n", releaseVersion)
}

func projectSetReleaseVersion(project internal.Project) {
	releaseVersion := project.ReleaseSemVersion().String()
	version := project.Version()
	project.SetVersion(releaseVersion)
	log.Infof("Set project '%s' build version from [%s] to [%s]\n", project.Filename(), version, releaseVersion)
}

func projectCreateRelease(project internal.Project, commitMessage, tagMessage string, major, skipPush bool) {
	releaseVersion := project.ReleaseSemVersion()
	tag := releaseVersion.String()
	if len(tagMessage) == 0 {
		tagMessage = tag
	}
	internal.ExecGitCmd("git", "tag", "-a", tag, "-m", tagMessage)

	// update project file with next version
	if len(project.Filename()) > 0 {
		ver := internal.AddPrerelease(internal.NextReleaseVersion(releaseVersion, major), project.NextReleaseSuffix())
		devVersion := ver.String()
		project.SetVersion(devVersion)
		internal.ExecGitCmd("git", "add", ".")
		internal.ExecGitCmd("git", "commit", "-m", commitMessage+" ["+devVersion+"]")
	}

	if !skipPush {
		//internal.ExecGitCmd("git", "push", "origin", "refs/heads/*:refs/heads/*", "refs/tags/*:refs/tags/*")
		internal.ExecGitCmd("git", "push")
		internal.ExecGitCmd("git", "push", "--tags")
	} else {
		log.Info("Skip git push for project release version: " + tag)
	}
	log.Infof("New release [%s] created.", tag)
}

func projectCreatePatch(project internal.Project, commitMessage, patchTag, branchPrefix string, skipPush bool) {
	tagVer, e := semver.NewVersion(patchTag)
	if e != nil {
		log.Errorf("The patch tag is not valid version. Value: " + patchTag)
		log.Panic(e)
	}
	if tagVer.Patch() != 0 || len(tagVer.Prerelease()) > 0 {
		log.Errorf("Can not created patch branch from the patch version  [%s]!", tagVer.Original())
		os.Exit(0)
	}

	branchName := branchPrefix + strconv.FormatInt(tagVer.Major(), 10) + "." + strconv.FormatInt(tagVer.Minor(), 10)
	internal.ExecGitCmd("git", "checkout", "-b", branchName, patchTag)
	log.Debugf("Branch  '%s' created", branchName)

	// update project file with next version
	if len(project.Filename()) > 0 {
		// remove the prerelease
		ver := tagVer.IncPatch()
		patchVersion := ver.String()
		project.SetVersion(patchVersion)

		internal.ExecGitCmd("git", "add", ".")
		internal.ExecGitCmd("git", "commit", "-m", commitMessage+" ["+patchVersion+"]")
	}

	if !skipPush {
		//internal.ExecGitCmd("git", "push", "origin", "refs/heads/*:refs/heads/*")
		if len(project.Filename()) > 0 {
			internal.ExecGitCmd("git", "push")
		}
		internal.ExecGitCmd("git", "push", "-u", "origin", branchName)
	} else {
		log.Info("Skip git push for project patch version: " + branchName)
	}
}

func dockerProjectImage(project internal.Project, repository string) string {
	if len(repository) == 0 {
		return project.Name()
	}
	return repository
}

func dockerProjectRepositoryImage(project internal.Project, registry, repositoryPrefix, repository string) string {
	tmp := dockerProjectImage(project, repository)
	if len(repositoryPrefix) > 0 {
		tmp = repositoryPrefix + tmp
	}
	if len(registry) > 0 {
		tmp = registry + "/" + tmp
	}
	return tmp
}

func dockerImageTag(name, tag string) string {
	return name + ":" + tag
}

func projectDockerBuild(project internal.Project, registry, repositoryPrefix, repository string, hashLength int,
	branch, latest, devTag bool, buildTag, dockerfile, context string, skipPull bool) {
	dockerImage := dockerProjectRepositoryImage(project, registry, repositoryPrefix, repository)
	ver := project.ReleaseSemVersion()

	pre := internal.UpdatePrereleaseToHashVersion(ver.Prerelease(), hashLength)
	gitHashVer := internal.SetPrerelease(*ver, pre)

	var command []string
	command = append(command, "build")
	if !skipPull {
		command = append(command, "--pull")
	}
	command = append(command, "-t", dockerImageTag(dockerImage, project.Version()))
	command = append(command, "-t", dockerImageTag(dockerImage, gitHashVer.String()))
	if branch {
		branch := gitBranch()
		command = append(command, "-t", dockerImageTag(dockerImage, branch))
	}
	if latest {
		command = append(command, "-t", dockerImageTag(dockerImage, "latest"))
	}
	if devTag {
		tmp := dockerProjectImage(project, repository)
		command = append(command, "-t", dockerImageTag(tmp, "latest"))
	}
	if len(buildTag) > 0 {
		command = append(command, "-t", dockerImageTag(dockerImage, buildTag))
	}
	if len(dockerfile) > 0 {
		command = append(command, "-f", dockerfile)
	}
	command = append(command, context)

	internal.ExecCmd("docker", command...)
}

func projectDockerBuildDev(project internal.Project, image, dockerfile, context string, skipPull bool) {
	dockerImage := dockerProjectImage(project, image)

	var command []string
	command = append(command, "build")
	if !skipPull {
		command = append(command, "--pull")
	}
	command = append(command, "-t", dockerImageTag(dockerImage, project.Version()))
	command = append(command, "-t", dockerImageTag(dockerImage, "latest"))

	if len(dockerfile) > 0 {
		command = append(command, "-f", dockerfile)
	}
	command = append(command, context)

	internal.ExecCmd("docker", command...)
}

func projectDockerPush(project internal.Project, registry, repositoryPrefix, repository string, ignoreLatest, skipPush bool) {
	dockerImage := dockerProjectRepositoryImage(project, registry, repositoryPrefix, repository)
	if ignoreLatest {
		tag := dockerImageTag(dockerImage, "latest")
		output := internal.ExecCmdOutput("docker", "images", "-q", tag)
		if len(output) > 0 {
			internal.ExecCmd("docker", "rmi", tag)
		}
	}
	reference := strings.TrimPrefix(dockerImage, "docker.io/")
	tags := internal.ExecCmdOutput("docker", "images", "-f", "reference="+reference+"", "--format", "{{.Tag}}")
	log.Infof("Push docker image %s tags:\n%s\n", reference, tags)

	if !skipPush {
		internal.ExecCmd("docker", "push", dockerImage)
		log.Infof("Push docker image %s done!", reference)
	} else {
		log.Info("Skip docker push for docker image: " + dockerImage)
	}
}

func projectDockerRelease(project internal.Project, registry, repositoryPrefix, repository string, hashLength int,
	releaseRegistry, releaseRepoPrefix, releaseRepository string,
	skipPush bool) {

	// x.x.x-hash
	_, _, hash := internal.GitCommit(hashLength)
	version := project.ReleaseSemVersion()
	pullVersion := internal.AddPrerelease(*version, hash)
	imagePull := dockerImageTag(dockerProjectRepositoryImage(project, registry, repositoryPrefix, repository), pullVersion.String())
	internal.ExecCmd("docker", "pull", imagePull)

	// check the release configuration
	if len(releaseRegistry) == 0 {
		releaseRegistry = registry
	}
	if len(releaseRepoPrefix) == 0 {
		releaseRepoPrefix = repositoryPrefix
	}
	if len(releaseRepository) == 0 {
		releaseRepository = repository
	}

	// x.x.x
	releaseVersion := project.ReleaseSemVersion()
	imageRelease := dockerImageTag(dockerProjectRepositoryImage(project, releaseRegistry, releaseRepoPrefix, releaseRepository), releaseVersion.String())
	internal.ExecCmd("docker", "tag", imagePull, imageRelease)

	if !skipPush {
		internal.ExecCmd("docker", "push", imageRelease)
	} else {
		log.Info("Skip docker push for docker release image: " + imageRelease)
	}
}

//////////////////////

func isGitHub() bool {
	tmp, exists := os.LookupEnv("GITHUB_REF")
	return exists && len(tmp) > 0
}

func isGitLab() bool {
	tmp, exists := os.LookupEnv("GITLAB_CI")
	return exists && len(tmp) > 0
}
