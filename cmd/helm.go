package cmd

import (
	"github.com/lorislab/samo/helm"
	"github.com/lorislab/samo/project"
	"github.com/spf13/cobra"
)

type helmFlags struct {
	Project            commonFlags `mapstructure:",squash"`
	HelmInputDir       string      `mapstructure:"helm-input"`
	HelmOutputDir      string      `mapstructure:"helm-output"`
	HelmClean          bool        `mapstructure:"helm-clean"`
	HelmFilterTemplate string      `mapstructure:"helm-filter-template"`
	HelmBuildFilter    bool        `mapstructure:"helm-filter"`
	HelmUpdateChart    []string    `mapstructure:"helm-update-chart"`
	HelmUpdateValues   []string    `mapstructure:"helm-update-values"`
	HelmSkipPush       bool        `mapstructure:"helm-push-skip"`
	HelmRepository     string      `mapstructure:"helm-repo"`
	HelmRepoUsername   string      `mapstructure:"helm-repo-username"`
	HelmRepoPassword   string      `mapstructure:"helm-repo-password"`
	HelmRepositoryURL  string      `mapstructure:"helm-repo-url"`
	HelmRepositoryAdd  bool        `mapstructure:"helm-repo-add"`
}

var (
	helmCmd = &cobra.Command{
		Use:              "helm",
		Short:            "Project helm operation",
		Long:             `Tasks for the helm of the project`,
		TraverseChildren: true,
	}
	helmBuildCmd = &cobra.Command{
		Use:   "build",
		Short: "Build helm chart",
		Long:  `Helm build helm chart`,
		Run: func(cmd *cobra.Command, args []string) {
			op, p := readHelmOptions()

			versions := createVersions(p, op.Project)
			versions.CheckUnique()

			helm := helm.HelmRequest{
				Project:       p,
				Versions:      versions,
				Input:         op.HelmInputDir,
				Output:        op.HelmOutputDir,
				Clean:         op.HelmClean,
				Template:      op.HelmFilterTemplate,
				Repository:    op.HelmRepository,
				RepositoryURL: op.HelmRepositoryURL,
				AddRepo:       op.HelmRepositoryAdd,
				Username:      op.HelmRepoUsername,
				Password:      op.HelmRepoPassword,
				Filter:        op.HelmBuildFilter,
			}
			helm.Build()
		},
		TraverseChildren: true,
	}
	helmPushCmd = &cobra.Command{
		Use:   "push",
		Short: "Push helm chart",
		Long:  `Push helm chart to the helm repository`,
		Run: func(cmd *cobra.Command, args []string) {
			op, p := readHelmOptions()

			versions := createVersions(p, op.Project)
			versions.CheckUnique()

			helm := helm.HelmRequest{
				Project:       p,
				Versions:      versions,
				Output:        op.HelmOutputDir,
				SkipPush:      op.HelmSkipPush,
				Username:      op.HelmRepoUsername,
				Password:      op.HelmRepoPassword,
				RepositoryURL: op.HelmRepositoryURL,
			}
			helm.Push()
		},
		TraverseChildren: true,
	}
	helmReleaseCmd = &cobra.Command{
		Use:   "release",
		Short: "Release helm chart",
		Long:  `Download build version of the helm chart and create final version`,
		Run: func(cmd *cobra.Command, args []string) {
			op, p := readHelmOptions()
			helm := helm.HelmRequest{
				Project:      p,
				Versions:     createVersionsFrom(p, op.Project, []string{project.VerBuild, project.VerRelease}),
				Output:       op.HelmOutputDir,
				Clean:        op.HelmClean,
				ChartUpdate:  op.HelmUpdateChart,
				ValuesUpdate: op.HelmUpdateValues,
				SkipPush:     op.HelmSkipPush,
			}
			helm.Release()
		},
		TraverseChildren: true,
	}
)

func init() {
	addChildCmd(projectCmd, helmCmd)
	addFlag(helmCmd, "helm-input", "", "src/main/helm", "filter project helm chart input directory")
	addFlag(helmCmd, "helm-output", "", "target/helm", "filter project helm chart output directory")
	addBoolFlag(helmCmd, "helm-clean", "", false, "clean output directory before filter")

	addChildCmd(helmCmd, helmBuildCmd)
	addFlag(helmBuildCmd, "helm-repo", "", "", "helm repository name")
	addFlag(helmBuildCmd, "helm-repo-url", "", "", "helm repository name")
	addFlag(helmBuildCmd, "helm-repo-username", "u", "", "helm repository username")
	addFlag(helmBuildCmd, "helm-repo-password", "p", "", "helm repository password")
	addBoolFlag(helmCmd, "helm-filter", "", false, "filter helm reousrces from input to output directory")
	addFlagRequired(helmBuildCmd, "helm-filter-template", "", "maven", "Use the maven template for filter")

	addChildCmd(helmCmd, helmPushCmd)
	addBoolFlag(helmCmd, "helm-skip-push", "", false, "skip helm push")

	addChildCmd(helmCmd, helmReleaseCmd)
	addFlag(helmReleaseCmd, "helm-update-chart", "", "version={{ .Version }},appVersion={{ .Version }}", "list of key value to be replaced in the Chart.yaml")
	addFlag(helmReleaseCmd, "helm-update-values", "", "image.tag={{ .Version }}", "list of key value to be replaced in the values.yaml")

}

func readHelmOptions() (helmFlags, project.Project) {
	options := helmFlags{}
	readOptions(&options)
	return options, loadProject(options.Project)
}
