package cmd

import (
	"github.com/lorislab/samo/helm"
	"github.com/lorislab/samo/project"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type helmFlags struct {
	Project            commonFlags `mapstructure:",squash"`
	HelmInputDir       string      `mapstructure:"helm-input"`
	HelmOutputDir      string      `mapstructure:"helm-output"`
	HelmClean          bool        `mapstructure:"helm-clean"`
	HelmFilterTemplate string      `mapstructure:"helm-filter-template"`
	HelmBuildFilter    bool        `mapstructure:"helm-filter"`
	HelmFilterChart    []string    `mapstructure:"helm-update-chart"`
	HelmFilterValues   []string    `mapstructure:"helm-update-values"`
	HelmSkipPush       bool        `mapstructure:"helm-push-skip"`
	HelmRepository     string      `mapstructure:"helm-repo"`
	HelmRepoUsername   string      `mapstructure:"helm-repo-username"`
	HelmRepoPassword   string      `mapstructure:"helm-repo-password"`
	HelmRepositoryURL  string      `mapstructure:"helm-repo-url"`
	HelmRepositoryAdd  bool        `mapstructure:"helm-repo-add"`
	HelmUpdateVersion  bool        `mapstructure:"helm-update-version"`
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
				FilterChart:   op.HelmFilterChart,
				FilterValues:  op.HelmFilterValues,
				UpdateVersion: op.HelmUpdateVersion,
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

			if len(op.HelmRepositoryURL) <= 0 {
				log.WithField("--helm-repo-url", op.HelmRepositoryURL).Fatal("Helm repository URL is required!")
			}

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

			if len(op.HelmRepositoryURL) <= 0 {
				log.WithField("--helm-repo-url", op.HelmRepositoryURL).Fatal("Helm repository URL is required!")
			}

			helm := helm.HelmRequest{
				Project:       p,
				Versions:      createVersionsFrom(p, op.Project, []string{project.VerBuild, project.VerRelease}),
				Output:        op.HelmOutputDir,
				Clean:         op.HelmClean,
				FilterChart:   op.HelmFilterChart,
				FilterValues:  op.HelmFilterValues,
				SkipPush:      op.HelmSkipPush,
				Username:      op.HelmRepoUsername,
				Password:      op.HelmRepoPassword,
				RepositoryURL: op.HelmRepositoryURL,
			}
			helm.Release()
		},
		TraverseChildren: true,
	}
)

func initHelm() {
	addChildCmd(projectCmd, helmCmd)
	addFlag(helmCmd, "helm-input", "", "src/main/helm", "filter project helm chart input directory")
	addFlag(helmCmd, "helm-output", "", "target/helm", "filter project helm chart output directory")
	addBoolFlag(helmCmd, "helm-clean", "", false, "clean output directory before filter")

	addChildCmd(helmCmd, helmBuildCmd)
	addFlag(helmBuildCmd, "helm-repo", "", "", "helm repository name")
	hurl := addFlag(helmBuildCmd, "helm-repo-url", "", "", "helm repository URL")
	addBoolFlag(helmBuildCmd, "helm-repo-add", "", false, "add helm repository before build")
	hu := addFlag(helmBuildCmd, "helm-repo-username", "u", "", "helm repository username")
	hp := addFlag(helmBuildCmd, "helm-repo-password", "p", "", "helm repository password")
	addBoolFlag(helmBuildCmd, "helm-filter", "", false, "filter helm reousrces from input to output directory")
	addFlag(helmBuildCmd, "helm-filter-template", "", "maven", "use the maven template for filter")
	fc := addFlag(helmBuildCmd, "helm-update-chart", "", "version={{ .Version }},appVersion={{ .Version }}", "list of key value to be replaced in the Chart.yaml")
	fv := addFlag(helmBuildCmd, "helm-update-values", "", "", "list of key value to be replaced in the values.yaml Example: image.tag={{ .Version }}")
	addBoolFlag(helmBuildCmd, "helm-update-version", "", false, "update version before package")

	addChildCmd(helmCmd, helmPushCmd)
	skipPush := addBoolFlag(helmPushCmd, "helm-push-skip", "", false, "skip helm push")
	addFlagRef(helmPushCmd, hurl)
	addFlagRef(helmPushCmd, hu)
	addFlagRef(helmPushCmd, hp)

	addChildCmd(helmCmd, helmReleaseCmd)
	addFlagRef(helmReleaseCmd, hurl)
	addFlagRef(helmReleaseCmd, hu)
	addFlagRef(helmReleaseCmd, hp)
	addFlagRef(helmReleaseCmd, skipPush)
	addFlagRef(helmReleaseCmd, fc)
	addFlagRef(helmReleaseCmd, fv)

}

func readHelmOptions() (helmFlags, project.Project) {
	options := helmFlags{}
	readOptions(&options)
	return options, loadProject(options.Project)
}
