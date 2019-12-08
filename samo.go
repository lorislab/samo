package main

import (
	"fmt"

	"github.com/lorislab/samo/cmd"
	homedir "github.com/mitchellh/go-homedir"
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
	// Verbose command line flag
	Verbose bool
	cfgFile string
	rootCmd = &cobra.Command{
		Use:   "samo",
		Short: "samo semantic version release utility",
		Long:  "Simple semantic version release utility for maven, docker and helm chart",
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
	rootCmd.Execute()
}

func init() {
	versionCmd.Flags().BoolVarP(&shortened, "short", "s", false, "Print just the version number.")
	versionCmd.Flags().StringVarP(&output, "output", "o", "json", "Output format. One of 'yaml' or 'json'.")
	rootCmd.AddCommand(versionCmd)
	cmd.Main(rootCmd)

	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&cfgFile, "config", "", "config file (default is $HOME/.samo.yaml)")
	rootCmd.PersistentFlags().BoolVarP(&Verbose, "verbose", "v", false, "verbose output")
	viper.SetDefault("author", "Andrej Petras andrej@lorislab.org")
}

func initConfig() {
	if cfgFile != "" {
		// Use config file from the flag.
		viper.SetConfigFile(cfgFile)
	} else {
		// Find home directory.
		home, err := homedir.Dir()
		if err != nil {
			panic(err)
		}

		// Search config in home directory with name ".cobra" (without extension).
		viper.AddConfigPath(home)
		viper.SetConfigName(".cobra")
	}

	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err == nil {
		fmt.Println("Using config file:", viper.ConfigFileUsed())
	}
}
