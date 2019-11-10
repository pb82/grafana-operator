package grafana

import (
	"github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
	"github.com/integr8ly/grafana-operator/pkg/controller/common"
	"github.com/integr8ly/grafana-operator/pkg/controller/model"
)

type GrafanaReconciler struct{}

func NewGrafanaReconciler() *GrafanaReconciler {
	return &GrafanaReconciler{}
}

func (i *GrafanaReconciler) Reconcile(state *common.ClusterState, cr *v1alpha1.Grafana) common.DesiredClusterState {
	desired := common.DesiredClusterState{}

	desired = desired.AddAction(i.getGrafanaServiceDesiredState(state, cr))
	desired = desired.AddAction(i.getGrafanaServiceAccountDesiredState(state, cr))
	desired = desired.AddAction(i.getGrafanaConfigDesiredState(state, cr))

	return desired
}

func (i *GrafanaReconciler) getGrafanaServiceDesiredState(state *common.ClusterState, cr *v1alpha1.Grafana) common.ClusterAction {
	if state.GrafanaService == nil {
		return common.GenericCreateAction{
			Ref: model.GrafanaService(cr),
			Msg: "create grafana service",
		}
	}

	return common.GenericUpdateAction{
		Ref: model.GrafanaServiceReconciled(cr, state.GrafanaService),
		Msg: "update grafana service",
	}
}

func (i *GrafanaReconciler) getGrafanaServiceAccountDesiredState(state *common.ClusterState, cr *v1alpha1.Grafana) common.ClusterAction {
	if state.GrafanaServiceAccount == nil {
		return common.GenericCreateAction{
			Ref: model.GrafanaServiceAccount(cr),
			Msg: "create grafana service account",
		}
	}

	return common.GenericUpdateAction{
		Ref: model.GrafanaServiceAccountReconciled(cr, state.GrafanaServiceAccount),
		Msg: "update grafana service account",
	}
}

func (i *GrafanaReconciler) getGrafanaConfigDesiredState(state *common.ClusterState, cr *v1alpha1.Grafana) common.ClusterAction {
	if state.GrafanaConfig == nil {
		config, err := model.GrafanaConfig(cr)
		if err != nil {
			log.Error(err, "error creating grafana config")
			return nil
		}

		return common.GenericCreateAction{
			Ref: config,
			Msg: "create grafana service account",
		}
	}

	return common.GenericUpdateAction{
		Ref: model.GrafanaServiceAccountReconciled(cr, state.GrafanaServiceAccount),
		Msg: "update grafana service account",
	}
}
