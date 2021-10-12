package tools

import (
	"github.com/Masterminds/semver"
	log "github.com/sirupsen/logrus"
)

// SemVer create SemVer version for the project
func CreateSemVer(version string) *semver.Version {
	result, e := semver.NewVersion(version)
	if e != nil {
		log.WithField("version", version).Fatal("Version value is not valid semver 2.0.")
	}
	return result
}
