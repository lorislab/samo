package cmd

import (
	"bytes"
	"os"
	"os/exec"
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
}

func addFlagRequired(command *cobra.Command, name, shorthand string, value string, usage string) {
	addFlag(command, name, shorthand, value, usage)
	err := command.MarkFlagRequired(name)
	if err != nil {
		log.Panic(err)
	}
}

func addFlag(command *cobra.Command, name, shorthand string, value string, usage string) {
	command.Flags().StringP(name, shorthand, value, usage)
	addViper(command, name)
}

func addBoolFlag(command *cobra.Command, name, shorthand string, value bool, usage string) {
	command.Flags().BoolP(name, shorthand, value, usage)
	addViper(command, name)
}

func addGitHashLength(command *cobra.Command, name string) {
	addFlag(command, name, "l", "12", "The git hash length")
}

func addViper(command *cobra.Command, name string) {
	err := viper.BindPFlag(name, command.Flags().Lookup(name))
	if err != nil {
		panic(err)
	}
}

func execCmd(name string, arg ...string) {
	log.Info(name+" ", strings.Join(arg, " "))
	out, err := exec.Command(name, arg...).CombinedOutput()
	log.Debug("Output:\n", string(out))
	if err != nil {
		log.Panic(err)
	}
}

func execCmdOutput(name string, arg ...string) string {
	log.Debug(name+" ", strings.Join(arg, " "))
	out, err := exec.Command(name, arg...).CombinedOutput()
	log.Debug("Output:\n", string(out))
	if err != nil {
		log.Panic(err)
	}
	return string(bytes.TrimRight(out, "\n"))
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
		log.Panic(err)
	}
	return result
}

// Create patch branch name from the version
func createPatchBranchName(ver *semver.Version) string {
	return strconv.FormatInt(ver.Major(), 10) + "." + strconv.FormatInt(ver.Minor(), 10)
}

// <VERSION>-<BUILD>-<HASH> - do not increment the version
func createMavenBuildVersion(ver, count, hash, prefix string) semver.Version {
	tmp := createVersion(ver)
	return addBuildInfo(*tmp, count, hash, prefix)
}

// <VERSION>-<BUILD>-<HASH> - increment the version
func createBuildVersion(ver, count, hash, prefix string) semver.Version {
	tmp := nextReleaseVersion(createVersion(ver), false)
	return addBuildInfo(tmp, count, hash, prefix)
}

func addBuildInfo(tmp semver.Version, count, hash, prefix string) semver.Version {
	pre := tmp.Prerelease()
	if len(count) > 0 {
		if len(pre) > 0 {
			pre = pre + "."
		}
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
