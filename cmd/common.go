package cmd

import (
	log "github.com/sirupsen/logrus"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

func addChildCmd(parent, child *cobra.Command) {
	parent.AddCommand(child)
	child.Flags().AddFlagSet(parent.Flags())
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
