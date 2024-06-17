package cmd

import (
	"github.com/lorislab/samo/log"
	"github.com/lorislab/samo/tools"
	"github.com/spf13/cobra"
	"os"
)

type helmLockUpdateFlags struct {
	Helm       helmFlags `mapstructure:",squash"`
	KeepCharts bool      `mapstructure:"helm-keep-charts"`
}

func createHelmLockUpdateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "lock-update",
		Short: "Update Chart.lock file",
		Long:  `Update the Helm chart Chart.lock or create a new one if it doesn't exist`,
		Run: func(cmd *cobra.Command, args []string) {
			flags := helmLockUpdateFlags{}
			readOptions(&flags)
			project := loadProject(flags.Helm.Project)
			helmLockUpdate(project, flags)
		},
		TraverseChildren: true,
	}

	addBoolFlag(cmd, "helm-clean-charts", "", false, "keep chart directory after Chart.lock update")
	return cmd
}

func helmLockUpdate(project *Project, flags helmLockUpdateFlags) {

	// add repo from chart dependencies
	if flags.Helm.AddRepoDeps {
		chart := loadChart(project, flags.Helm)
		helmAddRepoDeps(chart)
	}

	dir := helmDir(project, flags.Helm)

	// update helm Chart.lock
	tools.ExecCmd("helm", "dependency", "update", dir)

	// keep chart directory?
	if flags.KeepCharts {
		return
	}

	charts := dir + "/charts"
	if _, err := os.Stat(charts); !os.IsNotExist(err) {
		log.Debug("Clean helm dependencies directory", log.F("dir", charts))
		err := os.RemoveAll(charts)
		if err != nil {
			log.Panic("error delete directory", log.F("output", charts).E(err))
		}
	}

}
