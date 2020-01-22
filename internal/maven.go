package internal

import (
	"fmt"
	"github.com/Masterminds/semver"
	log "github.com/sirupsen/logrus"
	"os"
)

const (
	projectVersion          string = "/project/version"
	projectGroupID          string = "/project/groupId"
	projectArtifactID       string = "/project/artifactId"
	parentProjectVersion    string = "/project/parent/version"
	parentProjectGroupID    string = "/project/parent/groupId"
	parentProjectArtifactID string = "/project/parent/artifactId"
	mavenSettings			string = "/settings"
	mavenSettingsServers	string = "/settings/servers"
)

// MavenID maven project ID
type MavenID struct {
	groupID, artifactID, version *XPathItem
}

// Equals returns true if the maven ID's are equal
func (r MavenID) Equals(m *MavenID) bool {
	return r.ID() == m.ID()
}

// ID the string representation of the maven ID
func (r MavenID) ID() string {
	return fmt.Sprintf("%s:%s:%s", r.groupID.value, r.artifactID.value, r.version.value)
}

// Version the maven project version
func (r MavenID) Version() string {
	return r.version.value
}

// GroupID the maven project version
func (r MavenID) GroupID() string {
	return r.groupID.value
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
	tmp := r.semVer()
	ver := setPrerelease(*tmp, "SNAPSHOT")
	ver = ver.IncPatch()
	return ver.String()
}

// SetPrerelease release version of the project
func (r MavenProject) SetPrerelease(prerelease string) string {
	ver := r.semVer()
	newVersion := setPrerelease(*ver, prerelease)
	return newVersion.String()
}

// SetVersion set project version
func (r MavenProject) SetVersion(version string) {
	ReplaceTextInFile(r.filename, version, r.projectID.version.begin(), r.projectID.version.end())
	log.Infof("Update project '%s' version from [%s] to [%s]\n", r.filename, r.projectID.version.value, version)
}

// SetParentVersion set project parent version
func (r MavenProject) SetParentVersion(version string) {
	ReplaceTextInFile(r.filename, version, r.parentProjectID.version.begin(), r.parentProjectID.version.end())
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
	projectID := &MavenID{groupID: result.items[projectGroupID], artifactID: result.items[projectArtifactID], version: result.items[projectVersion]}

	var parentProjectID *MavenID
	if result.items[parentProjectGroupID] != nil {
		parentProjectID = &MavenID{groupID: result.items[parentProjectGroupID], artifactID: result.items[parentProjectArtifactID], version: result.items[parentProjectVersion]}
	}
	project := &MavenProject{filename: filename, projectID: projectID, parentProjectID: parentProjectID}
	return project
}

// Creates or updates the maven server settings in the maven settings file
func CreateMavenSettingsServer(filename, id, username, password string) {
	info, err := os.Stat(filename)

	// if not exists create new file with maven server settings
	if os.IsNotExist(err) {
		WriteToFile(filename, createMavenSettingsServer(id, username, password, "<settings><servers>","</servers></settings>"))
		log.Infof("New maven settings file was created: %s", filename)
	} else if !info.IsDir() {

		// search for xml path
		items := []string{mavenSettings, mavenSettingsServers}
		xpath := FindXPathInFile(filename, items)
		if xpath.IsEmpty() {
			log.Warnf("The file '%s' does not have maven settings structure.\n", filename)
			return
		}

		// update xml for the xpath /settings/servers
		servers := xpath.items[mavenSettingsServers]
		if servers != nil {
			ReplaceTextInFile(filename, createMavenSettingsServer(id, username, password, "", ""), servers.end(), servers.end())
		} else {
			// update xml for the xpath /settings
			settings := xpath.items[mavenSettings]
			if settings != nil {
				ReplaceTextInFile(filename, createMavenSettingsServer(id, username, password, "<servers>", "</servers>"), settings.end(), settings.end())
			} else {
				log.Warnf("The maven settings file '%s' does not have %s or %s\n", filename, mavenSettings, mavenSettingsServers)
				return
			}
		}
		log.Infof("The maven settings file was updated: %s", filename)
	} else {
		log.Errorf("The maven settings file %s is not valid file.", filename)
	}
}

func createMavenSettingsServer(id, username, password, prefix, suffix string) string {
	return prefix + "<server><id>" + id+ "</id><username>"+username+"</username><password>"+password+"</password></server>" + suffix
}



