package config

import (
	"fmt"
	"sync"
	"time"

	"github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
)

const (
	ConfigGrafanaImage              = "grafana.image.url"
	ConfigGrafanaImageTag           = "grafana.image.tag"
	ConfigPluginsInitContainerImage = "grafana.plugins.init.container.image.url"
	ConfigPluginsInitContainerTag   = "grafana.plugins.init.container.image.tag"
	ConfigPodLabelValue             = "grafana.pod.label"
	ConfigOperatorNamespace         = "grafana.operator.namespace"
	ConfigDashboardLabelSelector    = "grafana.dashboard.selector"
	ConfigGrafanaPluginsUpdated     = "grafana.plugins.updated"
	ConfigOpenshift                 = "mode.openshift"
	GrafanaImage                    = "quay.io/openshift/origin-grafana"
	GrafanaVersion                  = "4.2"
	GrafanaConfigMapName            = "grafana-config"
	GrafanaConfigFileName           = "grafana.ini"
	GrafanaProvidersConfigMapName   = "grafana-providers"
	GrafanaDatasourcesConfigMapName = "grafana-datasources"
	GrafanaDashboardsConfigMapName  = "grafana-dashboards"
	GrafanaServiceAccountName       = "grafana-serviceaccount"
	GrafanaDeploymentName           = "grafana-deployment"
	GrafanaRouteName                = "grafana-route"
	GrafanaIngressName              = "grafana-ingress"
	GrafanaServiceName              = "grafana-service"
	GrafanaDataPath                 = "/var/lib/grafana"
	GrafanaLogsPath                 = "/var/log/grafana"
	GrafanaPluginsPath              = "/var/lib/grafana/plugins"
	GrafanaProvisioningPath         = "/etc/grafana/provisioning"
	PluginsInitContainerImage       = "quay.io/integreatly/grafana_plugins_init"
	PluginsInitContainerTag         = "0.0.2"
	PluginsEnvVar                   = "GRAFANA_PLUGINS"
	PluginsUrl                      = "https://grafana.com/api/plugins/%s/versions/%s"
	PluginsMinAge                   = 5
	InitContainerName               = "grafana-plugins-init"
	ResourceFinalizerName           = "grafana.cleanup"
	RequeueDelay                    = time.Second * 15
	PodLabelDefaultValue            = "grafana"
	DefaultServiceType              = "ClusterIP"
	DefaultLogLevel                 = "info"
	SecretsMountDir                 = "/etc/grafana-secrets/"
	ConfigMapsMountDir              = "/etc/grafana-configmaps/"
)

type ControllerConfig struct {
	*sync.Mutex
	Values  map[string]interface{}
	Plugins map[string]v1alpha1.PluginList
}

var instance *ControllerConfig
var once sync.Once

func GetControllerConfig() *ControllerConfig {
	once.Do(func() {
		instance = &ControllerConfig{
			Mutex:   &sync.Mutex{},
			Values:  map[string]interface{}{},
			Plugins: map[string]v1alpha1.PluginList{},
		}
	})
	return instance
}

func (c *ControllerConfig) GetDashboardId(dashboard *v1alpha1.GrafanaDashboard) string {
	return fmt.Sprintf("%v/%v", dashboard.Namespace, dashboard.Spec.Name)
}

func (c *ControllerConfig) GetPluginsFor(dashboard *v1alpha1.GrafanaDashboard) v1alpha1.PluginList {
	c.Lock()
	defer c.Unlock()
	return c.Plugins[c.GetDashboardId(dashboard)]
}

func (c *ControllerConfig) SetPluginsFor(dashboard *v1alpha1.GrafanaDashboard) {
	c.Lock()
	defer c.Unlock()
	id := c.GetDashboardId(dashboard)
	c.Plugins[id] = dashboard.Spec.Plugins
	c.AddConfigItem(ConfigGrafanaPluginsUpdated, time.Now())
}

func (c *ControllerConfig) RemovePluginsFor(dashboard *v1alpha1.GrafanaDashboard) {
	c.Lock()
	defer c.Unlock()
	id := c.GetDashboardId(dashboard)
	if _, ok := c.Plugins[id]; ok {
		delete(c.Plugins, id)
		c.AddConfigItem(ConfigGrafanaPluginsUpdated, time.Now())
	}
}

func (c *ControllerConfig) AddConfigItem(key string, value interface{}) {
	c.Lock()
	defer c.Unlock()
	if key != "" && value != nil && value != "" {
		c.Values[key] = value
	}
}

func (c *ControllerConfig) RemoveConfigItem(key string) {
	c.Lock()
	defer c.Unlock()
	if _, ok := c.Values[key]; ok {
		delete(c.Values, key)
	}
}

func (c *ControllerConfig) GetConfigItem(key string, defaultValue interface{}) interface{} {
	c.Lock()
	defer c.Unlock()
	if c.HasConfigItem(key) {
		return c.Values[key]
	}
	return defaultValue
}

func (c *ControllerConfig) GetConfigString(key, defaultValue string) string {
	c.Lock()
	defer c.Unlock()
	if c.HasConfigItem(key) {
		return c.Values[key].(string)
	}
	return defaultValue
}

func (c *ControllerConfig) GetConfigBool(key string, defaultValue bool) bool {
	c.Lock()
	defer c.Unlock()
	if c.HasConfigItem(key) {
		return c.Values[key].(bool)
	}
	return defaultValue
}

func (c *ControllerConfig) GetConfigTimestamp(key string, defaultValue time.Time) time.Time {
	c.Lock()
	defer c.Unlock()
	if c.HasConfigItem(key) {
		return c.Values[key].(time.Time)
	}
	return defaultValue
}

func (c *ControllerConfig) HasConfigItem(key string) bool {
	_, ok := c.Values[key]
	return ok
}
