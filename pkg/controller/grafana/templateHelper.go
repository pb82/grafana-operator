package grafana

import (
	"bytes"
	"fmt"
	"github.com/integr8ly/grafana-operator/pkg/controller/config"
	"io/ioutil"
	v1 "k8s.io/api/core/v1"
	"os"
	"strings"
	"text/template"

	integreatly "github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
)

const (
	// DefaultLogLevel is the default logging level
	DefaultLogLevel = "info"
)

// GrafanaParameters provides the context for the template
type GrafanaParameters struct {
	AdminPassword                   string
	AdminUser                       string
	Anonymous                       bool
	BasicAuth                       bool
	DisableLoginForm                bool
	DisableSignoutMenu              bool
	GrafanaConfigHash               string
	GrafanaConfigMapName            string
	GrafanaDashboardsConfigMapName  string
	GrafanaDatasourcesConfigMapName string
	GrafanaDeploymentName           string
	GrafanaImage                    string
	GrafanaIngressAnnotations       map[string]string
	GrafanaIngressLabels            map[string]string
	GrafanaIngressName              string
	GrafanaIngressPath              string
	GrafanaIngressTLSEnabled        bool
	GrafanaIngressTLSSecretName     string
	GrafanaProvidersConfigMapName   string
	GrafanaRouteName                string
	GrafanaServiceAccountName       string
	GrafanaServiceAnnotations       map[string]string
	GrafanaServiceLabels            map[string]string
	GrafanaServiceName              string
	GrafanaServiceType              string
	GrafanaVersion                  string
	Hostname                        string
	LogLevel                        string
	Namespace                       string
	PluginsInitContainerImage       string
	PluginsInitContainerTag         string
	PodLabelValue                   string
	Replicas                        int
}

// TemplateHelper is the deployment helper object
type TemplateHelper struct {
	Parameters   GrafanaParameters
	TemplatePath string
}

func option(val, defaultVal string) string {
	if val == "" {
		return defaultVal
	}
	return val
}

func getServiceType(serviceType string) string {
	switch v1.ServiceType(strings.TrimSpace(serviceType)) {
	case v1.ServiceTypeClusterIP:
		return serviceType
	case v1.ServiceTypeNodePort:
		return serviceType
	case v1.ServiceTypeLoadBalancer:
		return serviceType
	default:
		return config.DefaultServiceType
	}
}

func getLogLevel(userLogLevel string) string {
	level := strings.TrimSpace(userLogLevel)
	level = strings.ToLower(level)

	switch level {
	case "debug":
		return level
	case "info":
		return level
	case "warn":
		return level
	case "error":
		return level
	case "critical":
		return level
	default:
		return config.DefaultLogLevel
	}
}

// Creates a new templates helper and populates the values for all
// templates properties. Some of them (like the hostname) are set
// by the user in the custom resource
func newTemplateHelper(cr *integreatly.Grafana) *TemplateHelper {
	controllerConfig := config.GetControllerConfig()

	param := GrafanaParameters{
		AdminPassword:                   option(cr.Spec.AdminPassword, "secret"),
		AdminUser:                       option(cr.Spec.AdminUser, "root"),
		Anonymous:                       cr.Spec.Anonymous,
		BasicAuth:                       cr.Spec.BasicAuth,
		DisableLoginForm:                cr.Spec.DisableLoginForm,
		DisableSignoutMenu:              cr.Spec.DisableSignoutMenu,
		GrafanaConfigHash:               cr.Status.LastConfig,
		GrafanaConfigMapName:            config.GrafanaConfigMapName,
		GrafanaDashboardsConfigMapName:  config.GrafanaDashboardsConfigMapName,
		GrafanaDatasourcesConfigMapName: config.GrafanaDatasourcesConfigMapName,
		GrafanaDeploymentName:           config.GrafanaDeploymentName,
		GrafanaImage:                    controllerConfig.GetConfigString(config.ConfigGrafanaImage, config.GrafanaImage),
		GrafanaIngressAnnotations:       cr.Spec.Ingress.Annotations,
		GrafanaIngressLabels:            cr.Spec.Ingress.Labels,
		GrafanaIngressName:              config.GrafanaIngressName,
		GrafanaIngressPath:              cr.Spec.Ingress.Path,
		GrafanaIngressTLSEnabled:        cr.Spec.Ingress.TLSEnabled,
		GrafanaIngressTLSSecretName:     cr.Spec.Ingress.TLSSecretName,
		GrafanaProvidersConfigMapName:   config.GrafanaProvidersConfigMapName,
		GrafanaRouteName:                config.GrafanaRouteName,
		GrafanaServiceAccountName:       config.GrafanaServiceAccountName,
		GrafanaServiceAnnotations:       cr.Spec.Service.Annotations,
		GrafanaServiceLabels:            cr.Spec.Service.Labels,
		GrafanaServiceName:              config.GrafanaServiceName,
		GrafanaVersion:                  controllerConfig.GetConfigString(config.ConfigGrafanaImageTag, config.GrafanaVersion),
		Hostname:                        cr.Spec.Ingress.Hostname,
		LogLevel:                        getLogLevel(cr.Spec.LogLevel),
		Namespace:                       cr.Namespace,
		PluginsInitContainerImage:       controllerConfig.GetConfigString(config.ConfigPluginsInitContainerImage, config.PluginsInitContainerImage),
		PluginsInitContainerTag:         controllerConfig.GetConfigString(config.ConfigPluginsInitContainerTag, config.PluginsInitContainerTag),
		PodLabelValue:                   controllerConfig.GetConfigString(config.ConfigPodLabelValue, config.PodLabelDefaultValue),
		Replicas:                        cr.Spec.InitialReplicas,
	}

	templatePath := os.Getenv("TEMPLATE_PATH")
	if templatePath == "" {
		templatePath = "./templates"
	}

	return &TemplateHelper{
		Parameters:   param,
		TemplatePath: templatePath,
	}
}

// load a templates from a given resource name. The templates must be located
// under ./templates and the filename must be <resource-name>.yaml
func (h *TemplateHelper) loadTemplate(name string) ([]byte, error) {
	path := fmt.Sprintf("%s/%s.yaml", h.TemplatePath, name)
	tpl, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}

	parsed, err := template.New("grafana").Parse(string(tpl))
	if err != nil {
		return nil, err
	}

	var buffer bytes.Buffer
	err = parsed.Execute(&buffer, h.Parameters)
	if err != nil {
		return nil, err
	}

	return buffer.Bytes(), nil
}
