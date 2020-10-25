package cmd

import (
	"fmt"
	"io/ioutil"
	"os"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/gosuri/uiprogress/util/strutil"

	"github.com/Masterminds/semver"

	"github.com/gosuri/uitable"

	"github.com/gosuri/uiprogress"
	"github.com/lorislab/samo/internal"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

func init() {
	clusterCmd.AddCommand(clusterInfoCmd)
	configFile := addFlag(clusterInfoCmd, "config-file", "", "cluster.yaml", "cluster client configuration file.")
	app := addStringSliceFlag(clusterInfoCmd, "app-name", "a", []string{}, "application name for the action")
	tags := addStringSliceFlag(clusterInfoCmd, "tags", "", []string{}, "comma separated list of tags")
	helmUpdate := addBoolFlag(mvnCreateReleaseCmd, "helm-repo-update", "", false, "helm repo update")

	clusterCmd.AddCommand(clusterStatusCmd)
	addFlagRef(clusterStatusCmd, configFile)
	addFlagRef(clusterStatusCmd, app)
	addFlagRef(clusterStatusCmd, tags)
	addFlagRef(clusterStatusCmd, helmUpdate)

	clusterCmd.AddCommand(clusterSyncCmd)
	addFlagRef(clusterSyncCmd, configFile)
	addFlagRef(clusterSyncCmd, app)
	addFlagRef(clusterSyncCmd, tags)
	addFlagRef(clusterSyncCmd, helmUpdate)

	clusterCmd.AddCommand(clusterRemoveCmd)
	addFlagRef(clusterRemoveCmd, configFile)
	addFlagRef(clusterRemoveCmd, app)
	addFlagRef(clusterRemoveCmd, tags)
}

type clusterFlags struct {
	ConfigFile     string   `mapstructure:"config-file"`
	HelmRepoUpdate bool     `mapstructure:"helm-repo-update"`
	Apps           []string `mapstructure:"app-name"`
	Tags           []string `mapstructure:"tags"`
}

type DeclarationApp struct {
	Name        string      `yaml:"name"`
	Chart       string      `yaml:"chart"`
	Version     string      `yaml:"version"`
	Values      interface{} `yaml:"values"`
	ValuesFiles []string    `yaml:"files"`
	Namespace   string      `yaml:"namespace"`
	Tags        []string    `yaml:"tags"`
}

type Cluster struct {
	Helm struct {
		Alias string `yaml:"alias"`
	}
	Cluster struct {
		Name      string `yaml:"name"`
		Context   string `yaml:"context"`
		Namespace string `yaml:"namespace"`
	} `yaml:"cluster"`
	Apps []DeclarationApp `yaml:"apps"`
}

type HelmSearchResult struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Version     string `yaml:"version"`
	AppVersion  string `yaml:"app_version"`
}

type HelmListResult struct {
	Name       string `yaml:"name"`
	Status     string `yaml:"status"`
	Revision   string `yaml:"revision"`
	AppVersion string `yaml:"app_version"`
	Chart      string `yaml:"chart"`
	Namespace  string `yaml:"namespace"`
	Updated    string `yaml:"updated"`
}

type ClusterAction int

const (
	NOTHING ClusterAction = iota
	INSTALL
	UPGRADE
	DOWNGRADE
)

var clusterActionStr = []string{
	"",
	"install",
	"upgrade",
	"downgrade",
}

type Application struct {
	Namespace      string
	Declaration    DeclarationApp
	CurrentVersion *semver.Version
	NextVersion    *semver.Version
	Action         ClusterAction
	Chart          string
	Cluster        *HelmListResult
}

func (a Application) Status() string {
	if a.Cluster != nil {
		return a.Cluster.Status
	}
	return ""
}

func (a Application) ActionStr() string {
	return clusterActionStr[a.Action]
}

func (a Application) NextVersionStr() string {
	if a.NextVersion == nil {
		return ""
	}
	return a.NextVersion.String()
}

func (a Application) CurrentVersionStr() string {
	if a.CurrentVersion == nil {
		return ""
	}
	return a.CurrentVersion.String()
}

var (
	clusterCmd = &cobra.Command{
		Use:              "cluster",
		Short:            "Cluster operation",
		Long:             `Cluster operation`,
		TraverseChildren: true,
	}
	clusterInfoCmd = &cobra.Command{
		Use:   "info",
		Short: "Info of the cluster",
		Long:  `Info of the cluster`,
		Run: func(cmd *cobra.Command, args []string) {
			options, cluster := readClusterOptions()
			log.Infof("Cluster %s: %s %s", cmd.Use, options.ConfigFile, cluster)
		},
		TraverseChildren: true,
	}
	clusterSyncCmd = &cobra.Command{
		Use:   "sync",
		Short: "Sync applications in the cluster",
		Long:  `Sync applications in the cluster`,
		Run: func(cmd *cobra.Command, args []string) {
			options, cluster := readClusterOptions()
			apps := loadApplications(cluster, options.Tags, options.Apps)

			count := 0
			uiprogress.Start()
			var wg sync.WaitGroup
			for _, app := range apps {

				count++
				wg.Add(1)

				switch app.Action {
				case INSTALL:
					go install(app, &wg)
				case UPGRADE:
					go upgrade(app, &wg)
				case DOWNGRADE:
					go downgrade(app, &wg)
				}
			}
			wg.Wait()
			log.Infof("Sync apps finished. Count: %d", count)
		},
		TraverseChildren: true,
	}
	clusterRemoveCmd = &cobra.Command{
		Use:   "remove",
		Short: "Remove applications in the cluster",
		Long:  `Remove applications in the cluster`,
		Run: func(cmd *cobra.Command, args []string) {
			options, cluster := readClusterOptions()
			apps := loadApplications(cluster, options.Tags, options.Apps)

			count := 0
			uiprogress.Start()
			var wg sync.WaitGroup
			for _, app := range apps {
				if app.CurrentVersion != nil {
					count++
					wg.Add(1)
					go uninstall(app, &wg)
				}
			}
			wg.Wait()
			log.Infof("Uninstall apps finished. Count: %d", count)
		},
		TraverseChildren: true,
	}
	clusterStatusCmd = &cobra.Command{
		Use:   "status",
		Short: "Status of the cluster",
		Long:  `Status of the cluster`,
		Run: func(cmd *cobra.Command, args []string) {
			options, cluster := readClusterOptions()

			if options.HelmRepoUpdate {
				log.Infof("Update helm repositories...")
				internal.ExecCmdOutput("helm", "repo", "update")
			}

			apps := loadApplications(cluster, options.Tags, options.Apps)

			table := uitable.New()
			table.MaxColWidth = 50
			table.AddRow("NAME", "NAMESPACE", "CHART", "RULE", "CLUSTER", "REPOSITORY", "STATUS", "ACTION")
			for _, app := range apps {
				table.AddRow(app.Declaration.Name, app.Namespace, app.Chart, app.Declaration.Version, app.CurrentVersionStr(), app.NextVersionStr(), app.Status(), app.ActionStr())
			}
			fmt.Println(table)
		},
		TraverseChildren: true,
	}
)

func install(app Application, wg *sync.WaitGroup) {

	defer wg.Done()
	var uninstallSteps = []string{"waiting", "installing", "finished"}
	bar := uiprogress.AddBar(len(uninstallSteps)).PrependElapsed()
	bar.Width = 50
	bar.Incr()

	// prepend the deploy step to the bar
	bar.PrependFunc(func(b *uiprogress.Bar) string {
		return strutil.Resize(app.Declaration.Name+": "+uninstallSteps[b.Current()-1], 32)
	})

	bar.Incr()
	internal.ExecCmdOutput("helm", "install", app.Declaration.Name, app.Chart, "--version", app.NextVersionStr(), "--wait", "-n", app.Namespace)
	bar.Incr()
	// wait to refresh console
	time.Sleep(time.Millisecond * 10)
}

func downgrade(app Application, wg *sync.WaitGroup) {

	defer wg.Done()
	var uninstallSteps = []string{"waiting", "uninstall", "install", "finished"}
	bar := uiprogress.AddBar(len(uninstallSteps)).PrependElapsed()
	bar.Width = 50
	bar.Incr()

	// prepend the deploy step to the bar
	bar.PrependFunc(func(b *uiprogress.Bar) string {
		return strutil.Resize(app.Declaration.Name+": "+uninstallSteps[b.Current()-1], 32)
	})

	bar.Incr()
	internal.ExecCmdOutput("helm", "uninstall", app.Declaration.Name, "-n", app.Namespace)
	bar.Incr()
	internal.ExecCmdOutput("helm", "install", app.Declaration.Name, app.Chart, "--version", app.NextVersionStr(), "--wait", "-n", app.Namespace)
	bar.Incr()
	// wait to refresh console
	time.Sleep(time.Millisecond * 10)
}

func upgrade(app Application, wg *sync.WaitGroup) {

	defer wg.Done()
	var uninstallSteps = []string{"waiting", "upgrade", "finished"}
	bar := uiprogress.AddBar(len(uninstallSteps)).PrependElapsed()
	bar.Width = 50
	bar.Incr()

	// prepend the deploy step to the bar
	bar.PrependFunc(func(b *uiprogress.Bar) string {
		return strutil.Resize(app.Declaration.Name+": "+uninstallSteps[b.Current()-1], 32)
	})

	bar.Incr()
	internal.ExecCmdOutput("helm", "upgrade", app.Declaration.Name, app.Chart, "--version", app.NextVersionStr(), "--wait", "-n", app.Namespace)
	bar.Incr()
	// wait to refresh console
	time.Sleep(time.Millisecond * 10)
}

func uninstall(app Application, wg *sync.WaitGroup) {

	defer wg.Done()
	var uninstallSteps = []string{"waiting", "uninstalling", "finished"}
	bar := uiprogress.AddBar(len(uninstallSteps)).PrependElapsed()
	bar.Width = 50
	bar.Incr()

	// prepend the deploy step to the bar
	bar.PrependFunc(func(b *uiprogress.Bar) string {
		return strutil.Resize(app.Declaration.Name+": "+uninstallSteps[b.Current()-1], 32)
	})

	bar.Incr()
	internal.ExecCmdOutput("helm", "uninstall", app.Declaration.Name, "-n", app.Namespace)
	bar.Incr()
	// wait to refresh console
	time.Sleep(time.Millisecond * 10)
}

func loadApplications(cluster Cluster, tags, apps []string) map[string]Application {
	log.Debugf("Load application info for the cluster filter tags %s apps %s", tags, apps)

	// load all helm releases in the repository
	log.Infof("Load apps releases from helm repo...")
	helmSearchResults := readHelmSearchResult()

	// load all helm releases in the cluster
	log.Infof("Load apps releases from cluster...")
	clusterReleases := clusterReleases()

	result := make(map[string]Application)

	type void struct{}
	var member void
	mTags := make(map[string]void)
	if len(tags) > 0 {
		for _, t := range tags {
			mTags[t] = member
		}
	}
	mApps := make(map[string]void)
	if len(apps) > 0 {
		for _, t := range apps {
			mApps[t] = member
		}
	}

	// loop over all apps and check the status
	for _, app := range cluster.Apps {
		if len(mApps) > 0 {
			_, exists := mApps[app.Name]
			if !exists {
				continue
			}
		}
		if len(mTags) > 0 {
			contains := false
			for i := 0; i < len(app.Tags) && !contains; i++ {
				_, exists := mTags[app.Tags[i]]
				contains = contains || exists
			}
			if !contains {
				continue
			}
		}
		namespace := cluster.Cluster.Namespace
		if len(app.Namespace) > 0 {
			namespace = app.Namespace
		}
		id := namespace + "-" + app.Name
		var nextVersion *semver.Version
		var currentVersion *semver.Version
		clusterVersion, exists := clusterReleases[id]
		if exists {
			currentVersion, _ = semver.NewVersion(clusterVersion.AppVersion)
		}
		chart := app.Chart
		if !(len(chart) > 0) {
			if len(cluster.Helm.Alias) > 0 {
				chart = cluster.Helm.Alias + "/" + chart
			}
			chart = chart + app.Name
		} else {
			if !strings.Contains(chart, "/") {
				if len(cluster.Helm.Alias) > 0 {
					chart = cluster.Helm.Alias + "/" + chart
				}
			}
		}
		repoVersions, exists := helmSearchResults[chart]
		if exists {
			nextVersion = findLatestBaseOnTheRules(repoVersions, app.Version)
		}
		action := NOTHING
		if nextVersion != nil {
			if currentVersion == nil {
				action = INSTALL
			} else {
				if currentVersion.LessThan(nextVersion) {
					action = UPGRADE
				} else if currentVersion.GreaterThan(nextVersion) {
					action = DOWNGRADE
				}
			}
		}
		result[app.Name+"-"+namespace] = Application{
			Namespace:      namespace,
			Declaration:    app,
			CurrentVersion: currentVersion,
			NextVersion:    nextVersion,
			Action:         action,
			Chart:          chart,
			Cluster:        &clusterVersion,
		}
	}
	return result
}

func findLatestBaseOnTheRules(items []HelmSearchResult, rule string) *semver.Version {
	vs := make([]*semver.Version, len(items))
	for i, r := range items {
		v, err := semver.NewVersion(r.Version)
		if err != nil {
			panic(err)
		}
		vs[i] = v
	}
	sort.Sort(sort.Reverse(semver.Collection(vs)))

	c, err := semver.NewConstraint(rule)
	if err != nil {
		panic(err)
	}

	for _, ver := range vs {
		if c.Check(ver) {
			return ver
		}
	}
	return nil
}

func clusterReleases() map[string]HelmListResult {
	list := make(map[string]HelmListResult)

	data := internal.ExecCmdOutput("helm", "list", "--output", "yaml", "--all-namespaces")
	var helmListResult []HelmListResult
	err := yaml.Unmarshal([]byte(data), &helmListResult)
	if err != nil {
		panic(err)
	}
	for _, item := range helmListResult {
		list[item.Namespace+"-"+item.Name] = item
	}
	return list
}

func readHelmSearchResult() map[string][]HelmSearchResult {
	data := internal.ExecCmdOutput("helm", "search", "repo", "--devel", "-l", "--output", "yaml")
	var items []HelmSearchResult
	err := yaml.Unmarshal([]byte(data), &items)
	if err != nil {
		panic(err)
	}
	result := make(map[string][]HelmSearchResult)
	for _, item := range items {
		i, e := result[item.Name]
		if !e {
			i = make([]HelmSearchResult, 0)
		}
		i = append(i, item)
		result[item.Name] = i
	}
	return result
}

func readClusterOptions() (clusterFlags, Cluster) {
	options := clusterFlags{}
	err := viper.Unmarshal(&options)
	if err != nil {
		panic(err)
	}
	log.Debug(options)
	return options, readClusterConfig(options)
}

func readClusterConfig(options clusterFlags) Cluster {
	cluster := Cluster{}
	yamlFile, err := ioutil.ReadFile(options.ConfigFile)
	if err != nil {
		panic(err)
	}
	file, err := os.Open(options.ConfigFile)
	if err != nil {
		panic(err)
	}
	if file != nil {
		defer func() {
			if err := file.Close(); err != nil {
				log.Panic(err)
			}
		}()
	}
	err = yaml.Unmarshal(yamlFile, &cluster)
	if err != nil {
		panic(err)
	}
	return cluster
}
