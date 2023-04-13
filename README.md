# samo

Samo is a tool to help with a project release.

[![License](https://img.shields.io/github/license/lorislab/samo?style=for-the-badge&logo=apache)](https://www.apache.org/licenses/LICENSE-2.0)
[![GitHub Workflow Status (branch)](https://img.shields.io/github/actions/workflow/status/lorislab/samo/master.yaml?logo=github&style=for-the-badge)](https://github.com/lorislab/samo/actions?query=workflow%3Abuild)
[![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/lorislab/samo?sort=semver&logo=github&style=for-the-badge)](https://github.com/lorislab/samo/releases/latest)


## Commands

```shell
samo help
```
The main commands:
* `samo project name` - name of the project
* `samo project version` - versions of the project
* `samo project docker` - project docker build,push,release
* `samo project helm` - project helm build,push,release
* `samo project release` - release project
* `samo project patch` - create patch branch


For example to build docker image of the project only with a build-version tag:
```shell
❯ samo project docker build
INFO Build docker image                     image= tags="[release-notes:3.1.0-rc.1]"
INFO docker build --pull -t release-notes:3.1.0-rc.1 -f src/main/docker/Dockerfile . 
INFO Docker build done!                     image=release-notes
```

## Development

### Local build
```
go install
samo version
{"Version":"dev","Commit":"none","Date":"unknown"}
```

### Local docker build
```
go build
docker build -t samo .
``` 

### Test release packages
```
goreleaser release --snapshot --clean
```
