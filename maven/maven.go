package maven

import (
	"fmt"
	"os"

	"github.com/lorislab/samo/project"
	"github.com/lorislab/samo/tools"
	"github.com/lorislab/samo/xml"
	log "github.com/sirupsen/logrus"
)

const (
	projectVersion          string = "/project/version"
	projectGroupID          string = "/project/groupId"
	projectArtifactID       string = "/project/artifactId"
	parentProjectVersion    string = "/project/parent/version"
	parentProjectGroupID    string = "/project/parent/groupId"
	parentProjectArtifactID string = "/project/parent/artifactId"
)

// MavenID maven project ID
type MavenID struct {
	groupID, artifactID, version *xml.XPathItem
}

// Equals returns true if the maven ID's are equal
func (r MavenID) Equals(m *MavenID) bool {
	return r.ID() == m.ID()
}

// ID the string representation of the maven ID
func (r MavenID) ID() string {
	return fmt.Sprintf("%s:%s:%s", r.groupID.Value, r.artifactID.Value, r.version.Value)
}

// Version the maven project version
func (r MavenID) Version() string {
	return r.version.Value
}

// GroupID the maven project version
func (r MavenID) GroupID() string {
	return r.groupID.Value
}

// ArtifactID the maven project version
func (r MavenID) ArtifactID() string {
	return r.artifactID.Value
}

// MavenProject maven project
type MavenProject struct {
	filename                   string
	projectID, parentProjectID *MavenID
}

// HasParent returns true if project has a parent project
func (r MavenProject) HasParent() bool {
	return r.parentProjectID != nil
}

// ID the maven project Id
func (r MavenProject) ID() string {
	return r.projectID.ID()
}

// Filename the maven project filename
func (r MavenProject) Filename() string {
	return r.filename
}

// Type the type of the project
func (r MavenProject) Type() project.Type {
	return project.Maven
}

// Name the maven project name
func (r MavenProject) Name() string {
	return r.projectID.ArtifactID()
}

// Version the maven project version
func (r MavenProject) Version() string {
	return r.projectID.Version()
}

// IsFile is project base on the project file
func (r MavenProject) IsFile() bool {
	return true
}

// SetVersion set project version
func (r MavenProject) SetVersion(version string) {
	tools.ReplaceTextInFile(r.filename, version, r.projectID.version.Begin(), r.projectID.version.End())
}

// Load load maven project
func Load(filename string) project.Project {
	if len(filename) == 0 {
		filename = "pom.xml"
		if _, err := os.Stat(filename); os.IsNotExist(err) {
			log.WithField("file", filename).Debug("Missing default maven project file pom.xml")
			return nil
		}
	}
	return create(filename)
}

func create(filename string) project.Project {
	items := []string{projectVersion, projectGroupID, projectArtifactID, parentProjectArtifactID, parentProjectGroupID, parentProjectVersion}
	result := xml.FindXPathInFile(filename, items)

	if result.IsEmpty() {
		log.WithField("file", filename).Fatal("The file does not have maven project structure.")
	}

	projectID := &MavenID{groupID: result.Items[projectGroupID], artifactID: result.Items[projectArtifactID], version: result.Items[projectVersion]}

	var parentProjectID *MavenID
	if result.Items[parentProjectGroupID] != nil {
		parentProjectID = &MavenID{groupID: result.Items[parentProjectGroupID], artifactID: result.Items[parentProjectArtifactID], version: result.Items[parentProjectVersion]}
	}
	project := MavenProject{
		filename:        filename,
		projectID:       projectID,
		parentProjectID: parentProjectID,
	}
	return &project
}
