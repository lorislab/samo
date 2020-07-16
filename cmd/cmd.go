package cmd

import (
	"bufio"
	"bytes"
	"fmt"
	"github.com/lorislab/samo/internal"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/spf13/pflag"

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
	_, count, hash := gitCommit(hashLength)
	ver := createProjectBuildVersion(releaseVersion, count, hash, prefix, length)
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
	execGitCmd("git", "tag", "-a", tag, "-m", tagMessage)

	ver := addPrerelease(nextReleaseVersion(releaseVersion, major), "SNAPSHOT")
	devVersion := ver.String()
	project.SetVersion(devVersion)

	execGitCmd("git", "add", ".")
	execGitCmd("git", "commit", "-m", commitMessage+" ["+devVersion+"]")
	if !skipPush {
		execGitCmd("git", "push", "origin", "refs/heads/*:refs/heads/*", "refs/tags/*:refs/tags/*")
	} else {
		log.Info("Skip git push for project release version: " + tag)
	}
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

	branchName := createPatchBranchName(tagVer, branchPrefix)
	execGitCmd("git", "checkout", "-b", branchName, patchTag)
	log.Debugf("Branch  '%s' created", branchName)

	// remove the prerelease
	ver := tagVer.IncPatch()
	patchVersion := ver.String()
	project.SetVersion(patchVersion)

	execGitCmd("git", "add", ".")
	execGitCmd("git", "commit", "-m", commitMessage+" ["+patchVersion+"]")
	execGitCmd("git", "push", "origin", "refs/heads/*:refs/heads/*")

	if !skipPush {
		execGitCmd("git", "push", "origin", "refs/heads/*:refs/heads/*")
	} else {
		log.Info("Skip git push for project patch version: " + patchVersion)
	}
}

func dockerProjectImage(project internal.Project, image string) string {
	if len(image) == 0 {
		return project.Name()
	}
	return image
}

func dockerProjectRepositoryImage(project internal.Project, repository, lib, image string) string {
	tmp := dockerProjectImage(project, image)
	if len(lib) > 0 {
		tmp = lib + "/" + tmp
	}
	if len(repository) > 0 {
		tmp = repository + "/" + tmp
	}
	return tmp
}

func dockerImageTag(name, tag string) string {
	return name + ":" + tag
}

func projectDockerBuild(project internal.Project, repository, lib, image string, hashLength int,
	branch, latest, devTag bool, buildTag, dockerfile, context string) {
	dockerImage := dockerProjectRepositoryImage(project, repository, lib, image)
	ver := project.ReleaseSemVersion()

	pre := updatePrereleaseToHashVersion(ver.Prerelease(), hashLength)
	gitHashVer := setPrerelease(*ver, pre)

	var command []string
	command = append(command, "build", "-t", dockerImageTag(dockerImage, project.Version()))

	command = append(command, "-t", dockerImageTag(dockerImage, gitHashVer.String()))
	if branch {
		branch := gitBranch()
		command = append(command, "-t", dockerImageTag(dockerImage, branch))
	}
	if latest {
		command = append(command, "-t", dockerImageTag(dockerImage, "latest"))
	}
	if devTag {
		tmp := dockerProjectImage(project, image)
		command = append(command, "-t", dockerImageTag(tmp, "latest"))
	}
	if len(buildTag) > 0 {
		command = append(command, "-t", dockerImageTag(dockerImage, buildTag))
	}
	if len(dockerfile) > 0 {
		command = append(command, "-f", dockerfile)
	}
	command = append(command, context)

	execCmd("docker", command...)
}

func projectDockerBuildDev(project internal.Project, image, dockerfile, context string) {
	dockerImage := dockerProjectImage(project, image)

	var command []string
	command = append(command, "build", "-t", dockerImageTag(dockerImage, project.Version()))
	command = append(command, "-t", dockerImageTag(dockerImage, "latest"))

	if len(dockerfile) > 0 {
		command = append(command, "-f", dockerfile)
	}
	command = append(command, context)

	execCmd("docker", command...)
}

func projectDockerPush(project internal.Project, repository, lib, image string, ignoreLatest, skipPush bool) {
	dockerImage := dockerProjectRepositoryImage(project, repository, lib, image)
	if ignoreLatest {
		tag := dockerImageTag(dockerImage, "latest")
		output := execCmdOutput("docker", "images", "-q", tag)
		if len(output) > 0 {
			execCmd("docker", "rmi", tag)
		}
	}
	tags := execCmdOutput("docker", "images", "-f", "'reference="+dockerImage+"'", "--format", "'{{.Repository}}: {{.Tag}}'")
	fmt.Printf("%s\n", tags)

	if !skipPush {
		execCmd("docker", "push", dockerImage)
	} else {
		log.Info("Skip docker push for docker image: " + dockerImage)
	}
}

func projectDockerRelease(project internal.Project, repository, lib, image string, hashLength int,
	releaseRepository, releaseLib, releaseImage string,
	skipPush bool) {

	// x.x.x-hash
	_, _, hash := gitCommit(hashLength)
	version := project.ReleaseSemVersion()
	pullVersion := addPrerelease(*version, hash)
	imagePull := dockerImageTag(dockerProjectRepositoryImage(project, repository, lib, image), pullVersion.String())
	execCmd("docker", "pull", imagePull)

	// check the release configuration
	if len(releaseRepository) == 0 {
		releaseRepository = repository
	}
	if len(releaseLib) == 0 {
		releaseLib = lib
	}
	if len(releaseImage) == 0 {
		releaseImage = image
	}

	// x.x.x
	releaseVersion := project.ReleaseSemVersion()
	imageRelease := dockerImageTag(dockerProjectRepositoryImage(project, releaseRepository, releaseLib, releaseImage), releaseVersion.String())
	execCmd("docker", "tag", imagePull, imageRelease)

	if !skipPush {
		execCmd("docker", "push", imageRelease)
	} else {
		log.Info("Skip docker push for docker release image: " + imageRelease)
	}
}

//////////////////////

func execCmd(name string, arg ...string) {
	log.Info(name+" ", strings.Join(arg, " "))
	cmd := exec.Command(name, arg...)

	// enable always error log for the command
	errorReader, err := cmd.StderrPipe()
	if err != nil {
		panic(err)
	}
	scannerError := bufio.NewScanner(errorReader)
	go func() {
		for scannerError.Scan() {
			log.Error(scannerError.Text())
		}
	}()

	// enable info log for the command
	if log.GetLevel() == log.DebugLevel {
		// create a pipe for the output of the script
		cmdReader, err := cmd.StdoutPipe()
		if err != nil {
			panic(err)
		}

		scanner := bufio.NewScanner(cmdReader)
		go func() {
			for scanner.Scan() {
				log.Debug(scanner.Text())
			}
		}()
	}

	err = cmd.Start()
	if err != nil {
		panic(err)
	}

	err = cmd.Wait()
	if err != nil {
		panic(err)
	}
}

func execCmdOutput(name string, arg ...string) string {
	log.Debug(name+" ", strings.Join(arg, " "))
	out, err := exec.Command(name, arg...).CombinedOutput()
	log.Debug("Output:\n", string(out))
	if err != nil {
		log.Error(string(out))
		log.Panic(err)
	}
	return string(bytes.TrimRight(out, "\n"))
}

func execCmdOutputErr(name string, arg ...string) (string, error) {
	log.Debug(name+" ", strings.Join(arg, " "))
	out, err := exec.Command(name, arg...).CombinedOutput()
	log.Debug("Output:\n", string(out))
	return string(bytes.TrimRight(out, "\n")), err
}

func execCmdErr(name string, arg ...string) error {
	log.Debug(name+" ", strings.Join(arg, " "))
	out, err := exec.Command(name, arg...).CombinedOutput()
	log.Debug("Output:\n", string(out))
	if err != nil {
		log.Error(err)
	}
	return err
}

func isGitHub() bool {
	tmp, exists := os.LookupEnv("GITHUB_REF")
	return exists && len(tmp) > 0
}

func isGitLab() bool {
	tmp, exists := os.LookupEnv("GITLAB_CI")
	return exists && len(tmp) > 0
}

// Adds the string to the prerelease in the version
func addPrerelease(ver semver.Version, prerelease string) semver.Version {
	tmp := ver.Prerelease()
	if len(tmp) > 0 {
		tmp = tmp + "-"
	}
	tmp = tmp + prerelease
	return setPrerelease(ver, tmp)
}

// Sets the prerelease in the version
func setPrerelease(ver semver.Version, prerelease string) semver.Version {
	result, err := ver.SetPrerelease(prerelease)
	if err != nil {
		log.Errorf("Error set pre-release %s version %s", prerelease, ver.String())
		log.Panic(err)
	}
	return result
}

// Create patch branch name from the version
func createPatchBranchName(ver *semver.Version, prefix string) string {
	return prefix + strconv.FormatInt(ver.Major(), 10) + "." + strconv.FormatInt(ver.Minor(), 10)
}

// <VERSION>-<BUILD>-<HASH> - do not increment the version
func createProjectBuildVersion(ver, count, hash, prefix string, length int) semver.Version {
	tmp := createVersion(ver)
	return addBuildInfo(*tmp, count, hash, prefix, length)
}

// <VERSION>-<BUILD>-<HASH> - increment the version
func createBuildVersion(ver, count, hash, prefix string, length int) semver.Version {
	tmp := nextReleaseVersion(createVersion(ver), false)
	return addBuildInfo(tmp, count, hash, prefix, length)
}

func addBuildInfo(tmp semver.Version, count, hash, prefix string, length int) semver.Version {
	pre := tmp.Prerelease()
	if len(count) > 0 {
		if len(pre) > 0 {
			pre = pre + "."
		}
		count = lpad(count, "0", length)
		pre = pre + prefix + count
	}
	if len(hash) > 0 {
		if len(pre) > 0 {
			pre = pre + "."
		}
		pre = pre + hash
	}
	return setPrerelease(tmp, pre)
}

// <VERSION>
func createVersion(ver string) *semver.Version {
	tmp := ver
	result, e := semver.NewVersion(tmp)
	if e != nil {
		log.Panic(e)
	}
	return result
}

// Creates next release version
func nextReleaseVersion(ver *semver.Version, major bool) semver.Version {
	if major {
		if ver.Patch() != 0 {
			log.Errorf("Can not created major release from the patch version  [%s]!", ver)
			os.Exit(0)
		}
		tmp := ver.IncMajor()
		return tmp
	}
	prerelease := ver.Prerelease()
	if len(prerelease) > 0 {
		update, data := nextPrerelease(prerelease)
		if update {
			return setPrerelease(*ver, data)
		}
	}
	if ver.Patch() != 0 {
		return ver.IncPatch()
	}
	return ver.IncMinor()
}

// Finds the number in the prerelease and increment
func nextPrerelease(data string) (bool, string) {
	if len(data) == 0 {
		return false, data
	}
	numbers := numberRegex.FindAllString(data, -1)
	if len(numbers) > 0 {
		number := numbers[len(numbers)-1]
		index, e := strconv.ParseInt(number, 10, 64)
		if e != nil {
			log.Panic(e)
		}
		index = index + 1
		data = strings.TrimSuffix(data, number)
		data = data + strconv.FormatInt(index, 10)
		return true, data
	}
	return false, data
}

func lpad(data, pad string, length int) string {
	for i := len(data); i < length; i++ {
		data = pad + data
	}
	return data
}

// x.x.x-<pre>.rc0.hash -> x.x.x-<pre>.hash
func updatePrereleaseToHashVersion(ver string, length int) string {
	pre := ver
	if len(pre) > 0 {
		hi := strings.LastIndex(pre, ".")
		if hi != -1 {
			pre = pre[0:hi]
			ri := strings.LastIndex(pre, ".")
			if ri != -1 {
				pre = pre[0:ri]
			} else {
				pre = ""
			}
		}
	}
	_, _, hash := gitCommit(length)
	if len(pre) > 0 {
		pre = pre + "."
	}
	return pre + hash
}
