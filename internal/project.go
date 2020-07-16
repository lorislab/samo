package internal

import (
	"bufio"
	"bytes"
	"github.com/Masterminds/semver"
	log "github.com/sirupsen/logrus"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
)

var numberRegex *regexp.Regexp

// Project common project interface
type Project interface {
	// Name project name
	Name() string
	// Version project version
	Version() string
	// Filename project file name
	Filename() string
	// ReleaseVersion project release version
	ReleaseSemVersion() *semver.Version
	// SetVersion set project version
	SetVersion(version string)
	// NextReleaseSuffix next release suffix
	NextReleaseSuffix() string
}

func init() {
	numberRegex = regexp.MustCompile("[0-9]+")
}

// CreateVersion create version
func CreateVersion(ver string) *semver.Version {
	tmp := ver
	result, e := semver.NewVersion(tmp)
	if e != nil {
		log.Error("Version value is not valid semver 2.0. Value: " + ver)
		log.Panic(e)
	}
	return result
}

// NextReleaseVersion creates next release version
func NextReleaseVersion(ver *semver.Version, major bool) semver.Version {
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
			return SetPrerelease(*ver, data)
		}
	}
	if ver.Patch() != 0 {
		return ver.IncPatch()
	}
	return ver.IncMinor()
}

// SetPrerelease sets the prerelease in the version
func SetPrerelease(ver semver.Version, prerelease string) semver.Version {
	result, err := ver.SetPrerelease(prerelease)
	if err != nil {
		log.Errorf("Error set pre-release %s version %s", prerelease, ver.String())
		log.Panic(err)
	}
	return result
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

// ExecCmdOutput execute command with output
func ExecCmdOutput(name string, arg ...string) string {
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

// ExecGitCmd execute git command
func ExecGitCmd(name string, arg ...string) {
	err := execCmdErr(name, arg...)
	if err != nil {
		ExecCmd("rm", "-f", "o.git/index.lock")
	}
}

// ExecCmd execute command
func ExecCmd(name string, arg ...string) {
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

// AddPrerelease adds the string to the prerelease in the version
func AddPrerelease(ver semver.Version, prerelease string) semver.Version {
	tmp := ver.Prerelease()
	if len(tmp) > 0 {
		tmp = tmp + "-"
	}
	tmp = tmp + prerelease
	return SetPrerelease(ver, tmp)
}

// AddBuildInfo add build info to version
func AddBuildInfo(tmp semver.Version, count, hash, prefix string, length int) semver.Version {
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
	return SetPrerelease(tmp, pre)
}

// UpdatePrereleaseToHashVersion x.x.x-<pre>.rc0.hash -> x.x.x-<pre>.hash
func UpdatePrereleaseToHashVersion(ver string, length int) string {
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
	_, _, hash := GitCommit(length)
	if len(pre) > 0 {
		pre = pre + "."
	}
	return pre + hash
}


func lpad(data, pad string, length int) string {
	for i := len(data); i < length; i++ {
		data = pad + data
	}
	return data
}
