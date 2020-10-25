package internal

import (
	"strings"

	"github.com/Masterminds/semver/v3"
	log "github.com/sirupsen/logrus"
)

const (
	npmVersion string = "/version"
	npmName    string = "/name"
)

// NpmProject npm project
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
	return CreateVersion(tmp)
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

// NextReleaseSuffix next release suffix
func (r NpmProject) NextReleaseSuffix() string {
	return ""
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
