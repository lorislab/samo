package internal

import (
	"github.com/Masterminds/semver"
	log "github.com/sirupsen/logrus"
)

// Project common project interface
type Project interface {
	// Name project name
	Name() string
	// Version project version
	Version() string
	// Filename project file name
	Filename() string
	// ReleaseVersion project release version
	ReleaseVersion() string
	// ReleaseVersion project release version
	ReleaseSemVersion() *semver.Version
	// SetVersion set project version
	SetVersion(version string)
}

// <VERSION>
func createVersion(ver string) *semver.Version {
	tmp := ver
	result, e := semver.NewVersion(tmp)
	if e != nil {
		log.Error("Version value is not valid semver 2.0. Value: " + ver)
		log.Panic(e)
	}
	return result
}
