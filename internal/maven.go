package internal

import (
	"fmt"
	"io/ioutil"

	"github.com/Masterminds/semver"
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
	groupdID, artifactID, version *XPathItem
}

// Equals returns true if the maven ID's are equal
func (r MavenID) Equals(m *MavenID) bool {
	return r.ID() == m.ID()
}

// ID the string representation of the maven ID
func (r MavenID) ID() string {
	return fmt.Sprintf("%s:%s:%s", r.groupdID.value, r.artifactID.value, r.version.value)
}

// Version the maven project version
func (r MavenID) Version() string {
	return r.version.value
}

// GroupID the maven project version
func (r MavenID) GroupID() string {
	return r.groupdID.value
}

// ArtifactID the maven project version
func (r MavenID) ArtifactID() string {
	return r.artifactID.value
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

// ArtifactID the maven project artifact id
func (r MavenProject) ArtifactID() string {
	return r.projectID.ArtifactID()
}

// Version the maven project version
func (r MavenProject) Version() string {
	return r.projectID.Version()
}

// NextReleaseVersion release version of the project
func (r MavenProject) NextReleaseVersion() string {
	tmp, err := semver.NewVersion(r.projectID.version.value)
	if err != nil {
		panic(err)
	}

	var result = semver.Version{}
	if tmp.Patch() == 0 {
		result = tmp.IncMinor()
	} else {
		result = tmp.IncPatch()
	}
	result, err = tmp.SetPrerelease("SNAPSHOT")
	if err != nil {
		panic(err)
	}
	return result.String()
}

// ReleaseVersion release version of the project
func (r MavenProject) ReleaseVersion() string {
	semver, err := semver.NewVersion(r.projectID.version.value)
	if err != nil {
		panic(err)
	}
	ver, err := semver.SetPrerelease("")
	if err != nil {
		panic(err)
	}
	ver, err = ver.SetMetadata("")
	if err != nil {
		panic(err)
	}
	return ver.String()
}

// SetPrerelease release version of the project
func (r MavenProject) SetPrerelease(prerelease string) string {
	semver, err := semver.NewVersion(r.projectID.version.value)
	if err != nil {
		panic(err)
	}
	newVersion, err := semver.SetPrerelease(prerelease)
	if err != nil {
		panic(err)
	}
	return newVersion.String()
}

// SetVersion set project version
func (r MavenProject) SetVersion(version string) {
	replaceTextInFile(r.filename, version, r.projectID.version.begin(), r.projectID.version.end())
	fmt.Printf("Update project '%s' version from [%s] to [%s]\n", r.filename, r.projectID.version.value, version)
}

// SetParentVersion set project parent version
func (r MavenProject) SetParentVersion(version string) {
	replaceTextInFile(r.filename, version, r.parentProjectID.version.begin(), r.parentProjectID.version.end())
	fmt.Printf("Update project '%s' parent version from [%s] to [%s]\n", r.filename, r.parentProjectID.version.value, version)
}

// LoadMavenProject load maven project
func LoadMavenProject(filename string) *MavenProject {
	items := []string{projectVersion, projectGroupID, projectArtifactID, parentProjectArtifactID, parentProjectGroupID, parentProjectVersion}
	result := FindXPathInFile(filename, items)
	if result.IsEmpty() {
		fmt.Printf("The file '%s' does not have maven structure.\n", filename)
		return nil
	}
	projectID := &MavenID{groupdID: result.items[projectGroupID], artifactID: result.items[projectArtifactID], version: result.items[projectVersion]}

	var parentProjectID *MavenID
	if result.items[parentProjectGroupID] != nil {
		parentProjectID = &MavenID{groupdID: result.items[parentProjectGroupID], artifactID: result.items[parentProjectArtifactID], version: result.items[parentProjectVersion]}
	}
	project := &MavenProject{filename: filename, projectID: projectID, parentProjectID: parentProjectID}
	return project
}

func replaceTextInFile(filename, text string, b, e int64) {
	buf, err := ioutil.ReadFile(filename)
	if err != nil {
		panic(err)
	}
	result := string(buf)
	result = result[:b] + text + result[e:]
	ioutil.WriteFile(filename, []byte(result), 0666)
}
