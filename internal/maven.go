package internal

import (
	"fmt"
	"github.com/Masterminds/semver"
	log "github.com/sirupsen/logrus"
	"io/ioutil"
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

func (r MavenProject) semVer() *semver.Version {
	tmp, err := semver.NewVersion(r.projectID.version.value)
	if err != nil {
		log.Panic(err)
	}
	return tmp
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

// NextPatchVersion release version of the project
func (r MavenProject) NextPatchVersion() string {
	tmp := r.semVer()
	result := setPrerelease(*tmp, "")
	result = result.IncPatch()
	result = setPrerelease(result, "SNAPSHOT")
	return result.String()
}

// NextReleaseVersion release version of the project
func (r MavenProject) NextReleaseVersion(major bool) string {
	tmp := r.semVer()
	var result = semver.Version{}
	if major {
		result = tmp.IncMajor()
	} else {
		if tmp.Patch() == 0 {
			result = tmp.IncMinor()
		} else {
			result = tmp.IncPatch()
		}
	}
	result = setPrerelease(result, "SNAPSHOT")
	return result.String()
}

func setPrerelease(ver semver.Version, pre string) semver.Version {
	result, err := ver.SetPrerelease(pre)
	if err != nil {
		log.Panic(err)
	}
	return result
}

// ReleaseVersion release version of the project
func (r MavenProject) ReleaseVersion() string {
	semver := r.semVer()
	ver := setPrerelease(*semver, "SNAPSHOT")
	ver = ver.IncPatch()
	return ver.String()
}

// SetPrerelease release version of the project
func (r MavenProject) SetPrerelease(prerelease string) string {
	semver := r.semVer()
	newVersion := setPrerelease(*semver, prerelease)
	return newVersion.String()
}

// SetVersion set project version
func (r MavenProject) SetVersion(version string) {
	replaceTextInFile(r.filename, version, r.projectID.version.begin(), r.projectID.version.end())
	log.Infof("Update project '%s' version from [%s] to [%s]\n", r.filename, r.projectID.version.value, version)
}

// SetParentVersion set project parent version
func (r MavenProject) SetParentVersion(version string) {
	replaceTextInFile(r.filename, version, r.parentProjectID.version.begin(), r.parentProjectID.version.end())
	log.Infof("Update project '%s' parent version from [%s] to [%s]\n", r.filename, r.parentProjectID.version.value, version)
}

// LoadMavenProject load maven project
func LoadMavenProject(filename string) *MavenProject {
	items := []string{projectVersion, projectGroupID, projectArtifactID, parentProjectArtifactID, parentProjectGroupID, parentProjectVersion}
	result := FindXPathInFile(filename, items)
	if result.IsEmpty() {
		log.Warnf("The file '%s' does not have maven structure.\n", filename)
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
		log.Panic(err)
	}
	result := string(buf)
	result = result[:b] + text + result[e:]
	ioutil.WriteFile(filename, []byte(result), 0666)
}
