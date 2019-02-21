package cockroachdb

import (
	"context"
	"fmt"
	"github.com/mcuadros/go-lookup"
	dbv1alpha1 "go.smartmachine.io/crdb-operator/pkg/apis/db/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"reflect"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"time"
)

func createIfNotExist(info *Info, db *dbv1alpha1.CockroachDB, r *ReconcileCockroachDB) (bool, reconcile.Result, error) {
	name := reflect.TypeOf(info.Object).String()
	reqLogger := log.WithValues(name + ".Name", db.Name, name + ".Namespace", db.Namespace)

	if info.SpecConditional != "" {
		val, err := lookup.LookupString(*db, info.SpecConditional)
		if err == nil {
			if !val.Interface().(bool) {
				return false, reconcile.Result{}, nil
			}
		} else {
			reqLogger.Error(err, "unable to parse SpecConditional")
		}
	}

	// Scope NamespacedName correctly
	var objectName types.NamespacedName
	if info.NoNamespace {
		objectName = types.NamespacedName{Name: fmt.Sprintf("%s%s",db.Name, info.Postfix)}
	} else {
		objectName = types.NamespacedName{Name: fmt.Sprintf("%s%s",db.Name, info.Postfix), Namespace: db.Namespace}
	}

	// Check if the Object already exists, if not create a new one
	err := r.client.Get(context.TODO(), objectName, info.Object)
	if err != nil && errors.IsNotFound(err) {
		// Define a new resource object
		dep := info.Resource.CallHandler(r, db)
		reqLogger.Info(fmt.Sprintf("Creating a new %s", name))
		err = r.client.Create(context.TODO(), dep.(runtime.Object))
		if err != nil {
			reqLogger.Error(err, fmt.Sprintf("Failed to create a new %s", name))
			return true, reconcile.Result{}, err
		}
		if !info.SuppressStatus {
			r.status = fmt.Sprintf("Created %s(%s)", name, objectName.Name)
		}
		// Resource created successfully - return and requeue
		return true, reconcile.Result{Requeue: true}, nil
	} else if err != nil {
		reqLogger.Error(err, fmt.Sprintf("Failed to get %s", name))
		return true, reconcile.Result{}, err
	}
	return false, reconcile.Result{}, nil
}

func updateStatus(info *Info, db *dbv1alpha1.CockroachDB, r *ReconcileCockroachDB) (bool, reconcile.Result, error) {
	reqLogger := log.WithValues("Name", db.Name, "Namespace", db.Namespace)
	// Update the CockroachDB status with the pod names
	// List the pods for this cockroachdb's deployment
	podList := &corev1.PodList{}
	labelSelector := labels.SelectorFromSet(labelsForCockroachDB(db.Name))
	listOps := &client.ListOptions{Namespace: db.Namespace, LabelSelector: labelSelector}
	err := r.client.List(context.TODO(), listOps, podList)
	if err != nil {
		reqLogger.Error(err, "Failed to list pods")
		return true, reconcile.Result{}, err
	}
	var nodes []dbv1alpha1.CockroachDBNode
	var node corev1.Pod
	for _, node = range podList.Items {
		nodes = append(nodes, dbv1alpha1.CockroachDBNode{
			Name: node.Name,
			Ready: podReadiness(node),
			Serving: podServing(node),
		})
	}

	if !reflect.DeepEqual(nodes, db.Status.Nodes) {
		db.Status.Nodes = nodes
		err := r.client.Status().Update(context.TODO(), db)
		if err != nil {
			reqLogger.Error(err, "Unable to update state of cluster.")
		}
		return true, reconcile.Result{Requeue: true}, nil
	}

	return false, reconcile.Result{}, nil
}

func initCluster(info *Info, db *dbv1alpha1.CockroachDB, r *ReconcileCockroachDB) (bool, reconcile.Result, error) {
	reqLogger := log.WithValues("Job.Name", db.Name, "Job.Namespace", db.Namespace)
	// Check if the Init Job has already run, if not create a new one
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: db.Name, Namespace: db.Namespace}, info.Object)
	if err != nil && errors.IsNotFound(err) {
		// Init Job Does not exist, let's check if the StatefulSet is ready for us
		if !isClusterReady(db) {
			if !isClusterServing(db) {
				reqLogger.Info("waiting for Pods to become ready")
				r.status = "waiting for Pods to become ready"
				return true, reconcile.Result{Requeue: true, RequeueAfter: time.Second * 10}, nil
			} else {
				reqLogger.Info("cluster seems to be serving, starting batch job")
			}
		}

		// All the initial nodes are Running but not Ready
		dep := info.Resource.CallHandler(r, db)
		reqLogger.Info("Creating an Init Batch Job")
		err = r.client.Create(context.TODO(), dep.(runtime.Object))
		if err != nil {
			reqLogger.Error(err, "Failed to create new Job")
			return true, reconcile.Result{}, err
		}
		// Job created successfully - return and requeue
		r.status = "Initialising Cluster"
		return true, reconcile.Result{Requeue: true, RequeueAfter: time.Second * 5}, nil
	} else if err != nil {
		reqLogger.Error(err, "Failed to get Init Job")
		return true, reconcile.Result{}, err
	}

	if !isClusterServing(db) {
		log.Info("waiting for cluster to start serving")
		return true, reconcile.Result{Requeue: true, RequeueAfter: time.Second * 5}, nil
	}

	if r.status != "Cluster Serving" {
		r.status = "Cluster Serving"
		return true, reconcile.Result{Requeue: true}, nil
	}

	return false, reconcile.Result{}, nil
}

// labelsForCockroachDB returns the labels for selecting the resources
// belonging to the given cockroachdb CR name.
func labelsForCockroachDB(name string) map[string]string {
	//return map[string]string{"app": name, "cockroachdb_cr": name}
	return map[string]string{"app": name}
}

// are Pods ready to be initialized?
func podReadiness(pod corev1.Pod) bool {
	if len(pod.Status.ContainerStatuses) > 0 {
		status := pod.Status.ContainerStatuses[0]
		if status.Name == "cockroachdb" && status.Ready == false && status.RestartCount == 0 && status.State.Running != nil {
			return true
		}
	}
	return false
}

// are Pods ready to be initialized?
func podServing(pod corev1.Pod) bool {
		if len(pod.Status.ContainerStatuses) > 0 {
			status := pod.Status.ContainerStatuses[0]
			if status.Name == "cockroachdb" && status.Ready == true {
				return true
			}
		}
		return false
}

func isClusterReady(db *dbv1alpha1.CockroachDB) bool {
	var readyNodes int32 = 0
	for _, node := range db.Status.Nodes {
		if node.Ready {
			readyNodes++
		}
	}
	if readyNodes == db.Spec.Cluster.Size {
		return true
	}
	return false
}

func isClusterServing(db *dbv1alpha1.CockroachDB) bool {
	var servingNodes int32 = 0
	for _, node := range db.Status.Nodes {
		if node.Serving {
			servingNodes++
		}
	}
	if servingNodes == db.Spec.Cluster.Size {
		return true
	}
	return false
}