package cmd

import (
	"github.com/lorislab/samo/git"
	"github.com/lorislab/samo/maven"
	"github.com/lorislab/samo/npm"
	"github.com/lorislab/samo/project"
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
	return project.CreateVersions(p, versions, op.HashLength, op.BuildNumberLength, op.BuildNumber)
}

func readOptions(options interface{}) {
	err := viper.Unmarshal(options)
	if err != nil {
		log.Panic(err)
	}
	log.WithField("options", options).Debug("Load options")
}

func loadProject(p commonFlags) project.Project {

	file := p.File
	projectType := project.Type(p.Type)

	result := findProject(file, projectType)
	if result != nil {
		log.WithFields(log.Fields{
			"type": result.Type(),
			"file": result.Filename(),
		}).Debug("Find project")
		return result
	}

	// failed loading the poject
	log.WithFields(log.Fields{
		"type": projectType,
		"file": file,
	}).Fatal("Could to find project file. Please specified the type --type.")
	return nil
}

func findProject(file string, projectType project.Type) project.Project {

	// find the project type
	if len(projectType) > 0 {
		switch projectType {
		case project.Maven:
			return maven.Load(file)
		case project.Npm:
			return npm.Load(file)
		case project.Git:
			return git.Load(file)
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
	project = git.Load("")
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

func addFlagRequired(command *cobra.Command, name, shorthand string, value string, usage string) *pflag.Flag {
	f := addFlag(command, name, shorthand, value, usage)
	err := command.MarkFlagRequired(name)
	if err != nil {
		log.WithField("name", name).Panic(err)
	}
	return f
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
