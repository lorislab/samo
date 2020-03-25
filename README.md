# samo

Samo is a tool to help with a project release.

[![License](https://img.shields.io/github/license/lorislab/semver-release-maven-plugin?style=for-the-badge&logo=apache)](https://www.apache.org/licenses/LICENSE-2.0)
[![GitHub tag (latest SemVer)](https://img.shields.io/github/v/tag/lorislab/samo?logo=github&style=for-the-badge)](https://github.com/lorislab/samo/releases/latest)

## Commands
```shell script
Samo is semantic version release utility for maven, git, docker and helm chart

Usage:
  samo [command]

Available Commands:
  docker      Docker operation
  git         Git operation
  help        Help about any command
  maven       Maven operation
  version     Version will output the current build information

Flags:
      --config string   config file (default is $HOME/.samo.yaml)
  -h, --help            help for samo
  -v, --verbose         verbose output

Use "samo [command] --help" for more information about a command.
```

## Maven project commands

```shell script
Tasks for the maven project

Usage:
  samo maven [command]

Available Commands:
  build-version       Show the maven project version to build version
  create-patch        Create patch of the release
  create-release      Create release of the current project and state
  docker-build        Build the docker image of the maven project
  docker-build-dev    Build the docker image of the maven project for local development
  docker-push         Push the docker image of the maven project
  docker-release      Release the docker image
  release-version     Show the maven project version to release
  set-build-version   Set the maven project version to build version
  set-release-version Set the maven project version to release
  settings-add-server Add the maven repository server to the settings
  version             Show the maven project version

Flags:
  -h, --help   help for maven

Global Flags:
      --config string   config file (default is $HOME/.samo.yaml)
  -v, --verbose         verbose output

Use "samo maven [command] --help" for more information about a command.
```

### Release process of project

Create new release run
```bash
samo git create-release
```

Create new patch branch run
```bash
samo git create-patch v0.1.0
```
