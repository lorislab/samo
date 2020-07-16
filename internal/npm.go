package internal

import (
	"github.com/Masterminds/semver"
	log "github.com/sirupsen/logrus"
	"strings"
)

const (
	npmVersion string = "/version"
	npmName    string = "/name"
)

// NpmProject maven project
type NpmProject struct {
	filename      string
	name, version *XPathItem
}

// ReleaseSemVersion the release version
func (r NpmProject) ReleaseSemVersion() *semver.Version {
	tmp := r.Version()
	index := strings.Index(tmp, "-")
	if index != -1 {
		tmp = tmp[0:index]
	}
	return createVersion(tmp)
}

// ReleaseVersion the npm project release version
func (r NpmProject) ReleaseVersion() string {
	return r.ReleaseSemVersion().String()
}

// Name the npm project name
func (r NpmProject) Name() string {
	return r.name.value
}

// Version the npm project version
func (r NpmProject) Version() string {
	return r.version.value
}

// Filename the project filename
func (r NpmProject) Filename() string {
	return r.filename
}

// SetVersion set project version
func (r NpmProject) SetVersion(version string) {
	ReplaceTextInFile(r.filename, version, r.version.begin(), r.version.end())
}

// LoadNpmProject load maven project
func LoadNpmProject(filename string) *NpmProject {
	items := []string{npmName, npmVersion}
	result := jsonPathInFile(filename, items)
	if result.IsEmpty() {
		log.Warnf("The file '%s' does not have npm structure.\n", filename)
		return nil
	}
	project := &NpmProject{filename: filename, name: result.items[npmName], version: result.items[npmVersion]}
	return project
}
