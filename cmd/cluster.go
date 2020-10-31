package cmd

import (
	"fmt"
	"github.com/Masterminds/semver"
	"io/ioutil"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/gosuri/uitable"

	"github.com/lorislab/samo/internal"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v2"
)

func init() {
	clusterCmd.AddCommand(clusterInfoCmd)
	configFile := addFlag(clusterInfoCmd, "config-file", "", "cluster.yaml", "clusterConfig client configuration file.")
	app := addStringSliceFlag(clusterInfoCmd, "app-name", "a", []string{}, "application name for the action")
	tags := addStringSliceFlag(clusterInfoCmd, "tags", "", []string{}, "comma separated list of tags")
	helmUpdate := addBoolFlag(mvnCreateReleaseCmd, "helm-repo-update", "", false, "helm repo update")

	clusterCmd.AddCommand(clusterCreateCmd)
	addFlagRef(clusterCreateCmd, configFile)

	clusterCmd.AddCommand(clusterStartCmd)
	addFlagRef(clusterStartCmd, configFile)

	clusterCmd.AddCommand(clusterStopCmd)
	addFlagRef(clusterStopCmd, configFile)

	clusterCmd.AddCommand(clusterDeleteCmd)
	addFlagRef(clusterDeleteCmd, configFile)

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
	addBoolFlag(clusterSyncCmd, "force-appUpgrade", "", false, "force appUpgrade for installed application in the clusterConfig")
	addBoolFlag(mvnCreateReleaseCmd, "no-wait", "", false, "helm repo update")

	clusterCmd.AddCommand(clusterUninstallCmd)
	addFlagRef(clusterUninstallCmd, configFile)
	addFlagRef(clusterUninstallCmd, app)
	addFlagRef(clusterUninstallCmd, tags)
}

type chart struct {
	Version string `yaml:"version"`
}

type clusterFlags struct {
	ConfigFile     string   `mapstructure:"config-file"`
	HelmRepoUpdate bool     `mapstructure:"helm-repo-update"`
	Apps           []string `mapstructure:"app-name"`
	Tags           []string `mapstructure:"tags"`
	ForceUpgrade   bool     `mapstructure:"force-appUpgrade"`
	NoWait         bool     `mapstructure:"no-wait"`
}

type declarationApp struct {
	Namespace string     `yaml:"namespace"`
	Tags      []string   `yaml:"tags"`
	Helm      helmConfig `yaml:"helm"`
	Priority  int        `yaml:"priority"`
	Ingress   struct {
		Enabled bool   `yaml:"enabled"`
		Host    bool   `yaml:"host"`
		Path    string `yaml:"path"`
	}
}

type helmConfig struct {
	Chart       string      `yaml:"chart"`
	Repo        string      `yaml:"repo"`
	Version     string      `yaml:"version"`
	Values      interface{} `yaml:"values"`
	ValuesFiles []string    `yaml:"files"`
}

type clusterConfig struct {
	Cluster struct {
		Name      string `yaml:"name"`
		Namespace string `yaml:"namespace"`
		K3d       struct {
			Registry string `yaml:"registry"`
			Agent    string `yaml:"agent"`
			Port     string `yaml:"port"`
		} `yaml:"k3d"`
	} `yaml:"cluster"`
	Apps map[string]declarationApp `yaml:"apps"`
}

func (c clusterConfig) namespace(appName string) string {
	app, exists := c.Apps[appName]
	if exists {
		if len(app.Namespace) > 0 {
			return app.Namespace
		}
	}
	return c.Cluster.Namespace
}

func (c clusterConfig) id(appName string) string {
	return id(c.namespace(appName), appName)
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

func (h helmListResult) id() string {
	return id(h.Namespace, h.Name)
}

func id(namespace, name string) string {
	return namespace + "-" + name
}

type appAction int

const (
	appNothing appAction = iota
	appNotfound
	appInstall
	appUpgrade
	appDowngrade
	appUninstall
)

type helmAction int

const (
	install helmAction = iota
	uninstall
	upgrade
)

var helmActionStr = []string{
	"install",
	"uninstall",
	"upgrade",
}

var appActionStr = []string{
	"",
	"",
	"install",
	"upgrade",
	"downgrade",
	"uninstall",
}

type application struct {
	Namespace      string
	AppName        string
	Declaration    declarationApp
	CurrentVersion *semver.Version
	NextVersion    *semver.Version
	Action         appAction
	Chart          string
	ChartRepo      string
	Cluster        *helmListResult
}

func (a *application) status() string {
	if a.Cluster != nil {
		return a.Cluster.Status
	}
	return ""
}

func (a *application) actionStr() string {
	return appActionStr[a.Action]
}

func (a *application) nextVersionStr() string {
	if a.NextVersion == nil {
		return ""
	}
	return a.NextVersion.String()
}

func (a *application) CurrentVersionStr() string {
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
			//options, cluster := readClusterOptions()
			log.Infof("Cluster info!!")
		},
		TraverseChildren: true,
	}
	clusterSyncCmd = &cobra.Command{
		Use:   "sync",
		Short: "Sync applications in the cluster -appInstall, appUpgrade or appDowngrade",
		Long:  `Sync applications in the cluster - appInstall, appUpgrade or appDowngrade`,
		Run: func(cmd *cobra.Command, args []string) {
			options, cluster := readClusterOptions()
			apps, keys := loadApplications(cluster, options.Tags, options.Apps)

			count := 0
			sum := 0

			for _, key := range keys {
				var wg sync.WaitGroup
				count = 0

				for _, app := range apps[key] {
					count++
					sum++
					wg.Add(1)
					go helmCmd(options, app, &wg, cmd.Use)
				}
				wg.Wait()
				log.Infof("Sync apps finished priority: %d. Count: %d Sum: %d", key, count, sum)
			}
			log.Infof("Sync apps finished. Sum: %d", sum)
		},
		TraverseChildren: true,
	}
	clusterUninstallCmd = &cobra.Command{
		Use:   "uninstall",
		Short: "Uninstall applications in the cluster",
		Long:  `Uninstall applications in the cluster`,
		Run: func(cmd *cobra.Command, args []string) {
			options, cluster := readClusterOptions()
			apps, keys := loadApplications(cluster, options.Tags, options.Apps)
			sort.Sort(sort.Reverse(sort.IntSlice(keys)))

			count := 0

			for _, key := range keys {
				var wg sync.WaitGroup
				for _, app := range apps[key] {
					if app.CurrentVersion != nil {
						count++
						wg.Add(1)
						app.Action = appUninstall
						go helmCmd(options, app, &wg, cmd.Use)
					}
				}
				wg.Wait()
			}

			log.Infof("Uninstall apps finished. Count: %d", count)
		},
		TraverseChildren: true,
	}
	clusterStatusCmd = &cobra.Command{
		Use:   "status",
		Short: "Status of the applications in the cluster",
		Long:  `Status of the applications in the cluster`,
		Run: func(cmd *cobra.Command, args []string) {
			options, cluster := readClusterOptions()

			if options.HelmRepoUpdate {
				log.Infof("Update helm repositories...")
				internal.ExecCmdOutput("helm", "repo", "update")
			}

			apps, keys := loadApplications(cluster, options.Tags, options.Apps)

			table := uitable.New()
			table.MaxColWidth = 50
			table.AddRow("PRIORITY", "NAME", "NAMESPACE", "CHART", "RULE", "CLUSTER", "REPOSITORY", "STATUS", "ACTION")
			for _, key := range keys {
				for _, app := range apps[key] {
					table.AddRow(key, app.AppName, app.Namespace, app.Chart, app.Declaration.Helm.Version, app.CurrentVersionStr(), app.nextVersionStr(), app.status(), app.actionStr())
				}
			}
			fmt.Println(table)
		},
		TraverseChildren: true,
	}
	clusterCreateCmd = &cobra.Command{
		Use:   "create",
		Short: "Create the cluster instance",
		Long:  `Create the cluster instance`,
		Run: func(cmd *cobra.Command, args []string) {
			_, config := readClusterOptions()
			k3dCreateCluster(config)
		},
		TraverseChildren: true,
	}
	clusterDeleteCmd = &cobra.Command{
		Use:   "delete",
		Short: "Delete the cluster instance",
		Long:  `Delete the cluster instance`,
		Run: func(cmd *cobra.Command, args []string) {
			_, config := readClusterOptions()
			execK3dCmd(config, "delete")
		},
		TraverseChildren: true,
	}
	clusterStopCmd = &cobra.Command{
		Use:   "stop",
		Short: "Stop the cluster instance",
		Long:  `Stop the cluster instance`,
		Run: func(cmd *cobra.Command, args []string) {
			_, config := readClusterOptions()
			execK3dCmd(config, "stop")
		},
		TraverseChildren: true,
	}
	clusterStartCmd = &cobra.Command{
		Use:   "start",
		Short: "Start the cluster instance",
		Long:  `Start the cluster instance`,
		Run: func(cmd *cobra.Command, args []string) {
			_, config := readClusterOptions()
			execK3dCmd(config, "start")
		},
		TraverseChildren: true,
	}
)

func k3dCreateCluster(config clusterConfig) {
	var command []string
	command = append(command, "cluster", "create", config.Cluster.Name)
	if len(config.Cluster.K3d.Port) > 0 {
		command = append(command, "-p", config.Cluster.K3d.Port+":80@loadbalancer")
	}
	if len(config.Cluster.K3d.Registry) > 0 {
		registryFile, err := filepath.Abs(config.Cluster.K3d.Registry)
		if err != nil {
			panic(err)
		}
		command = append(command, "--volume", registryFile+":/etc/rancher/k3s/registries.yaml")
	}
	agents := "2"
	if len(config.Cluster.K3d.Agent) > 0 {
		agents = config.Cluster.K3d.Agent
	}
	command = append(command, "--agents", agents)
	internal.ExecCmd("k3d", command...)
}

func execK3dCmd(config clusterConfig, cmd string) {
	internal.ExecCmd("k3d", "cluster", cmd, config.Cluster.Name)
}

func helmCmd(options clusterFlags, app *application, wg *sync.WaitGroup, cmd string) {
	defer wg.Done()
	log.Infof("[%s] Start execution %s....", app.AppName, cmd)
	switch app.Action {
	case appNothing:
		if options.ForceUpgrade {
			log.Infof("[%s] execute helm command `%s`....", app.AppName, helmActionStr[upgrade])
			helmExecuteCmd(options, app, upgrade)
		}
	case appInstall:
		log.Infof("[%s] execute helm command `%s`....", app.AppName, helmActionStr[install])
		helmExecuteCmd(options, app, install)
	case appUpgrade:
		log.Infof("[%s] execute helm command `%s`....", app.AppName, helmActionStr[upgrade])
		helmExecuteCmd(options, app, upgrade)
	case appDowngrade:
		// uninstall new version
		log.Infof("[%s] execute helm command `%s`....", app.AppName, helmActionStr[uninstall])
		helmExecuteSimpleCmd(app, uninstall)
		// appInstall appInstall old version
		log.Infof("[%s] execute helm command `%s`....", app.AppName, helmActionStr[install])
		helmExecuteCmd(options, app, install)

		if app.Declaration.Ingress.Enabled {
			// execute kubectl and apply ingress for k3d

		}
	case appUninstall:
		// uninstall new version
		log.Infof("[%s] execute helm command %s....", app.AppName, helmActionStr[uninstall])
		helmExecuteSimpleCmd(app, uninstall)
	}
	log.Infof("[%s] finished %s", app.AppName, cmd)
}

func helmExecuteCmd(options clusterFlags, app *application, action helmAction) {
	var command []string
	command = append(command, helmActionStr[action], app.AppName, app.ChartRepo, "--version", app.nextVersionStr(), "-n", app.Namespace)
	if !options.NoWait {
		command = append(command, "--wait")
	}
	command = append(command, "--create-namespace")
	command = append(command, "-o", "yaml")
	if len(app.Declaration.Helm.ValuesFiles) > 0 {
		for _, file := range app.Declaration.Helm.ValuesFiles {
			command = append(command, "-f", file)
		}
	}

	if app.Declaration.Helm.Values != nil {
		values := app.Declaration.Helm.Values
		valuesYaml, err := yaml.Marshal(&values)
		if err != nil {
			panic(err)
		}
		tmpFile, err := ioutil.TempFile(os.TempDir(), "samo-cluster-"+app.AppName+"-*.yaml")
		if err != nil {
			panic(err)
		}
		defer func() {
			if err := os.Remove(tmpFile.Name()); err != nil {
				panic(err)
			}
		}()

		if _, err = tmpFile.Write(valuesYaml); err != nil {
			panic(err)
		}
		if err := tmpFile.Close(); err != nil {
			panic(err)
		}
		log.Debugf("Create helm values file %s for application %s", tmpFile.Name(), app.AppName)
		command = append(command, "-f", tmpFile.Name())
	}
	internal.ExecCmdOutput("helm", command...)
}

func helmExecuteSimpleCmd(app *application, action helmAction) {
	internal.ExecCmdOutput("helm", helmActionStr[action], app.AppName, "-n", app.Namespace)
}

func loadApplications(cluster clusterConfig, tags, apps []string) (map[int][]*application, []int) {
	log.Debugf("Load application info for the clusterConfig filter tags %s apps %s", tags, apps)

	// load all helm releases in the repository
	log.Infof("Load apps releases from helm repo...")
	helmSearchResults := readHelmSearchResult()

	// load all helm releases in the clusterConfig
	log.Infof("Load apps releases from clusterConfig...")
	clusterReleases := clusterReleases()

	result := make(map[int][]*application)

	type void struct{}
	var member void

	index := make(map[int]void)
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
	for appName, app := range cluster.Apps {
		if len(mApps) > 0 {
			_, exists := mApps[appName]
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

		// id of the application
		id := cluster.id(appName)

		chart := app.Helm.Chart
		chartRepo := chart
		local := false
		if strings.HasPrefix(app.Helm.Repo, "alias:") {
			chartRepo = strings.TrimPrefix(app.Helm.Repo, "alias:") + "/" + chartRepo
		} else if strings.HasPrefix(app.Helm.Repo, "file://") {
			chartRepo = strings.TrimPrefix(app.Helm.Repo, "file://")
			local = true
		}

		var nextVersion *semver.Version
		var currentVersion *semver.Version
		clusterVersion, exists := clusterReleases[id]
		if exists {
			currentVersion, _ = semver.NewVersion(strings.TrimPrefix(clusterVersion.Chart, chart+"-"))
		}

		repoVersions, exists := helmSearchResults[chartRepo]
		if exists {
			nextVersion = findLatestBaseOnTheRules(repoVersions, app.Helm.Version)
		} else {
			if local {
				nextVersion = versionFromLocalChart(chartRepo, app.Helm.Version)
			}
		}
		action := appNothing
		if nextVersion != nil {
			if currentVersion == nil {
				action = appInstall
			} else {
				if currentVersion.LessThan(nextVersion) {
					action = appUpgrade
				} else if currentVersion.GreaterThan(nextVersion) {
					action = appDowngrade
				}
			}
		} else {
			action = appNotfound
		}

		tmp := &application{
			AppName:        appName,
			Namespace:      cluster.namespace(appName),
			Declaration:    app,
			CurrentVersion: currentVersion,
			NextVersion:    nextVersion,
			Action:         action,
			Chart:          chart,
			ChartRepo:      chartRepo,
			Cluster:        &clusterVersion,
		}
		list, exists := result[app.Priority]
		if !exists {
			var a []*application
			list = a
		}
		list = append(list, tmp)
		result[app.Priority] = list
		index[app.Priority] = void{}
	}
	keys := make([]int, 0, len(index))
	for key := range index {
		keys = append(keys, key)
	}
	sort.Ints(keys)
	return result, keys
}

func versionFromLocalChart(repo string, rule string) *semver.Version {
	chart := chart{}

	data, err := ioutil.ReadFile(repo + "/Chart.yaml")
	if err != nil {
		panic(err)
	}
	err = yaml.Unmarshal(data, &chart)
	if err != nil {
		panic(err)
	}
	ver, err := semver.NewVersion(chart.Version)
	if err != nil {
		panic(err)
	}

	c, err := semver.NewConstraint(rule)
	if err != nil {
		panic(err)
	}

	if c.Check(ver) {
		return ver
	}
	return nil
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
	data := internal.ExecCmdOutput("helm", "list", "--output", "yaml", "--all-namespaces", "--all")
	var helmListResult []helmListResult
	err := yaml.Unmarshal([]byte(data), &helmListResult)
	if err != nil {
		panic(err)
	}
	for _, item := range helmListResult {
		list[item.id()] = item
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

func readClusterOptions() (clusterFlags, clusterConfig) {
	options := clusterFlags{}
	err := viper.Unmarshal(&options)
	if err != nil {
		panic(err)
	}
	log.Debug(options)
	return options, readClusterConfig(options)
}

func readClusterConfig(options clusterFlags) clusterConfig {
	clusterConfig := clusterConfig{}
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
	err = yaml.Unmarshal(yamlFile, &clusterConfig)
	if err != nil {
		panic(err)
	}
	return clusterConfig
}
