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

	resource := info.Resource.CallHandler(r, db)
	resourceType := reflect.TypeOf(resource)
	resourceName := resourceType.String()
	object := reflect.New(resourceType).Elem().Interface()
	reqLogger := log.WithValues(resourceName + ".Name", db.Name, resourceName + ".Namespace", db.Namespace)

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
	err := r.client.Get(context.TODO(), objectName, object.(runtime.Object))
	if err != nil && errors.IsNotFound(err) {
		// Define a new resource object
		reqLogger.Info(fmt.Sprintf("Creating a new %s", resourceName))
		err = r.client.Create(context.TODO(), resource)
		if err != nil {
			reqLogger.Error(err, fmt.Sprintf("Failed to create a new %s", resourceName))
			return true, reconcile.Result{}, err
		}
		// Resource created successfully - return
		return false, reconcile.Result{}, nil
	} else if err != nil {
		reqLogger.Error(err, fmt.Sprintf("Failed to get %s", resourceName))
		return true, reconcile.Result{}, err
	}
	return false, reconcile.Result{}, nil
}

func waitForInit(info *Info, db *dbv1alpha1.CockroachDB, r *ReconcileCockroachDB) (bool, reconcile.Result, error) {
	reqLogger := log.WithValues("Name", db.Name, "Namespace", db.Namespace)

	if db.Status.ClusterReadyForInit {
		return false, reconcile.Result{}, nil
	}

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
	var readyNodes int32 = 0
	for _, node = range podList.Items {
		nodeReady := podReadyForInit(node)
		if nodeReady {
			readyNodes++
		}
		nodes = append(nodes, dbv1alpha1.CockroachDBNode{
			Name: node.Name,
			ReadyForInit: nodeReady,
		})
	}

	if !reflect.DeepEqual(nodes, db.Status.Nodes) {
		db.Status.Nodes = nodes
		if int32(podList.Size()) == db.Spec.Cluster.Size && readyNodes == db.Spec.Cluster.Size {
			db.Status.ClusterReadyForInit = true
		}
		err := r.client.Status().Update(context.TODO(), db)
		if err != nil {
			reqLogger.Error(err, "Unable to update state of cluster.")
		}
		reqLogger.Info(fmt.Sprintf("waitForInit: %+v", db.Status))
		return true, reconcile.Result{Requeue: true, RequeueAfter: time.Second * 5}, nil
	}

	return false, reconcile.Result{}, nil
}



func initCluster(info *Info, db *dbv1alpha1.CockroachDB, r *ReconcileCockroachDB) (bool, reconcile.Result, error) {

	// our work is done
	if db.Status.ClusterInitialised {
		return false, reconcile.Result{}, nil
	}

	resource := info.Resource.CallHandler(r, db)
	resourceType := reflect.TypeOf(resource)
	object := reflect.New(resourceType).Elem().Interface()

	reqLogger := log.WithValues("Job.Name", db.Name, "Job.Namespace", db.Namespace)
	// Try to retrieve the init job
	err := r.client.Get(context.TODO(), types.NamespacedName{Name: db.Name, Namespace: db.Namespace}, object.(runtime.Object))
	if err == nil {
		// job has just been scheduled, mark it as such
		db.Status.ClusterInitialised = true
		db.Status.ClusterReadyForInit = true
		err = r.client.Status().Update(context.TODO(), db)
		if err != nil {
			reqLogger.Error(err, "Unable to update state of cluster.")
		}
		return true, reconcile.Result{Requeue: true}, nil
	} else if errors.IsNotFound(err) {
		// Job not found, make it so
		reqLogger.Info("Creating an Init Batch Job")
		err = r.client.Create(context.TODO(), resource)
		if err != nil {
			reqLogger.Error(err, "Failed to create new Job")
			return true, reconcile.Result{}, err
		}
		return true, reconcile.Result{Requeue: true}, nil
	}

	reqLogger.Error(err, "Failed to get Init Job")
	return true, reconcile.Result{}, err
}

func waitForServing(info *Info, db *dbv1alpha1.CockroachDB, r *ReconcileCockroachDB) (bool, reconcile.Result, error) {
	reqLogger := log.WithValues("Name", db.Name, "Namespace", db.Namespace)

	if db.Status.ClusterServing {
		return false, reconcile.Result{}, nil
	}

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
	var servingNodes int32 = 0
	for _, node = range podList.Items {
		nodeServing := podServing(node)
		if nodeServing {
			servingNodes++
		}
		nodes = append(nodes, dbv1alpha1.CockroachDBNode{
			Name: node.Name,
			ReadyForInit: podReadyForInit(node),
			Serving:      nodeServing,
		})
	}

	if !reflect.DeepEqual(nodes, db.Status.Nodes) {
		db.Status.Nodes = nodes
		db.Status.ClusterReadyForInit = true
		db.Status.ClusterInitialised = true
		if db.Spec.Cluster.Size == servingNodes && int32(podList.Size()) == db.Spec.Cluster.Size {
			db.Status.ClusterServing = true
		}
		err := r.client.Status().Update(context.TODO(), db)
		if err != nil {
			reqLogger.Error(err, "Unable to update state of cluster.")
		}
		reqLogger.Info(fmt.Sprintf("waitForServing: %+v", db.Status))
		return true, reconcile.Result{Requeue: true}, nil
	}

	return false, reconcile.Result{}, nil
}

func podReadyForInit(pod corev1.Pod) bool {
	if len(pod.Status.ContainerStatuses) < 1 {
		return false
	}
	podStatus := pod.Status.ContainerStatuses[0]

	if podStatus.Name == "cockroachdb" && podStatus.State.Running != nil {
		return true
	}
	return false
}

func podServing(pod corev1.Pod) bool {
	if pod.Status.Phase == corev1.PodRunning {
		for _, condition := range pod.Status.Conditions {
			if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionTrue {
				return true
			}
		}
	}
	return false
}

// labelsForCockroachDB returns the labels for selecting the resources
// belonging to the given cockroachdb CR name.
func labelsForCockroachDB(name string) map[string]string {
	//return map[string]string{"app": name, "cockroachdb_cr": name}
	return map[string]string{"app": name}
}
