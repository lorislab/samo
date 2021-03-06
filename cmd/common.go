package cmd

import (
	"strings"

	"github.com/lorislab/samo/git"
	"github.com/lorislab/samo/maven"
	"github.com/lorislab/samo/npm"
	"github.com/lorislab/samo/project"
	"github.com/lorislab/samo/tools"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func init() {
	initRoot()
	initProject()
	initVersion()
	initName()
	initDocker()
	initHelm()
}

func createVersions(p project.Project, op commonFlags) project.Versions {
	return createVersionsFrom(p, op, op.Versions)
}

func createVersionsFrom(p project.Project, op commonFlags, versions []string) project.Versions {
	return project.CreateVersions(p, versions, op.HashLength, op.BuildNumberLength, op.BuildNumber, op.FirstVersion, op.ReleaseMajor, op.PatchBranchRegex)
}

func readOptions(options interface{}) {
	err := viper.Unmarshal(options)
	if err != nil {
		log.Panic(err)
	}
	log.WithField("options", options).Debug("Load options")
}

func loadProject(p commonFlags) project.Project {

	result := findProject(p)
	if result != nil {
		log.WithFields(log.Fields{
			"type": result.Type(),
			"file": result.Filename(),
		}).Debug("Find project")
		return result
	}

	// failed loading the poject
	log.WithFields(log.Fields{
		"type": p.Type,
		"file": p.File,
	}).Fatal("Could to find project file. Please specified the type --type.")
	return nil
}

func findProject(p commonFlags) project.Project {

	projectType := project.Type(p.Type)

	// find the project type
	if len(projectType) > 0 {
		switch projectType {
		case project.Maven:
			return maven.Load(p.File)
		case project.Npm:
			return npm.Load(p.File)
		case project.Git:
			return git.Load(p.File, p.FirstVersion, p.PatchBranchRegex, p.ReleaseMajor)
		}
	}

	// priority 1 maven
	project := maven.Load("")
	if project != nil {
		return project
	}
	// priority 2 npm
	project = npm.Load("")
	if project != nil {
		return project
	}
	// priority 3 git
	project = git.Load("", p.FirstVersion, p.PatchBranchRegex, p.ReleaseMajor)
	if project != nil {
		return project
	}

	return nil
}

func addChildCmd(parent, child *cobra.Command) {
	parent.AddCommand(child)
	child.Flags().AddFlagSet(parent.Flags())
}

func addFlagRef(command *cobra.Command, flag *pflag.Flag) {
	command.Flags().AddFlag(flag)
}

func addSliceFlag(command *cobra.Command, name, shorthand string, value []string, usage string) *pflag.Flag {
	command.Flags().StringSliceP(name, shorthand, value, usage)
	return addViper(command, name)
}

func addPersistentFlag(command *cobra.Command, name, shorthand string, value string, usage string) *pflag.Flag {
	command.PersistentFlags().StringP(name, shorthand, value, usage)
	return addPersistentViper(command, name)
}

func addFlag(command *cobra.Command, name, shorthand string, value string, usage string) *pflag.Flag {
	command.Flags().StringP(name, shorthand, value, usage)
	return addViper(command, name)
}

func addIntFlag(command *cobra.Command, name, shorthand string, value int, usage string) *pflag.Flag {
	command.Flags().IntP(name, shorthand, value, usage)
	return addViper(command, name)
}

func addBoolFlag(command *cobra.Command, name, shorthand string, value bool, usage string) *pflag.Flag {
	command.Flags().BoolP(name, shorthand, value, usage)
	return addViper(command, name)
}

func addFlagReq(command *cobra.Command, name, shorthand string, value string, usage string) *pflag.Flag {
	f := addFlag(command, name, shorthand, value, usage)
	markReq(command, name)
	return f
}

func markReq(command *cobra.Command, name string) {
	err := command.MarkFlagRequired(name)
	if err != nil {
		log.WithField("name", name).Panic(err)
	}
}

func addViper(command *cobra.Command, name string) *pflag.Flag {
	f := command.Flags().Lookup(name)
	err := viper.BindPFlag(name, f)
	if err != nil {
		log.WithField("name", name).Panic(err)
	}
	return f
}

func addPersistentViper(command *cobra.Command, name string) *pflag.Flag {
	f := command.PersistentFlags().Lookup(name)
	err := viper.BindPFlag(name, f)
	if err != nil {
		log.WithField("name", name).Panic(err)
	}
	return f
}

func removeCurrentTag(skip bool) {
	if skip {
		log.Debug("Skip remove tag from HEAD")
		return
	}

	// find all tags for the current commit
	list := tools.ExecCmdOutput("git", "--no-pager", "tag", "--points-at", "HEAD")
	if len(list) <= 0 {
		log.Debug("No tag found on current commit")
		return
	}
	// could be multiple tags
	tags := strings.Split(list, "\n")
	log.WithField("tags", tags).Info("Remove git tags for current commit")

	// delete the local tags
	var cmd []string
	cmd = append(cmd, "tag", "-d")
	cmd = append(cmd, tags...)
	tools.Git(cmd...)
}
