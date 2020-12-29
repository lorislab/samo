package project

import (
	"strconv"

	"github.com/Masterminds/semver"
	"github.com/lorislab/samo/tools"
	log "github.com/sirupsen/logrus"
)

type Type string

const (
	Maven = "maven"
	Npm   = "npm"
	Git   = "git"
)

// Project common project interface
type Project interface {
	// Type the type of the project
	Type() Type
	// Name project name
	Name() string
	// Version project version
	Version() string
	// Filename project file name
	Filename() string
	// SetVersion set project version
	SetVersion(version string)
}

// ReleaseVersion the release version
func ReleaseVersion(project Project) *semver.Version {
	tmp := CreateVersion(project)
	version, err := tmp.SetPrerelease("")
	if err != nil {
		log.WithField("version", project.Version()).Fatal("Error remove pre-release label from version")
	}
	version, err = version.SetMetadata("")
	if err != nil {
		log.WithField("version", project.Version()).Fatal("Error remove metadata label from version")
	}
	return &version
}

// CreateVersion create SemVer version for the project
func CreateVersion(project Project) *semver.Version {
	tmp := project.Version()
	result, e := semver.NewVersion(tmp)
	if e != nil {
		log.WithField("version", project.Version()).Fatal("Version value is not valid semver 2.0.")
	}
	return result
}

// BuildVersion build version of the project
func BuildVersion(project Project, hashLength, length int, prefix string) *semver.Version {
	tmp := ReleaseVersion(project)
	_, count, hash := tools.GitCommit(hashLength)

	pre := prefix

	// ldap count
	if len(count) > 0 {
		pre = pre + tools.Lpad(count, "0", length)
	}

	// add hash
	if len(hash) > 0 {
		if len(pre) > 0 {
			pre = pre + "."
		}
		pre = pre + hash
	}

	result, err := tmp.SetPrerelease(pre)
	if err != nil {
		log.WithFields(log.Fields{
			"prerelease": pre,
			"version":    tmp.String(),
		}).Fatal("Error set pre-release")
	}
	return &result
}

func HashVersion(project Project, hashLength int) *semver.Version {
	_, _, hash := tools.GitCommit(hashLength)
	tmp := ReleaseVersion(project)
	ver, err := tmp.SetPrerelease(hash)
	if err != nil {
		log.WithFields(log.Fields{
			"hash":    hash,
			"version": tmp.String(),
		}).Fatal("Error set hash pre-release label to the version")
	}
	return &ver
}

// CreateRelease create project release
func CreateRelease(project Project, commitMessage, tagMessage string, major, skipPush bool) string {

	tag := ReleaseVersion(project).String()
	if len(tagMessage) == 0 {
		tagMessage = tag
	}
	tools.Git("tag", "-a", tag, "-m", tagMessage)

	// update project file with next version
	if len(project.Filename()) > 0 {
		currentVersion := CreateVersion(project)
		tmp := NextReleaseVersion(project, major)
		tmp, err := tmp.SetPrerelease(currentVersion.Prerelease())
		if err != nil {
			log.WithFields(log.Fields{
				"prerelease": currentVersion.Prerelease(),
				"version":    tmp.String(),
			}).Fatal("Error set pre-release")
		}
		tmp, err = tmp.SetMetadata(currentVersion.Metadata())
		if err != nil {
			log.WithFields(log.Fields{
				"prerelease": currentVersion.Metadata(),
				"version":    tmp.String(),
			}).Fatal("Error set metada")
		}
		devVersion := tmp.String()

		project.SetVersion(devVersion)
		tools.Git("add", ".")
		tools.Git("commit", "-m", commitMessage+" ["+devVersion+"]")
		log.WithFields(log.Fields{
			"version":     currentVersion.String(),
			"new-version": devVersion,
		}).Debug("Switch project to the new version")
	}

	// push project to remote repository
	if !skipPush {
		tools.Git("push")
		tools.Git("push", "--tags")
	} else {
		log.WithField("tag", tag).Info("Skip git push for project release")
	}
	return tag
}

// NextReleaseVersion creates next release version
func NextReleaseVersion(project Project, major bool) semver.Version {
	ver := CreateVersion(project)
	if major {
		if ver.Patch() != 0 {
			log.WithField("version", ver.String()).Fatal("Can not created major release from the patch version!")
		}
		tmp := ver.IncMajor()
		return tmp
	}
	if ver.Patch() != 0 {
		return ver.IncPatch()
	}
	return ver.IncMinor()
}

// CreatePatch create patch fo the project
func CreatePatch(project Project, commitMessage, patchTag, branchPrefix string, skipPush bool) string {

	tagVer, e := semver.NewVersion(patchTag)
	if e != nil {
		log.WithField("tag", patchTag).Fatal("The patch tag is not valid version.")
	}
	if tagVer.Patch() != 0 || len(tagVer.Prerelease()) > 0 {
		log.WithField("tag", tagVer.Original()).Fatal("Can not created patch branch from the patch version!")
	}

	branch := branchPrefix + strconv.FormatInt(tagVer.Major(), 10) + "." + strconv.FormatInt(tagVer.Minor(), 10)
	tools.Git("checkout", "-b", branch, patchTag)
	log.WithField("branch", branch).Debug("Branch created")

	// update project file with next version
	if len(project.Filename()) > 0 {
		// remove the prerelease
		ver := tagVer.IncPatch()

		version := CreateVersion(project)

		// add suffix (maven = snapshot)
		ver, e = ver.SetPrerelease(version.Prerelease())
		if e != nil {
			log.WithField("version", ver.String()).Fatal("Error add pre-release to the version")
		}

		// set version to project file
		patchVersion := ver.String()
		project.SetVersion(patchVersion)

		tools.Git("add", ".")
		tools.Git("commit", "-m", commitMessage+" ["+patchVersion+"]")
	}

	if !skipPush {
		tools.Git("push", "-u", "origin", branch)
	} else {
		log.WithField("branch", branch).Info("Skip git push for project patch version")
	}
	return branch
}
