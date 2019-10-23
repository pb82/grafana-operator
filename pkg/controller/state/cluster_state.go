package state

import (
	"context"

	"github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
	"github.com/integr8ly/grafana-operator/pkg/controller/model"
	v1 "k8s.io/api/core/v1"

	"sigs.k8s.io/controller-runtime/pkg/client"
)

type ClusterState interface {
	Read(client client.Client) error
	GetService() *v1.Service
}

type ClusterStateImpl struct {
	ctx context.Context
	cr  *v1alpha1.Grafana

	// Grafana resoruces
	Service *v1.Service
}

func NewClusterState(cr *v1alpha1.Grafana, ctx context.Context) ClusterState {
	state := &ClusterStateImpl{
		ctx: ctx,
		cr:  cr,
	}

	state.Service = nil

	return state
}

func (i *ClusterStateImpl) Read(client client.Client) error {
	err := i.getService(client)
	if err != nil {
		return err
	}
}

func (i *ClusterStateImpl) getService(client client.Client) error {
	selector := model.ServiceSelector(i.cr)
	service := &v1.Service{}

	err := client.Get(i.ctx, selector, service)
	if err != nil {
		return err
	}

	i.Service = service
	return nil
}

func (i *ClusterStateImpl) GetService() *v1.Service {
	return i.Service
}
