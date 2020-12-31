package cmd

import (
	"fmt"

	"github.com/lorislab/samo/project"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

var (
	projectVersionCmd = &cobra.Command{
		Use:   "version",
		Short: "Show the project version",
		Long:  `Tasks to show the project version`,
		Run: func(cmd *cobra.Command, args []string) {
			op, p := readProjectOptions()
			versions := project.CreateVersions(p, op.Versions, op.HashLength, op.BuildNumberLength, op.BuildNumberPrefix)
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
			versions := project.CreateVersions(p, options.Versions, options.HashLength, options.BuildNumberLength, options.BuildNumberPrefix)
			if !versions.IsUnique() {
				log.WithFields(log.Fields{
					"versions": versions.List(),
					"custom":   versions.Custom(),
				}).Fatal("No unique version set!")
			}

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
