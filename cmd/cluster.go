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
	addBoolFlag(clusterSyncCmd, "force-upgrade", "", false, "force upgrade for installed application in the cluster")

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
	ForceUpgrade   bool     `mapstructure:"force-upgrade"`
}

type declarationApp struct {
	Name      string     `yaml:"name"`
	Namespace string     `yaml:"namespace"`
	Tags      []string   `yaml:"tags"`
	Helm      helmConfig `yaml:"helm"`
}

type helmConfig struct {
	Chart       string      `yaml:"chart"`
	Repo        string      `yaml:"repo"`
	Version     string      `yaml:"version"`
	Values      interface{} `yaml:"values"`
	ValuesFiles []string    `yaml:"files"`
}
type cluster struct {
	Cluster struct {
		Name      string `yaml:"name"`
		Context   string `yaml:"context"`
		Namespace string `yaml:"namespace"`
	} `yaml:"cluster"`
	Helm struct {
		Repo string `yaml:"repo"`
	} `yaml:"helm"`
	Apps []declarationApp `yaml:"apps"`
}

type helmSearchResult struct {
	Name        string `yaml:"name"`
	Description string `yaml:"description"`
	Version     string `yaml:"version"`
	AppVersion  string `yaml:"app_version"`
}

type helmListResult struct {
	Name       string `yaml:"name"`
	Status     string `yaml:"status"`
	Revision   string `yaml:"revision"`
	AppVersion string `yaml:"app_version"`
	Chart      string `yaml:"chart"`
	Namespace  string `yaml:"namespace"`
	Updated    string `yaml:"updated"`
}

type clusterAction int

const (
	nothing clusterAction = iota
	notfound
	install
	upgrade
	downgrade
)

var clusterActionStr = []string{
	"",
	"",
	"install",
	"upgrade",
	"downgrade",
}

type application struct {
	Namespace      string
	Declaration    declarationApp
	CurrentVersion *semver.Version
	NextVersion    *semver.Version
	Action         clusterAction
	Chart          string
	RepoChart      string
	Cluster        *helmListResult
}

func (a application) Status() string {
	if a.Cluster != nil {
		return a.Cluster.Status
	}
	return ""
}

func (a application) ActionStr() string {
	return clusterActionStr[a.Action]
}

func (a application) NextVersionStr() string {
	if a.NextVersion == nil {
		return ""
	}
	return a.NextVersion.String()
}

func (a application) CurrentVersionStr() string {
	if a.CurrentVersion == nil {
		return ""
	}
	return a.CurrentVersion.String()
}

var (
	clusterCmd = &cobra.Command{
		Use:              "cluster",
		Short:            "cluster operation",
		Long:             `cluster operation`,
		TraverseChildren: true,
	}
	clusterInfoCmd = &cobra.Command{
		Use:   "info",
		Short: "Info of the cluster",
		Long:  `Info of the cluster`,
		Run: func(cmd *cobra.Command, args []string) {
			options, cluster := readClusterOptions()
			log.Infof("cluster %s: %s %s", cmd.Use, options.ConfigFile, cluster)
		},
		TraverseChildren: true,
	}
	clusterSyncCmd = &cobra.Command{
		Use:   "sync",
		Short: "Sync applications in the cluster -install, upgrade or downgrade",
		Long:  `Sync applications in the cluster - install, upgrade or downgrade`,
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
				case nothing:
					if options.ForceUpgrade {
						go helmUpgrade(app, &wg)
					}
				case install:
					go helmInstall(app, &wg)
				case upgrade:
					go helmUpgrade(app, &wg)
				case downgrade:
					go helmDowngrade(app, &wg)
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
					go helmUninstall(app, &wg)
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
				table.AddRow(app.Declaration.Name, app.Namespace, app.Chart, app.Declaration.Helm.Version, app.CurrentVersionStr(), app.NextVersionStr(), app.Status(), app.ActionStr())
			}
			fmt.Println(table)
		},
		TraverseChildren: true,
	}
)

func helmInstall(app application, wg *sync.WaitGroup) {

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

func helmDowngrade(app application, wg *sync.WaitGroup) {

	defer wg.Done()
	var uninstallSteps = []string{"waiting", "helmUninstall", "helmInstall", "finished"}
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

func helmUpgrade(app application, wg *sync.WaitGroup) {

	defer wg.Done()
	var uninstallSteps = []string{"waiting", "helmUpgrade", "finished"}
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

func helmUninstall(app application, wg *sync.WaitGroup) {

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

func loadApplications(cluster cluster, tags, apps []string) map[string]application {
	log.Debugf("Load application info for the cluster filter tags %s apps %s", tags, apps)

	// load all helm releases in the repository
	log.Infof("Load apps releases from helm repo...")
	helmSearchResults := readHelmSearchResult()

	// load all helm releases in the cluster
	log.Infof("Load apps releases from cluster...")
	clusterReleases := clusterReleases()

	result := make(map[string]application)

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

		chart := app.Name
		if len(app.Helm.Chart) > 0 {
			chart = app.Helm.Chart
		}
		repoChart := chart
		if len(app.Helm.Repo) > 0 {
			repoChart = app.Helm.Repo + "/" + repoChart
		} else {
			if len(cluster.Helm.Repo) > 0 {
				repoChart = cluster.Helm.Repo + "/" + repoChart
			}
		}

		var nextVersion *semver.Version
		var currentVersion *semver.Version
		clusterVersion, exists := clusterReleases[id]
		if exists {
			currentVersion, _ = semver.NewVersion(strings.TrimPrefix(clusterVersion.Chart, chart+"-"))
		}

		repoVersions, exists := helmSearchResults[repoChart]
		if exists {
			nextVersion = findLatestBaseOnTheRules(repoVersions, app.Helm.Version)
		}
		action := nothing
		if nextVersion != nil {
			if currentVersion == nil {
				action = install
			} else {
				if currentVersion.LessThan(nextVersion) {
					action = upgrade
				} else if currentVersion.GreaterThan(nextVersion) {
					action = downgrade
				}
			}
		} else {
			action = notfound
		}
		result[app.Name+"-"+namespace] = application{
			Namespace:      namespace,
			Declaration:    app,
			CurrentVersion: currentVersion,
			NextVersion:    nextVersion,
			Action:         action,
			RepoChart:      repoChart,
			Chart:          chart,
			Cluster:        &clusterVersion,
		}
	}
	return result
}

func findLatestBaseOnTheRules(items []helmSearchResult, rule string) *semver.Version {
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

func clusterReleases() map[string]helmListResult {
	list := make(map[string]helmListResult)

	data := internal.ExecCmdOutput("helm", "list", "--output", "yaml", "--all-namespaces")
	var helmListResult []helmListResult
	err := yaml.Unmarshal([]byte(data), &helmListResult)
	if err != nil {
		panic(err)
	}
	for _, item := range helmListResult {
		list[item.Namespace+"-"+item.Name] = item
	}
	return list
}

func readHelmSearchResult() map[string][]helmSearchResult {
	data := internal.ExecCmdOutput("helm", "search", "repo", "--devel", "-l", "--output", "yaml")
	var items []helmSearchResult
	err := yaml.Unmarshal([]byte(data), &items)
	if err != nil {
		panic(err)
	}
	result := make(map[string][]helmSearchResult)
	for _, item := range items {
		i, e := result[item.Name]
		if !e {
			i = make([]helmSearchResult, 0)
		}
		i = append(i, item)
		result[item.Name] = i
	}
	return result
}

func readClusterOptions() (clusterFlags, cluster) {
	options := clusterFlags{}
	err := viper.Unmarshal(&options)
	if err != nil {
		panic(err)
	}
	log.Debug(options)
	return options, readClusterConfig(options)
}

func readClusterConfig(options clusterFlags) cluster {
	cluster := cluster{}
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
