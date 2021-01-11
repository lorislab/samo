# samo

Samo is a tool to help with a project release.

[![License](https://img.shields.io/github/license/lorislab/samo?style=for-the-badge&logo=apache)](https://www.apache.org/licenses/LICENSE-2.0)
[![GitHub Workflow Status (branch)](https://img.shields.io/github/workflow/status/lorislab/samo/build/master?logo=github&style=for-the-badge)](https://github.com/lorislab/samo/actions?query=workflow%3Abuild)
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
* `samo project patch` - create patch branch for the release

Version types:
* `version` - current project version
* `build` - build version base on the template `<project_version>-<template>`.   
Default template: `<project_version>-rc<git_count>.<git_hash>`
* `hash` - hash version `<project_version>-<git_hash>`
* `branch` - branch version `<project_version>-<git_branch>`
* `release` - release/final version of the project
* `latest` - latest verison for the docker image
* `dev` - local developer version for the docker image without repository
* `<custom_version>` - custom version which could will be use


For example to build docker image of the project only with a build-version tag:
```shell
❯ samo project docker build --version build
INFO Build docker image                     image= tags="[release-notes:3.1.0-rc001.g2a5d0e7c5fcb]"
INFO docker build --pull -t release-notes:3.1.0-rc001.g2a5d0e7c5fcb -f src/main/docker/Dockerfile . 
INFO Docker build done!                     image=release-notes
```