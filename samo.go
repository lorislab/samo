package main

import (
	"fmt"
	"strings"

	"github.com/lorislab/samo/cmd"
	homedir "github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	goVersion "go.hein.dev/go-version"
)

var (
	shortened = false
	version   = "dev"
	commit    = "none"
	date      = "unknown"
	output    = "json"
	verbose   bool
	cfgFile   string
	rootCmd   = &cobra.Command{
		Use:   "samo",
		Short: "samo release utility",
		Long:  "Samo is semantic version release utility for maven, git, docker and helm chart",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			if verbose {
				log.SetLevel(log.DebugLevel)
			}
		},
	}
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Version will output the current build information",
		Long:  ``,
		Run: func(_ *cobra.Command, _ []string) {
			resp := goVersion.FuncWithOutput(shortened, version, commit, date, output)
			fmt.Print(resp)
			return
		},
	}
)

// Execute executes the root command.
func main() {
	log.SetFormatter(&log.TextFormatter{
		DisableTimestamp: true,
	})
	err := rootCmd.Execute()
	if err != nil {
		log.Panic(err)
	}
}

func init() {
	versionCmd.Flags().BoolVarP(&shortened, "short", "s", false, "Print just the version number.")
	versionCmd.Flags().StringVarP(&output, "output", "o", "json", "Output format. One of 'yaml' or 'json'.")
	rootCmd.AddCommand(versionCmd)
	cmd.Main(rootCmd)

	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.samo.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&verbose, "verbose", "v", false, "verbose output")
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			log.Panic(err)
		}

		viper.AddConfigPath(home)
		viper.SetConfigName(".samo")
	}

	viper.SetEnvKeyReplacer(strings.NewReplacer("-", "_"))
	viper.SetEnvPrefix("SAMO")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
