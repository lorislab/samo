package cmd

import (
	"fmt"

	"github.com/lorislab/samo/project"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

type versionFlags struct {
	Project     commonFlags `mapstructure:",squash"`
	OutputValue bool        `mapstructure:"value-only"`
}

var (
	projectVersionCmd = &cobra.Command{
		Use:   "version",
		Short: "Show the project version",
		Long:  `Tasks to show the project version`,
		Run: func(cmd *cobra.Command, args []string) {
			op, p := readVersionOptions()
			versions := project.CreateVersions(p, op.Project.Versions, op.Project.HashLength, op.Project.BuildNumberLength, op.Project.BuildNumber)
			if !versions.IsEmpty() {
				for k, v := range versions.Versions() {
					if op.OutputValue {
						fmt.Printf("%s\n", v)
					} else {
						fmt.Printf("%7s: %s\n", k, v)
					}
				}
			}
			if versions.IsCustom() {
				for _, v := range versions.Custom() {
					if op.OutputValue {
						fmt.Printf("%s\n", v)
					} else {
						fmt.Printf("%7s: %s\n", "custom", v)
					}
				}
			}
		},
		TraverseChildren: true,
	}
	setVersionCmd = &cobra.Command{
		Use:   "set",
		Short: "Set the version to the project",
		Long:  `Change the version of the project to the new version`,
		Run: func(cmd *cobra.Command, args []string) {
			options, p := readProjectOptions()
			versions := project.CreateVersions(p, options.Project.Versions, options.Project.HashLength, options.Project.BuildNumberLength, options.Project.BuildNumber)
			versions.CheckUnique()

			version := p.Version()
			p.SetVersion(versions.Unique())

			log.WithFields(log.Fields{
				"file": p.Filename(),
				"old":  version,
				"new":  versions.Unique(),
			}).Info("Change the version of the project to the new version")
		},
		TraverseChildren: true,
	}
)

func init() {
	addChildCmd(projectCmd, projectVersionCmd)
	addBoolFlag(projectVersionCmd, "value-only", "", false, "write only the value to the console")

	addChildCmd(projectVersionCmd, setVersionCmd)
}

func readVersionOptions() (versionFlags, project.Project) {
	options := versionFlags{}
	err := viper.Unmarshal(&options)
	if err != nil {
		panic(err)
	}
	log.WithField("options", options).Debug("Load version options")
	return options, loadProject(options.Project.File, project.Type(options.Project.Type))
}
