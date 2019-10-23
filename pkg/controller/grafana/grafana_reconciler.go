package grafana

import (
	"github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
	"github.com/integr8ly/grafana-operator/pkg/controller/actions"
	"github.com/integr8ly/grafana-operator/pkg/controller/model"
	"github.com/integr8ly/grafana-operator/pkg/controller/state"
)

type Reconciler interface {
	Reconcile(cr *v1alpha1.Grafana, state *state.ClusterState) actions.TargetClusterState
}

type ReconcilerImpl struct{}

func NewReconciler() Reconciler {
	return &ReconcilerImpl{}
}

func (r *ReconcilerImpl) Reconcile(cr *v1alpha1.Grafana, state *state.ClusterState) actions.TargetClusterState {
	targetState := actions.TargetClusterState{}
	targetState.Add(model.Service(cr, state))
}
