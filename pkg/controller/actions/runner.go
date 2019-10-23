package actions

import (
	"context"

	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Action struct {
	Action ClusterAction
	Object runtime.Object
}

type TargetClusterState []Action

func (t *TargetClusterState) Add(action ClusterAction, object runtime.Object) {
	*t = append(*t, Action{
		Action: action,
		Object: object,
	})
}

type ActionRunner interface {
	Run(state TargetClusterState) error
}

type ActionRunnerImpl struct {
	client client.Client
	ctx    context.Context
}

func NewActionRunner(client client.Client, ctx context.Context) ActionRunner {
	return &ActionRunnerImpl{
		client: client,
		ctx:    ctx,
	}
}

func (r *ActionRunnerImpl) Run(state TargetClusterState) error {
	for _, action := range state {
		switch action.Action {
		case ClusterActionCreate:
			err := r.client.Create(r.ctx, action.Object)
			if err != nil {
				return err
			}
		case ClusterActionUpdate:
			err := r.client.Update(r.ctx, action.Object)
			if err != nil {
				return err
			}
		}
	}
}
