# samo

Samo is a tool to help with a project release.

[![License](https://img.shields.io/github/license/lorislab/semver-release-maven-plugin?style=for-the-badge&logo=apache)](https://www.apache.org/licenses/LICENSE-2.0)
[![GitHub tag (latest SemVer)](https://img.shields.io/github/v/tag/lorislab/samo?logo=github&style=for-the-badge)](https://github.com/lorislab/samo/releases/latest)

## Commands
```bash
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

### Release process of project

Create new release run
```bash
samo git create-release
```

Create new patch branch run
```bash
samo git create-patch v0.1.0
```
