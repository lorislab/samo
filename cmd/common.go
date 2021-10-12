package cmd

import (
	"github.com/lorislab/samo/project"
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func readOptions(options interface{}) interface{} {
	err := viper.Unmarshal(options)
	if err != nil {
		log.Panic(err)
	}
	log.WithField("options", options).Debug("Load options")
	return options
}

func loadProject(p projectFlags) *project.Project {
	return project.LoadProject(p.FirstVersion, p.VersionTemplate, p.ReleaseMajor, p.ReleasePatch)
}

func addChildCmd(parent, child *cobra.Command) {
	parent.AddCommand(child)
	child.Flags().AddFlagSet(parent.Flags())
}

// func addFlagRef(command *cobra.Command, flag *pflag.Flag) {
// 	command.Flags().AddFlag(flag)
// }

// func addSliceFlag(command *cobra.Command, name, shorthand string, value []string, usage string) *pflag.Flag {
// 	command.Flags().StringSliceP(name, shorthand, value, usage)
// 	return addViper(command, name)
// }

// func addPersistentFlag(command *cobra.Command, name, shorthand string, value string, usage string) *pflag.Flag {
// 	command.PersistentFlags().StringP(name, shorthand, value, usage)
// 	return addPersistentViper(command, name)
// }

func addStringFlag(command *cobra.Command, name, shorthand string, value string, usage string) *pflag.Flag {
	command.Flags().StringP(name, shorthand, value, usage)
	return addViper(command, name)
}

// func addIntFlag(command *cobra.Command, name, shorthand string, value int, usage string) *pflag.Flag {
// 	command.Flags().IntP(name, shorthand, value, usage)
// 	return addViper(command, name)
// }

func addBoolFlag(command *cobra.Command, name, shorthand string, value bool, usage string) *pflag.Flag {
	command.Flags().BoolP(name, shorthand, value, usage)
	return addViper(command, name)
}

func addStringFlagReq(command *cobra.Command, name, shorthand string, value string, usage string) *pflag.Flag {
	f := addStringFlag(command, name, shorthand, value, usage)
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

// func addPersistentViper(command *cobra.Command, name string) *pflag.Flag {
// 	f := command.PersistentFlags().Lookup(name)
// 	err := viper.BindPFlag(name, f)
// 	if err != nil {
// 		log.WithField("name", name).Panic(err)
// 	}
// 	return f
// }

// func removeCurrentTag(skip bool) {
// 	if skip {
// 		log.Debug("Skip remove tag from HEAD")
// 		return
// 	}

// 	// find all tags for the current commit
// 	list := tools.ExecCmdOutput("git", "--no-pager", "tag", "--points-at", "HEAD")
// 	if len(list) <= 0 {
// 		log.Debug("No tag found on current commit")
// 		return
// 	}
// 	// could be multiple tags
// 	tags := strings.Split(list, "\n")
// 	log.WithField("tags", tags).Info("Remove git tags for current commit")

// 	// delete the local tags
// 	var cmd []string
// 	cmd = append(cmd, "tag", "-d")
// 	cmd = append(cmd, tags...)
// 	tools.Git(cmd...)
// }
