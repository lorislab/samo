package cmd

import (
	"bufio"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"os"
	"path/filepath"
)

const dockerConfigDefault string = "~/.docker/config.json"
const dockerConfigGithub string = "/home/runner/.docker/config.json"

func init() {
	dockerCmd.AddCommand(dockerConfigCmd)
	addFlag(dockerConfigCmd, "docker-config", "e", "", "The docker configuration value")
	addFlag(dockerConfigCmd, "docker-config-file", "j", dockerConfigDefault, "Docker client configuration file.\n Github default: '"+dockerConfigGithub+"'\n")
}

type dockerFlags struct {
	Config     string `mapstructure:"docker-config"`
	ConfigFile string `mapstructure:"docker-config-file"`
}

var (
	dockerCmd = &cobra.Command{
		Use:              "docker",
		Short:            "Docker operation",
		Long:             `Docker operation`,
		TraverseChildren: true,
	}
	dockerConfigCmd = &cobra.Command{
		Use:   "config",
		Short: "Config the docker client",
		Long:  `Config the docker client`,
		Run: func(cmd *cobra.Command, args []string) {

			options := dockerFlags{}
			err := viper.Unmarshal(&options)
			if err != nil {
				panic(err)
			}

			// check default for the pipeline
			if options.Config == dockerConfigDefault {
				if isGitHub() {
					options.Config = dockerConfigGithub
				}
			}

			// create docker config.json
			if len(options.Config) > 0 {

				dir := filepath.Dir(options.ConfigFile)
				err := os.MkdirAll(dir, os.ModePerm)
				if err != nil {
					panic(err)
				}

				file, err := os.Create(options.ConfigFile)
				if err != nil {
					panic(err)
				}
				w := bufio.NewWriter(file)
				_, err = w.WriteString(options.Config)
				if err != nil {
					panic(err)
				}
				err = w.Flush()
				if err != nil {
					panic(err)
				}
			}
		},
		TraverseChildren: true,
	}
)
