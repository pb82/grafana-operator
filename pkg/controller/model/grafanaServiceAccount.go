package model

import (
	"github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
	v1 "k8s.io/api/core/v1"
	v12 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func GrafanaServiceAccount(cr *v1alpha1.Grafana) *v1.ServiceAccount {
	return &v1.ServiceAccount{
		ObjectMeta: v12.ObjectMeta{
			Name:      GrafanaServiceAccountName,
			Namespace: cr.Namespace,
		},
	}
}

func GrafanaServiceAccountSelector(cr *v1alpha1.Grafana) client.ObjectKey {
	return client.ObjectKey{
		Namespace: cr.Namespace,
		Name:      GrafanaServiceAccountName,
	}
}

func GrafanaServiceAccountReconciled(cr *v1alpha1.Grafana, currentState *v1.ServiceAccount) *v1.ServiceAccount {
	return currentState.DeepCopy()
}
