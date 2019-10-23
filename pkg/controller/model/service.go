package model

import (
	"github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
	"github.com/integr8ly/grafana-operator/pkg/controller/actions"
	"github.com/integr8ly/grafana-operator/pkg/controller/common"
	"github.com/integr8ly/grafana-operator/pkg/controller/state"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func reconcileWithExisting(oldService, newService *v1.Service) *v1.Service {
	if oldService == nil {
		return newService
	}

	newService.Spec.ClusterIP = oldService.Spec.ClusterIP
	newService.ResourceVersion = oldService.ResourceVersion
	return newService
}

func calculateNextAction(oldService, newService *v1.Service) actions.ClusterAction {
	if oldService == nil {
		return actions.ClusterActionCreate
	}

	return actions.ClusterActionUpdate
}

func ServiceSelector(cr *v1alpha1.Grafana) client.ObjectKey {
	return client.ObjectKey{
		Namespace: cr.Namespace,
		Name:      GrafanaModelServiceName,
	}
}

func Service(cr *v1alpha1.Grafana, state *state.ClusterState) (runtime.Object, actions.ClusterAction) {
	config := common.GetControllerConfig()
	oldService := state.GetService()

	newService := &v1.Service{
		ObjectMeta: v12.ObjectMeta{
			Name:        GrafanaModelServiceName,
			Namespace:   cr.Namespace,
			Labels:      cr.Spec.Service.Labels,
			Annotations: cr.Spec.Service.Annotations,
		},
		Spec: v1.ServiceSpec{
			Ports: []v1.ServicePort{
				{
					Protocol:   "TCP",
					Port:       GrafanaModelContainerPort,
					TargetPort: intstr.FromString(GrafanaModelTargetPortName),
				},
			},
			Selector: map[string]string{
				"app": config.GetConfigString(common.ConfigPodLabelValue, common.PodLabelDefaultValue),
			},
			Type: option(cr.Spec.Service.Type, common.DefaultServiceType).(v1.ServiceType),
		},
	}

	newService = reconcileWithExisting(oldService, newService)
	action := calculateNextAction(oldService, newService)
	return newService, action
}
