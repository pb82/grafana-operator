package common

import (
	"context"
	"github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
	"github.com/integr8ly/grafana-operator/pkg/controller/model"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ClusterState struct {
	GrafanaService        *v1.Service
	GrafanaServiceAccount *v1.ServiceAccount
	GrafanaConfig         *v1.ConfigMap
}

func NewClusterState() *ClusterState {
	return &ClusterState{}
}

func (i *ClusterState) Read(ctx context.Context, cr *v1alpha1.Grafana, client client.Client) error {
	err := i.readGrafanaService(ctx, cr, client)
	if err != nil {
		return err
	}

	err = i.readGrafanaServiceAccount(ctx, cr, client)
	if err != nil {
		return err
	}

	err = i.readGrafanaConfig(ctx, cr, client)
	if err != nil {
		return err
	}

	return nil
}

func (i *ClusterState) readGrafanaService(ctx context.Context, cr *v1alpha1.Grafana, client client.Client) error {
	service := model.GrafanaService(cr)
	selector := model.GrafanaServiceSelector(cr)

	err := client.Get(ctx, selector, service)
	if err != nil {
		if meta.IsNoMatchError(err) || errors.IsNotFound(err) {
			i.GrafanaService = nil
			return nil
		} else {
			return err
		}
	}

	i.GrafanaService = service
	return nil
}

func (i *ClusterState) readGrafanaServiceAccount(ctx context.Context, cr *v1alpha1.Grafana, client client.Client) error {
	serviceAccount := model.GrafanaServiceAccount(cr)
	selector := model.GrafanaServiceAccountSelector(cr)

	err := client.Get(ctx, selector, serviceAccount)
	if err != nil {
		if meta.IsNoMatchError(err) || errors.IsNotFound(err) {
			i.GrafanaServiceAccount = nil
			return nil
		} else {
			return err
		}
	}

	i.GrafanaServiceAccount = serviceAccount
	return nil
}

func (i *ClusterState) readGrafanaConfig(ctx context.Context, cr *v1alpha1.Grafana, client client.Client) error {
	config, err := model.GrafanaConfig(cr)
	if err != nil {
		return err
	}

	selector := model.GrafanaConfigSelector(cr)

	err = client.Get(ctx, selector, config)
	if err != nil {
		if meta.IsNoMatchError(err) || errors.IsNotFound(err) {
			i.GrafanaConfig = nil
			return nil
		} else {
			return err
		}
	}

	i.GrafanaConfig = config
	return nil
}
