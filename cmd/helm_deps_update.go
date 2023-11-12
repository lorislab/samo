package cmd

import (
	"github.com/lorislab/samo/log"
	"github.com/spf13/cobra"
	"helm.sh/helm/v3/pkg/chart"
)

type helmDepsUpdateFlags struct {
	Helm           helmFlags `mapstructure:",squash"`
	DepsName       string    `mapstructure:"helm-deps-name"`
	DepsVersion    string    `mapstructure:"helm-deps-version"`
	DepsRepository string    `mapstructure:"helm-deps-repo"`
	DepsAlias      string    `mapstructure:"helm-deps-alias"`
	DepsCondition  string    `mapstructure:"helm-deps-condition"`
}

func createHelmDepsUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "deps-update",
		Short: "Update or create helm chart dependency",
		Long:  `Update the Helm chart dependency in Chart.yaml or create a new one if it doesn't exist`,
		Run: func(cmd *cobra.Command, args []string) {
			flags := helmDepsUpdateFlags{}
			readOptions(&flags)
			project := loadProject(flags.Helm.Project)
			helmDepsUpdate(project, flags)
		},
		TraverseChildren: true,
	}

	addStringFlagReq(cmd, "helm-deps-name", "", "", `name of the helm dependency in the Chart.yaml`)
	addStringFlag(cmd, "helm-deps-version", "", "", `version of the helm dependency in the Chart.yaml`)
	addStringFlag(cmd, "helm-deps-repo", "", "", `repository of the helm dependency in the Chart.yaml`)
	addStringFlag(cmd, "helm-deps-alias", "", "", `alias of the helm dependency in the Chart.yaml`)
	addStringFlag(cmd, "helm-deps-condition", "", "", `condition of the helm dependency in the Chart.yaml`)
	return cmd
}

func helmDepsUpdate(project *Project, flags helmDepsUpdateFlags) {

	c := loadChart(project, flags.Helm)

	notFound := true
	update := false
	for _, d := range c.Metadata.Dependencies {
		if flags.DepsName == d.Name {
			notFound = false
			log.Info("Update dependency", log.F("name", flags.DepsName))
			if len(flags.DepsVersion) > 0 && flags.DepsVersion != d.Version {
				log.Info("Update version", log.F("old", d.Version).F("new", flags.DepsVersion))
				d.Version = flags.DepsVersion
				update = true
			}
			if len(flags.DepsRepository) > 0 && flags.DepsRepository != d.Repository {
				log.Info("Update repository", log.F("old", d.Repository).F("new", flags.DepsRepository))
				d.Repository = flags.DepsRepository
				update = true
			}
			if len(flags.DepsAlias) > 0 && flags.DepsAlias != d.Alias {
				log.Info("Update alias", log.F("old", d.Alias).F("new", flags.DepsAlias))
				d.Alias = flags.DepsAlias
				update = true
			}
			if len(flags.DepsCondition) > 0 && flags.DepsCondition != d.Condition {
				log.Info("Update condition", log.F("old", d.Condition).F("new", flags.DepsCondition))
				d.Condition = flags.DepsCondition
				update = true
			}
		}
	}
	if notFound {
		log.Info("Add new dependency", log.F("name", flags.DepsName))

		deps := new(chart.Dependency)
		deps.Name = flags.DepsName
		if len(flags.DepsVersion) > 0 {
			deps.Version = flags.DepsVersion
		}
		if len(flags.DepsCondition) > 0 {
			deps.Condition = flags.DepsCondition
		}
		if len(flags.DepsAlias) > 0 {
			deps.Alias = flags.DepsAlias
		}
		if len(flags.DepsRepository) > 0 {
			deps.Repository = flags.DepsRepository
		}
		c.Metadata.Dependencies = append(c.Metadata.Dependencies, deps)

		update = true
	}

	if update {
		saveChart(project, flags.Helm, c)
	} else {
		log.Info("No changes found.")
	}

}
