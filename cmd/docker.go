package cmd

import (
	"github.com/lorislab/samo/internal"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

const dockerConfigDefault string = "~/.docker/config.json"
const dockerConfigGithub string = "/home/runner/.docker/config.json"

func init() {
	dockerCmd.AddCommand(dockerConfigCmd)
	addFlag(dockerConfigCmd, "docker-config", "i", "", "docker configuration value")
	addFlag(dockerConfigCmd, "docker-config-file", "o", dockerConfigDefault, "docker client configuration file.\n Github default: '"+dockerConfigGithub+"'\n")
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
			if options.ConfigFile == dockerConfigDefault {
				if isGitHub() {
					options.ConfigFile = dockerConfigGithub
				}
			}

			// create docker config.json
			if len(options.Config) > 0 {
				internal.WriteToFile(options.ConfigFile, options.Config)
				log.Infof("New docker configuration file was created: %s", options.ConfigFile)
			}
		},
		TraverseChildren: true,
	}
)
