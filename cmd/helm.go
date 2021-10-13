package cmd

import (
	"os"

	"github.com/lorislab/samo/tools"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type helmFlags struct {
	Project       projectFlags `mapstructure:",squash"`
	Repo          string       `mapstructure:"repo"`
	RepoUsername  string       `mapstructure:"repo-username"`
	RepoPassword  string       `mapstructure:"repo-password"`
	RepositoryURL string       `mapstructure:"repo-url"`
	Clean         bool         `mapstructure:"clean"`
	PushURL       string       `mapstructure:"push-url"`
	Dir           string       `mapstructure:"dir"`
}

func createHelmCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:              "helm",
		Short:            "Project helm operation",
		Long:             `Tasks for the helm of the project`,
		TraverseChildren: true,
	}

	addBoolFlag(cmd, "clean", "", false, "clean output directory before filter")
	addStringFlag(cmd, "dir", "", "target/helm", "filter project helm chart output directory")
	addStringFlag(cmd, "repo", "", "", "helm repository name")
	addStringFlag(cmd, "repo-url", "", "", "helm repository URL")
	addStringFlag(cmd, "repo-username", "u", "", "helm repository username")
	addStringFlag(cmd, "repo-password", "p", "", "helm repository password")
	addStringFlag(cmd, "push-url", "", "", "helm repository push URL")

	addChildCmd(cmd, createHealmBuildCmd())
	addChildCmd(cmd, createHealmPushCmd())
	addChildCmd(cmd, createHealmReleaseCmd())
	return cmd
}

func helmPackage(flags helmFlags) {
	tools.ExecCmd("helm", "package", flags.Dir)
}

func healmClean(flags helmFlags) {
	// clean output directory
	if !flags.Clean {
		log.Debug("Helm clean disabled.")
		return
	}
	if _, err := os.Stat(flags.Dir); !os.IsNotExist(err) {
		log.WithField("dir", flags.Dir).Debug("Clean directory")
		err := os.RemoveAll(flags.Dir)
		if err != nil {
			log.WithField("output", flags.Dir).Panic(err)
		}
	}
}

func healmAddRepo(flags helmFlags) {
	if len(flags.Repo) == 0 {
		return
	}

	// add repository
	var command []string
	command = append(command, "repo", "add")
	if len(flags.RepoPassword) > 0 {
		command = append(command, "--password", flags.RepoPassword)
	}
	if len(flags.RepoUsername) > 0 {
		command = append(command, "--username", flags.RepoUsername)
	}
	command = append(command, flags.Repo, flags.RepositoryURL)
	tools.ExecCmd("helm", command...)
}

func helmRepoUpdate() {
	tools.ExecCmd("helm", "repo", "update")
}

func helmPush(version string, project *Project, flags helmFlags) {

	if len(flags.PushURL) == 0 {
		log.WithFields(log.Fields{"push-url": flags.PushURL, "version": version}).Fatal("Flag --push-url is mandatory!")
	}

	// upload helm chart
	if flags.Project.SkipPush {
		log.WithFields(log.Fields{"push-url": flags.PushURL, "version": version}).Info("Skip push release version of the helm chart")
		return
	}

	filename := project.Name() + `-` + version + `.tgz`
	if !tools.Exists(filename) {
		log.WithField("helm-file", filename).Fatal("Helm package file does not exists!")
	}

	var command []string
	command = append(command, "-fis", "--show-error")
	if len(flags.RepoPassword) > 0 {
		command = append(command, "-u", flags.RepoUsername+`:`+flags.RepoPassword)
	}
	command = append(command, flags.PushURL, "--upload-file", filename)
	tools.ExecCmd("curl", command...)
}
