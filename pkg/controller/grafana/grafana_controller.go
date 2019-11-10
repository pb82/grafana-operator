package grafana

import (
	"context"
	i8ly "github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
	"github.com/integr8ly/grafana-operator/pkg/controller/common"
	"github.com/integr8ly/grafana-operator/pkg/controller/config"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/client-go/tools/record"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

const ControllerName = "grafana-controller"

var log = logf.Log.WithName(ControllerName)

const OpenShiftOAuthRedirect = "serviceaccounts.openshift.io/oauth-redirectreference.primary"

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Grafana Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	ctx := context.Background()
	ctx, cancel := context.WithCancel(ctx)

	return &ReconcileGrafana{
		client:   mgr.GetClient(),
		scheme:   mgr.GetScheme(),
		helper:   common.NewKubeHelper(),
		plugins:  newPluginsHelper(),
		config:   config.GetControllerConfig(),
		context:  ctx,
		cancel:   cancel,
		recorder: mgr.GetRecorder(ControllerName),
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("grafana-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Grafana
	return c.Watch(&source.Kind{Type: &i8ly.Grafana{}}, &handler.EnqueueRequestForObject{})
}

var _ reconcile.Reconciler = &ReconcileGrafana{}

// ReconcileGrafana reconciles a Grafana object
type ReconcileGrafana struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client   client.Client
	scheme   *runtime.Scheme
	helper   *common.KubeHelperImpl
	plugins  *PluginsHelperImpl
	config   *config.ControllerConfig
	context  context.Context
	cancel   context.CancelFunc
	recorder record.EventRecorder
}

// Reconcile reads that state of the cluster for a Grafana object and makes changes based on the state read
// and what is in the Grafana.Spec
func (r *ReconcileGrafana) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	instance := &i8ly.Grafana{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Stop the dashboard controller from reconciling when grafana is not installed
			r.config.RemoveConfigItem(config.ConfigDashboardLabelSelector)
			return reconcile.Result{}, nil
		}
		return reconcile.Result{}, err
	}

	cr := instance.DeepCopy()

	// Read current state
	currentState := common.NewClusterState()
	err = currentState.Read(r.context, cr, r.client)
	if err != nil {
		return r.manageError(cr, err)
	}

	// Get the actions required to reach the desired state
	reconciler := NewGrafanaReconciler()
	desiredState := reconciler.Reconcile(currentState, cr)

	// Run the actions to reach the desired state
	actionRunner := common.NewClusterActionRunner(r.context, r.client, r.scheme, cr)
	err = actionRunner.RunAll(desiredState)
	if err != nil {
		return r.manageError(cr, err)
	}

	return r.manageSuccess(cr)
}

func (r *ReconcileGrafana) manageError(cr *i8ly.Grafana, issue error) (reconcile.Result, error) {
	r.recorder.Event(cr, "Warning", "ProcessingError", issue.Error())

	cr.Status.Phase = i8ly.PhaseFailing

	err := r.client.Update(r.context, cr)
	if err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{Requeue: false}, nil
}

func (r *ReconcileGrafana) manageSuccess(cr *i8ly.Grafana) (reconcile.Result, error) {
	cr.Status.Phase = i8ly.PhaseReconciling
	err := r.client.Update(r.context, cr)
	if err != nil {
		return r.manageError(cr, err)
	}

	log.Info("desired cluster state met")
	return reconcile.Result{RequeueAfter: config.RequeueDelay}, nil
}
