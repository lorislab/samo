package project

import (
	"regexp"
	"strings"

	"github.com/Masterminds/semver"
	"github.com/lorislab/samo/tools"
	log "github.com/sirupsen/logrus"
)

const (
	VerVersion = "version"
	VerBuild   = "build"
	VerRelease = "release"
	VerHash    = "hash"
	VerLatest  = "latest"
	VerBranch  = "branch"
	VerDev     = "dev"
	Latest     = "latest"
)

func VersionsList() []string {
	return []string{VerVersion, VerBuild, VerRelease, VerHash, VerLatest, VerBranch, VerDev}
}

func VersionsText() string {
	return "[" + strings.Join(VersionsList(), ",") + "]"
}

type Versions struct {
	custom            []string
	HashLength        int
	BuildNumberLength int
	BuildNumber       string
	PatchBranch       bool
	Major             bool
	versions          map[string]string
	semVer            *semver.Version
}

// NextReleaseVersion creates next release version
func (v Versions) NextReleaseVersion() semver.Version {
	return *nextReleaseVersion(v.semVer, v.Major, v.PatchBranch)
}

func (v Versions) SemVer() *semver.Version {
	return v.semVer
}

func (v Versions) CheckUnique() {
	if v.IsUnique() {
		return
	}
	log.WithFields(log.Fields{
		"versions": v.List(),
		"custom":   v.Custom(),
	}).Fatal("No unique version set!")
}

func (v Versions) IsUnique() bool {
	return len(v.versions)+len(v.custom) == 1
}

func (v Versions) Unique() string {
	if len(v.versions) == 1 {
		for _, v := range v.versions {
			return v
		}
	}
	if len(v.custom) == 1 {
		return v.custom[0]
	}
	return ""
}

func (v Versions) IsCustom() bool {
	return len(v.custom) > 0
}

func (v Versions) Custom() []string {
	return v.custom
}

func (v Versions) IsEmpty() bool {
	return len(v.versions) <= 0
}

func (v Versions) List() []string {
	tmp := []string{}
	for _, v := range v.versions {
		tmp = append(tmp, v)
	}
	return tmp
}

func (v Versions) Versions() map[string]string {
	return v.versions
}

func (v Versions) IsVersion() bool {
	return v.is(VerVersion)
}

func (v Versions) Version() string {
	return v.versions[VerVersion]
}

func (v Versions) IsBuildVersion() bool {
	return v.is(VerBuild)
}

func (v Versions) BuildVersion() string {
	return v.versions[VerBuild]
}

func (v Versions) IsReleaseVersion() bool {
	return v.is(VerRelease)
}

func (v Versions) ReleaseVersion() string {
	return v.versions[VerRelease]
}

func (v Versions) IsHashVersion() bool {
	return v.is(VerHash)
}

func (v Versions) HashVersion() string {
	return v.versions[VerHash]
}

func (v Versions) IsLatestVersion() bool {
	return v.is(VerLatest)
}

func (v Versions) LatestVersion() string {
	return v.versions[VerLatest]
}

func (v Versions) IsBranchVersion() bool {
	return v.is(VerBranch)
}

func (v Versions) BranchVersion() string {
	return v.versions[VerBranch]
}

func (v Versions) IsDevVersion() bool {
	return v.is(VerDev)
}

func (v Versions) DevVersion() string {
	return v.versions[VerDev]
}

func (v Versions) is(key string) bool {
	return len(v.versions[key]) > 0
}

func CreateVersions(project Project, versions []string, hashLength, buildNumberLength int, buildNumber, firstVer string,
	major bool, patchBranchRegex string) Versions {

	patchBranh := false
	if len(patchBranchRegex) > 0 {
		branch := tools.GitBranch()

		match, e := regexp.MatchString(patchBranchRegex, branch)
		if e != nil {
			log.WithFields(log.Fields{
				"branch": branch,
				"regex":  patchBranchRegex,
				"error":  e,
			}).Fatal("Error parsing patch branch regex")
		}
		log.WithFields(log.Fields{
			"branch": branch,
			"regex":  patchBranchRegex,
			"match":  match,
		}).Debug("Check patch branch regex")
		patchBranh = match
	}

	semVer := SemVer(project.Version())
	ver, custom := createVersions(semVer, versions, hashLength, buildNumberLength, buildNumber, firstVer, major, patchBranh)
	return Versions{
		custom:            custom,
		HashLength:        hashLength,
		BuildNumberLength: buildNumberLength,
		BuildNumber:       buildNumber,
		versions:          ver,
		semVer:            semVer,
		Major:             major,
		PatchBranch:       patchBranh,
	}
}

func createVersions(semVer *semver.Version, versions []string, hashLength, buildNumberLength int, buildNumber, firstVer string,
	major, patchBranch bool) (map[string]string, []string) {
	result := make(map[string]string)
	custom := []string{}

	// if no version return
	if versions == nil {
		return result, custom
	}

	types := make(map[string]bool)
	for _, n := range versions {
		types[n] = true
	}

	// project version
	if types[VerVersion] {
		result[VerVersion] = semVer.String()
	}
	// build version
	if types[VerBuild] {
		result[VerBuild] = buildVersion(semVer, hashLength, buildNumberLength, buildNumber, firstVer).String()
	}
	// release version
	if types[VerRelease] {
		result[VerRelease] = releaseVersion(semVer).String()
	}
	// latest version
	if types[VerLatest] {
		result[VerLatest] = Latest
	}
	// hash version
	if types[VerHash] {
		result[VerHash] = hashVersion(semVer, hashLength, firstVer).String()
	}
	// branch tag
	if types[VerBranch] {
		result[VerBranch] = tools.GitBranch()
	}
	// latest version
	if types[VerDev] {
		result[VerDev] = Latest
	}

	// find custom versions
	for _, k := range VersionsList() {
		delete(types, k)
	}
	for k := range types {
		custom = append(custom, k)
	}
	return result, custom
}

// releaseVersion the release version
func releaseVersion(tmp *semver.Version) *semver.Version {
	version, err := tmp.SetPrerelease("")
	if err != nil {
		log.WithField("version", tmp.String()).Fatal("Error remove pre-release label from version")
	}
	version, err = version.SetMetadata("")
	if err != nil {
		log.WithField("version", tmp.String()).Fatal("Error remove metadata label from version")
	}
	return &version
}

func hashVersion(tmp *semver.Version, hashLength int, firstVer string) *semver.Version {
	_, _, hash := tools.GitCommit(hashLength, firstVer)
	ver, err := tmp.SetPrerelease(hash)
	if err != nil {
		log.WithFields(log.Fields{
			"hash":    hash,
			"version": tmp.String(),
		}).Fatal("Error set hash pre-release label to the version")
	}
	return &ver
}

type buildVersionData struct {
	Number string
	Hash   string
	Count  string
}

// BuildVersion build version of the project
func buildVersion(semVer *semver.Version, hashLength, length int, template, firstVer string) *semver.Version {
	tmp := releaseVersion(semVer)
	_, count, hash := tools.GitCommit(hashLength, firstVer)

	// number
	number := tools.Lpad(count, "0", length)

	data := buildVersionData{
		Hash:   hash,
		Number: number,
		Count:  count,
	}
	pre := createText(data, template)

	result, err := tmp.SetPrerelease(pre)
	if err != nil {
		log.WithFields(log.Fields{
			"prerelease": pre,
			"data":       data,
			"version":    tmp.String(),
		}).Fatal("Error set pre-release")
	}
	return &result
}

// SemVer create SemVer version for the project
func SemVer(version string) *semver.Version {
	result, e := semver.NewVersion(version)
	if e != nil {
		log.WithField("version", version).Fatal("Version value is not valid semver 2.0.")
	}
	return result
}

// NextReleaseVersion creates next release version
func nextReleaseVersion(ver *semver.Version, major, patchBranch bool) *semver.Version {
	if major {
		if ver.Patch() != 0 {
			log.WithField("version", ver.String()).Fatal("Can not created major release from the patch version!")
		}
		tmp := ver.IncMajor()
		return &tmp
	}
	if ver.Patch() != 0 || patchBranch {
		tmp := ver.IncPatch()
		return &tmp
	}
	tmp := ver.IncMinor()
	return &tmp
}
