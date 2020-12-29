package npm

import (
	"os"

	"github.com/lorislab/samo/file"
	"github.com/lorislab/samo/json"
	"github.com/lorislab/samo/project"
	"github.com/lorislab/samo/xml"
	log "github.com/sirupsen/logrus"
)

const (
	npmVersion string = "/version"
	npmName    string = "/name"
)

// NpmProject npm project
type NpmProject struct {
	filename      string
	name, version *xml.XPathItem
}

// Type the npm project type
func (r NpmProject) Type() project.Type {
	return project.Npm
}

// Name the npm project name
func (r NpmProject) Name() string {
	return r.name.Value
}

// Version the npm project version
func (r NpmProject) Version() string {
	return r.version.Value
}

// Filename the project filename
func (r NpmProject) Filename() string {
	return r.filename
}

// SetVersion set project version
func (r NpmProject) SetVersion(version string) {
	file.ReplaceTextInFile(r.filename, version, r.version.Begin(), r.version.End())
}

func Load(filename string) project.Project {
	if len(filename) == 0 {
		filename = "package.json"
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			log.WithField("file", filename).Debug("Missing default npm project file package.json")
			return nil
		}
	}
	return create(filename)
}

func create(filename string) project.Project {
	items := []string{npmName, npmVersion}
	result := json.PathInFile(filename, items)
	if result.IsEmpty() {
		log.WithField("filename", filename).Fatal("The file does not have npm structure.")
		return nil
	}
	project := NpmProject{
		filename: filename,
		name:     result.Items[npmName],
		version:  result.Items[npmVersion],
	}
	return &project
}
