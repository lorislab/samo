package cmd

import (
	"bytes"
	"github.com/Masterminds/semver"
	"os"
	"os/exec"
	"strconv"
	"strings"

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
			t, e := ver.SetPrerelease(data)
			if e != nil {
				log.Panic(e)
			}
			return t
		}
	}

	if ver.Patch() != 0 {
		return ver.IncPatch()
	}

	return ver.IncMinor()
}

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

func createBuildVersionItems(ver *semver.Version) (int64, int64, int64, string) {
	prerelease := ver.Prerelease()
	patch := ver.Patch()
	minor := ver.Minor()

	if len(prerelease) > 0 {
		update, data := nextPrerelease(prerelease)
		if update {
			prerelease = data + "."
		}
	} else if ver.Patch() != 0 {
		patch = patch + 1
	} else {
		minor = minor + 1
	}
	return ver.Major(), minor, patch, prerelease
}

func createBuildVersionFromItems(major, minor, patch int64, prerelease, buildPrefix, count, build string) *semver.Version {
	prerelease = prerelease + buildPrefix + count
	return createVersion(major, minor, patch, prerelease, build)
}

//func createBuildVersion(ver *semver.Version, buildPrefix, count, build string) *semver.Version {
//	major, minor, patch, prerelease := createBuildVersionItems(ver)
//	return createBuildVersionFromItems(major, minor, patch, prerelease, buildPrefix, count, build)
//}

func createVersion(major, minor, path int64, prerelease, build string) *semver.Version {
	r := strconv.FormatInt(major, 10) + "." + strconv.FormatInt(minor, 10) + "." + strconv.FormatInt(path, 10)
	if len(prerelease) > 0 {
		r = r + "-" + prerelease
	}
	if len(build) > 0 {
		r = r + "+" + build
	}
	result, e := semver.NewVersion(r)
	if e != nil {
		log.Panic(e)
	}
	return result
}
