package internal

import (
	log "github.com/sirupsen/logrus"
)

const (
	npmVersion string = "/version"
	npmName    string = "/name"
)

// MavenProject maven project
type NpmProject struct {
	filename      string
	name, version *XPathItem
}

// Version the npm project name
func (r NpmProject) Name() string {
	return r.name.value
}

// Version the npm project version
func (r NpmProject) Version() string {
	return r.version.value
}

// SetVersion set project version
func (r NpmProject) SetVersion(version string) {
	ReplaceTextInFile(r.filename, version, r.version.begin(), r.version.end())
	log.Infof("Update project '%s' version from [%s] to [%s]\n", r.filename, r.version.value, version)
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
