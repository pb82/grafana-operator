package grafanadashboard

import (
	"context"
	"crypto/md5"
	"encoding/json"
	defaultErrors "errors"
	"fmt"
	"github.com/integr8ly/grafana-operator/pkg/controller/config"
	"io/ioutil"
	"net/http"
	"net/url"

	i8ly "github.com/integr8ly/grafana-operator/pkg/apis/integreatly/v1alpha1"
	"github.com/integr8ly/grafana-operator/pkg/controller/common"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_grafanadashboard")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new GrafanaDashboard Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileGrafanaDashboard{
		client: mgr.GetClient(),
		scheme: mgr.GetScheme(),
		config: config.GetControllerConfig(),
		helper: common.NewKubeHelper(),
	}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("grafanadashboard-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource GrafanaDashboard
	err = c.Watch(&source.Kind{Type: &i8ly.GrafanaDashboard{}}, &handler.EnqueueRequestForObject{})
	if err == nil {
		log.Info("Starting dashboard controller")
	}

	return err
}

var _ reconcile.Reconciler = &ReconcileGrafanaDashboard{}

// ReconcileGrafanaDashboard reconciles a GrafanaDashboard object
type ReconcileGrafanaDashboard struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
	config *config.ControllerConfig
	helper *common.KubeHelperImpl
}

// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileGrafanaDashboard) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	dashboardLabelSelectors := r.config.GetConfigItem(config.ConfigDashboardLabelSelector, nil)
	if dashboardLabelSelectors == nil {
		return reconcile.Result{RequeueAfter: config.RequeueDelay}, nil
	}

	// Fetch the GrafanaDashboard instance
	instance := &i8ly.GrafanaDashboard{}
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	cr := instance.DeepCopy()
	if match, err := cr.MatchesSelectors(dashboardLabelSelectors.([]*v1.LabelSelector)); err != nil {
		return reconcile.Result{}, err
	} else if !match {
		log.Info(fmt.Sprintf("found dashboard '%s/%s' but labels do not match", cr.Namespace, cr.Name))
		return reconcile.Result{}, nil
	}

	// Resource deleted?
	if cr.DeletionTimestamp != nil {
		return r.deleteDashboard(cr)
	}

	switch cr.Status.Phase {
	case common.StatusResourceUninitialized:
		// New resource
		return r.updatePhase(cr, common.StatusResourceSetFinalizer)
	case common.StatusResourceSetFinalizer:
		// Set finalizer first
		if len(cr.Finalizers) > 0 {
			return r.updatePhase(cr, common.StatusResourceCreated)
		} else {
			return r.setFinalizer(cr)
		}
	case common.StatusResourceCreated:
		// Import / update dashboard
		res, err := r.importDashboard(cr)

		// Requeue periodically to find dashboards that have not been updated
		// but are not yet imported (can happen if Grafana is uninstalled and
		// then reinstalled without an Operator restart
		res.RequeueAfter = config.RequeueDelay
		return res, err
	default:
		return reconcile.Result{}, nil
	}
}

func (r *ReconcileGrafanaDashboard) hasDashboardChanged(d *i8ly.GrafanaDashboard) bool {
	hash := fmt.Sprintf("%x", md5.Sum([]byte(d.Spec.Json+d.Spec.Url)))
	changed := hash != d.Status.LastConfig

	if !changed {
		known, err := r.helper.IsKnown(i8ly.GrafanaDashboardKind, d)
		if err != nil {
			log.Error(err, "error checking dashboard status")
			return false
		}

		// If the dashboard is known and unchanged we don't have to
		// import it again
		if known {
			log.Info("dashboard reconciled but no changes")
			return false
		}
	}
	d.Status.LastConfig = hash

	return true
}

func (r *ReconcileGrafanaDashboard) isJsonValid(d *i8ly.GrafanaDashboard, dashboardJson string) bool {
	var js map[string]interface{}
	err := json.Unmarshal([]byte(dashboardJson), &js)

	if err == nil {
		return true
	}

	// Don't append the same error twice
	msg := fmt.Sprintf("invalid JSON, error: %s", err)
	for _, statusMessage := range d.Status.Messages {
		if statusMessage.Message == msg {
			log.Info("dashboard reconciled but json still invalid")
			return false
		}
	}

	common.AppendMessage(msg, d)
	err = r.client.Update(context.TODO(), d)
	if err != nil {
		log.Error(err, "update dashboard messages failed")
	}

	log.Info("dashboard reconciled but json invalid")
	return false
}

func (r *ReconcileGrafanaDashboard) importDashboard(d *i8ly.GrafanaDashboard) (reconcile.Result, error) {
	operatorNamespace := r.config.GetConfigString(config.ConfigOperatorNamespace, "")
	if operatorNamespace == "" {
		return reconcile.Result{}, defaultErrors.New("operator namespace not yet known")
	}

	changed := r.hasDashboardChanged(d)
	if !changed {
		return reconcile.Result{Requeue: false}, nil
	}

	dashboardJson := d.Spec.Json

	// If a URL is provided, try to fetch the dashboard json from there
	if d.Spec.Url != "" {
		remoteJson, err := r.loadDashboardFromURL(d)
		if err != nil {
			log.Info(fmt.Sprintf("cannot load dashboard from %s, falling back to embedded json", d.Spec.Url))
		} else {
			dashboardJson = remoteJson
		}
	}

	valid := r.isJsonValid(d, dashboardJson)
	if !valid {
		return reconcile.Result{Requeue: false}, nil
	}

	updated, err := r.helper.UpdateDashboard(d, dashboardJson)
	if err != nil {
		return reconcile.Result{}, err
	}

	if !updated {
		return reconcile.Result{RequeueAfter: config.RequeueDelay}, err
	}

	// Reconcile dashboard plugins
	r.config.SetPluginsFor(d)

	msg := fmt.Sprintf("dashboard '%s/%s' imported", d.Namespace, d.Spec.Name)
	common.AppendMessage(msg, d)

	// Update the dashboard to persist the new hash and the satus message
	err = r.client.Update(context.TODO(), d)
	if err == nil {
		log.Info(msg)
	}

	return reconcile.Result{}, err
}

// Try to load remote dashboard json from an url
func (r *ReconcileGrafanaDashboard) loadDashboardFromURL(d *i8ly.GrafanaDashboard) (string, error) {
	_, err := url.ParseRequestURI(d.Spec.Url)
	if err != nil {
		return "", defaultErrors.New("dashboard url specified is not valid")
	}

	resp, err := http.Get(d.Spec.Url)
	if err != nil {
		return "", defaultErrors.New("request to import dashboard failed")
	}
	defer resp.Body.Close()

	body, _ := ioutil.ReadAll(resp.Body)
	response := string(body)

	if resp.StatusCode != 200 {
		return "", defaultErrors.New(fmt.Sprintf("request to import dashboard returned with status %v", resp.StatusCode))
	}

	return response, nil
}

func (r *ReconcileGrafanaDashboard) deleteDashboard(d *i8ly.GrafanaDashboard) (reconcile.Result, error) {
	operatorNamespace := r.config.GetConfigString(config.ConfigOperatorNamespace, "")
	if operatorNamespace == "" {
		return reconcile.Result{}, defaultErrors.New("no monitoring namespace set")
	}

	err := r.helper.DeleteDashboard(d)
	if err == nil {
		log.Info(fmt.Sprintf("dashboard '%s/%s' deleted", d.Namespace, d.Spec.Name))
	}

	r.config.RemovePluginsFor(d)
	return r.removeFinalizer(d)
}

func (r *ReconcileGrafanaDashboard) removeFinalizer(cr *i8ly.GrafanaDashboard) (reconcile.Result, error) {
	cr.Finalizers = nil
	err := r.client.Update(context.TODO(), cr)
	return reconcile.Result{}, err
}

func (r *ReconcileGrafanaDashboard) setFinalizer(cr *i8ly.GrafanaDashboard) (reconcile.Result, error) {
	if len(cr.Finalizers) == 0 {
		cr.Finalizers = append(cr.Finalizers, config.ResourceFinalizerName)
	}
	err := r.client.Update(context.TODO(), cr)
	return reconcile.Result{}, err
}

func (r *ReconcileGrafanaDashboard) updatePhase(cr *i8ly.GrafanaDashboard, phase int) (reconcile.Result, error) {
	cr.Status.Phase = phase
	err := r.client.Update(context.TODO(), cr)
	return reconcile.Result{}, err
}
