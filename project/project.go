package project

import (
	"bytes"
	"html/template"

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
	// IsFile Project file
	IsFile() bool
	// SetVersion set project version
	SetVersion(version string)
}

type ProjectRequest struct {
	Project     Project
	Versions    Versions
	CommitMsg   string
	TagMsg      string
	SkipPush    bool
	SkipNextDev bool
	Tag         string
	PathBranch  string
}

// CreateRelease create project release
func (r ProjectRequest) Release() {

	tag := r.Versions.ReleaseVersion()

	data := NextDevMsg{
		Version: tag,
	}
	msg := createText(data, r.TagMsg)
	tools.Git("tag", "-a", tag, "-m", msg)

	// update project file with next version
	r.releaseNextDev()

	// push project to remote repository
	if r.SkipPush {
		log.WithField("tag", tag).Info("Skip git push for project release")
	} else {
		tools.Git("push", "--tags")
	}
	log.WithField("version", tag).Info("New release created.")
}

// Update project file with new dev version
func (r ProjectRequest) releaseNextDev() {
	if r.SkipNextDev || !r.Project.IsFile() {
		return
	}

	currentVersion := r.Versions.SemVer()
	tmp := r.Versions.NextReleaseVersion()

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

	r.Project.SetVersion(devVersion)
	log.WithField("dev-version", devVersion).Info("Update project version to next development version")

	data := NextDevMsg{
		Version: devVersion,
	}
	msg := createText(data, r.CommitMsg)

	tools.Git("add", ".")
	tools.Git("commit", "-m", msg)

	log.WithFields(log.Fields{
		"version":     currentVersion.String(),
		"new-version": devVersion,
	}).Debug("Switch project to the new version")

	if r.SkipPush {
		log.WithField("dev", devVersion).Info("Skip git push for next project dev version")
	} else {
		tools.Git("push")
	}
}

// CreatePatch create patch fo the project
func (r ProjectRequest) Patch() {

	tagVer := SemVer(r.Tag)
	if tagVer.Patch() != 0 || len(tagVer.Prerelease()) > 0 {
		log.WithField("tag", tagVer.Original()).Fatal("Can not created patch branch from the patch version!")
	}

	branch := createText(tagVer, r.PathBranch)
	tools.Git("checkout", "-b", branch, r.Tag)
	log.WithField("branch", branch).Debug("Branch created")

	// update project file
	r.patchNextDev(tagVer)

	// push changes
	if r.SkipPush {
		log.WithField("branch", branch).Info("Skip git push for project patch version")
	} else {
		tools.Git("push", "-u", "origin", branch)
	}
	log.WithField("branch", branch).Info("New patch branch created.")
}

type NextDevMsg struct {
	Version string
}

// update project file with next version
func (r ProjectRequest) patchNextDev(tagVer *semver.Version) {
	if r.SkipNextDev || !r.Project.IsFile() {
		return
	}

	// remove the prerelease
	patchVer := tagVer.IncPatch()

	version := r.Versions.SemVer()

	// add suffix (maven = snapshot)
	patchVer, e := patchVer.SetPrerelease(version.Prerelease())
	if e != nil {
		log.WithField("version", patchVer.String()).Fatal("Error add pre-release to the version")
	}

	// set version to project file
	patchVersion := patchVer.String()
	r.Project.SetVersion(patchVersion)
	log.WithField("patch-version", patchVersion).Info("Update project version to next patch development version")

	tools.Git("add", ".")

	data := NextDevMsg{
		Version: patchVersion,
	}
	msg := createText(data, r.CommitMsg)
	tools.Git("commit", "-m", msg)
}

func createText(obj interface{}, data string) string {
	template := template.New("template")
	t, err := template.Parse(data)
	if err != nil {
		log.Panic(err)
	}

	var tpl bytes.Buffer
	err = t.Execute(&tpl, obj)
	if err != nil {
		log.Panic(err)
	}
	return tpl.String()
}
