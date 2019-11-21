# samo

Samo is a tool to help with a project release.

[![License](https://img.shields.io/github/license/lorislab/semver-release-maven-plugin?style=for-the-badge&logo=apache)](https://www.apache.org/licenses/LICENSE-2.0)
[![GitHub tag (latest SemVer)](https://img.shields.io/github/v/tag/lorislab/samo?logo=github&style=for-the-badge)](https://github.com/lorislab/samo/releases/latest)

## Commands
```bash
Usage: samo [-hvV] [COMMAND]
  -h, --help      Show this help message and exit.
  -v, --verbose   the verbose output
  -V, --version   Print version information and exit.
Commands:
  maven   Maven version commands
  create  Create project commands
  docker  Docker version commands
  helm    Helm commands
  git     Git commands
  help    Displays help information about the specified command

```

### Release process of project

Create new release run
```bash
mvn semver-release:release-create
```

Create new patch branch run
```bash
mvn semver-release:patch-create -DpatchVersion=X.X.0
```
