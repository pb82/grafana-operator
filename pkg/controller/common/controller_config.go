package common

import "sync"

const (
	ConfigGrafanaImage           = "grafana.image.url"
	ConfigGrafanaImageTag        = "grafana.image.tag"
	ConfigOperatorNamespace      = "grafana.operator.namespace"
	ConfigDashboardLabelSelector = "grafana.dashboard.selector"
)

type ControllerConfig struct {
	Values map[string]interface{}
}

var instance *ControllerConfig
var once sync.Once

func GetControllerConfig() *ControllerConfig {
	once.Do(func() {
		instance = &ControllerConfig{
			Values: map[string]interface{}{},
		}
	})
	return instance
}

func (c *ControllerConfig) AddConfigItem(key string, value interface{}) {
	if key != "" && value != nil && value != "" {
		c.Values[key] = value
	}
}

func (c *ControllerConfig) GetConfigItem(key string, defaultValue interface{}) interface{} {
	if c.HasConfigItem(key) {
		return c.Values[key]
	}
	return defaultValue
}

func (c *ControllerConfig) GetConfigString(key, defaultValue string) string {
	if c.HasConfigItem(key) {
		return c.Values[key].(string)
	}
	return defaultValue
}

func (c *ControllerConfig) HasConfigItem(key string) bool {
	_, ok := c.Values[key]
	return ok
}
