package cmd

import (
	"fmt"
	"github.com/Masterminds/semver/v3"
	"github.com/lorislab/samo/log"
	"github.com/spf13/cobra"
)

type helmDepsValidateFlags struct {
	Helm         helmFlags `mapstructure:",squash"`
	ValidateType string    `mapstructure:"helm-deps-validate"`
}

func createHelmDepsValidateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deps-validate",
		Short: "Validate helm chart dependencies",
		Long:  `Validate helm chart dependencies`,
		Run: func(cmd *cobra.Command, args []string) {
			flags := helmDepsValidateFlags{}
			readOptions(&flags)
			project := loadProject(flags.Helm.Project)
			helmDepsValidate(project, flags)
		},
		TraverseChildren: true,
	}

	addStringFlag(cmd, "helm-deps-validate", "", "final", `validate all dependencies version type in the Chart.yaml. Type is one of [ final | exact ]`)
	return cmd
}

func helmDepsValidate(project *Project, flags helmDepsValidateFlags) {

	chart := loadChart(project, flags.Helm)
	failed := false

	// add repo from chart dependencies
	if flags.Helm.AddRepoDeps {
		helmAddRepoDeps(chart)
	}

	log.Info("Dependencies validation", log.F("validate-type", flags.ValidateType).F("chart", chart.Name()).F("version", chart.Metadata.Version))

	fmt.Println("Dependencies:")
	for _, d := range chart.Metadata.Dependencies {

		fmt.Println(d.Name + "\t" + d.Version)

		ver, err := semver.NewVersion(d.Version)
		if err != nil {
			log.Debug("error parse dependency version", log.E(err).F("version", d.Version).F("name", d.Name))
			failed = true
		} else {
			switch flags.ValidateType {
			case "final":
				if ver.Patch() != 0 || len(ver.Prerelease()) > 0 {
					log.Debug("Dependency does not have final version", log.F("name", d.Name).F("version", ver.Original()))
					failed = true
				} else {
					log.Debug("Dependency final version", log.F("name", d.Name).F("version", ver.Original()))
				}
			case "exact":
				//
			}
		}
	}

	if failed {
		log.Fatal("One or more dependencies version are not valid! Validation: '" + flags.ValidateType + "'")
	} else {
		log.Info("All dependencies version are valid. Validation: '" + flags.ValidateType + "'")
	}
}
