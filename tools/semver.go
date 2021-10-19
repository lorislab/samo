package tools

import (
	"github.com/Masterminds/semver"
	"github.com/lorislab/samo/log"
)

// SemVer create SemVer version for the project
func CreateSemVer(version string) *semver.Version {
	result, e := semver.NewVersion(version)
	if e != nil {
		log.Fatal("Version value is not valid semver 2.0.", log.F("version", version))
	}
	return result
}
