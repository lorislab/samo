package cmd

import (
	"io"
	"os"
	"strings"

	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var (
	// Used for flags.
	cfgFile string
	v       string
	cmd     *cobra.Command
)

// Execute executes the root command.
func Execute(version BuildVersion) {
	bv = version

	log.SetFormatter(&log.TextFormatter{
		DisableTimestamp: true,
	})
	err := cmd.Execute()
	if err != nil {
		log.Fatal(err)
	}
}

func init() {

	cmd = &cobra.Command{
		Use:   "samo",
		Short: "samo build and release tool",
		Long:  `Samo is semantic version release utility for git project.`,
		PersistentPreRunE: func(cmd *cobra.Command, args []string) error {
			if err := setUpLogs(os.Stdout, v); err != nil {
				return err
			}
			return nil
		},
		TraverseChildren: true,
	}

	cobra.OnInitialize(initConfig)

	cmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.samo.yaml)")
	cmd.PersistentFlags().StringVarP(&v, "verbosity", "v", log.InfoLevel.String(), "Log level (debug, info, warn, error, fatal, panic")

	addChildCmd(cmd, createVersionCmd())
	addChildCmd(cmd, createProjectCmd())
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			log.Fatal(err)
		}

		// Search confing in current directory
		viper.AddConfigPath(".")
		// Search config in home directory with name ".samo" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".samo")
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.SetEnvPrefix("SAMO")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		log.WithField("file", viper.ConfigFileUsed()).Debug("Using config")
	}
}

func setUpLogs(out io.Writer, level string) error {
	log.SetOutput(out)
	lvl, err := log.ParseLevel(level)
	if err != nil {
		return err
	}
	log.SetLevel(lvl)
	return nil
}
