package cmd

import (
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

func addChildCmd(parent, child *cobra.Command) {
	parent.AddCommand(child)
	child.Flags().AddFlagSet(parent.Flags())
}

func addStringFlag(command *cobra.Command, name, shorthand string, value string, usage string) *pflag.Flag {
	command.Flags().StringP(name, shorthand, value, usage)
	return addViper(command, name)
}

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
