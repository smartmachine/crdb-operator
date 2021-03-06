package cockroachdb

import (
	"context"
	"fmt"
	dbv1alpha1 "go.smartmachine.io/crdb-operator/pkg/apis/db/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_cockroachdb")

// Add creates a new CockroachDB Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileCockroachDB{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("cockroachdb-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource CockroachDB
	err = c.Watch(&source.Kind{Type: &dbv1alpha1.CockroachDB{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource Pods and requeue the owner CockroachDB
	err = c.Watch(&source.Kind{Type: &appsv1.StatefulSet{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &dbv1alpha1.CockroachDB{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileCockroachDB{}

// ReconcileCockroachDB reconciles a CockroachDB object
type ReconcileCockroachDB struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
	nodes  dbv1alpha1.CockroachDBNode
	status string
}

// Reconcile reads that state of the cluster for a CockroachDB object and makes changes based on the state read
// and what is in the CockroachDB.Spec
// a CockroachDB Deployment for each CockroachDB CR
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileCockroachDB) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)

	// Fetch the CockroachDB instance
	cockroachdb := &dbv1alpha1.CockroachDB{}
	err := r.client.Get(context.TODO(), request.NamespacedName, cockroachdb)
	if err != nil {
		if errors.IsNotFound(err) {
			reqLogger.Info("No CockroachDB resources found.")
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		reqLogger.Error(err, "Failed to get CockroachDB.")
		return reconcile.Result{}, err
	}

	defer r.UpdateStatus(cockroachdb)

	for _, info := range Enumerate() {
		if recon, result, err := info.Reconcile.CallHandler(&info, cockroachdb, r); recon == true {
			return result, err
		}
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileCockroachDB) UpdateStatus(db *dbv1alpha1.CockroachDB) {
	if r.status != db.Status.State {
		db.Status.State = r.status
		err := r.client.Status().Update(context.TODO(), db)
		if err != nil {
			log.Error(err, "unable to update Status on CockroachDB object")
			return
		}
		log.Info(fmt.Sprintf("UpdateStatus State: %+v ", db.Status))
	}
}