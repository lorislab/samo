package cmd

import (
	"strings"

	"github.com/lorislab/samo/log"
	"github.com/mitchellh/go-homedir"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Used for flags.
	cfgFile string
	v       string
	rootCmd *cobra.Command
)

// Execute executes the root command.
func Execute(version BuildVersion) {
	bv = version

	err := rootCmd.Execute()
	if err != nil {
		log.Fatal("error execute command", log.E(err))
	}
}

func init() {

	rootCmd = &cobra.Command{
		Use:   "samo",
		Short: "samo build and release tool",
		Long:  `Samo is semantic version release utility for git project.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			log.SetLevel(v)
			return nil
		},
		TraverseChildren: true,
	}

	cobra.OnInitialize(initConfig)

	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is .samo.yaml or $HOME/.samo.yaml)")
	rootCmd.PersistentFlags().StringVarP(&v, "verbosity", "v", log.DefaultLevel(), "Log level (debug, info, warn, error, fatal, panic)")

	addChildCmd(rootCmd, createVersionCmd())
	addChildCmd(rootCmd, createProjectCmd())
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			log.Fatal("error read home dir", log.E(err))
		}

		// Search config in current directory
		viper.AddConfigPath(".")
		// Search config in home directory with name ".samo" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".samo")
		viper.SetConfigType("yaml")
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.SetEnvPrefix("SAMO")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		if _, ok := err.(viper.ConfigFileNotFoundError); ok {
			log.Info("Configuration file not found", log.F("file", viper.ConfigFileUsed()), log.E(err))
		} else {
			log.Info("Using config", log.F("file", viper.ConfigFileUsed()))
		}
	}

}
