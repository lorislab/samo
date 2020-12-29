# samo

Samo is a tool to help with a project release.

[![License](https://img.shields.io/github/license/lorislab/semver-release-maven-plugin?style=for-the-badge&logo=apache)](https://www.apache.org/licenses/LICENSE-2.0)
[![GitHub Workflow Status (branch)](https://img.shields.io/github/workflow/status/lorislab/samo/build/master?logo=github&style=for-the-badge)](https://github.com/lorislab/samo/actions?query=workflow%3Abuild)
[![GitHub release (latest SemVer)](https://img.shields.io/github/v/release/lorislab/samo?sort=semver&logo=github&style=for-the-badge)](https://github.com/lorislab/samo/releases/latest)


## Commands

```shell
samo help
```

For example to build docker image of the project only with a build-version tag:
```shell
❯ samo project docker-build --docker-build-tags build-version
INFO Build docker image                        image= tags="[release-notes:3.1.0-rc001.g2a5d0e7c5fcb]"
INFO docker build --pull -t release-notes:3.1.0-rc001.g2a5d0e7c5fcb -f src/main/docker/Dockerfile . 
INFO Docker build done!                        image=release-notes
```