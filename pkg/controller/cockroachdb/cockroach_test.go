package cockroachdb

import (
	"context"
	dbv1alpha1 "go.smartmachine.io/crdb-operator/pkg/apis/db/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/api/policy/v1beta1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"os"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"testing"
)

var (
	name                = "crdb-test"
	namespace           = "crdb"
	image               = "cockroachdb/cockroach:v2.1.6"
	size int32          = 3
	requestMemory       = "300Mi"
	limitMemory         = "500Mi"
	storagePerNode      = "100Gi"
	maxUnavailableNodes = 1

	cl                        client.Client
	r                         ReconcileCockroachDB
	req                       reconcile.Request
	reqNoNamespace            reconcile.Request

)

func TestMain(m *testing.M) {
	crdb := &dbv1alpha1.CockroachDB{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		Spec: dbv1alpha1.CockroachDBSpec{
			Cluster: dbv1alpha1.CockroachDBClusterSpec{
				Image: image,
				Size: size,
				RequestMemory: requestMemory,
				LimitMemory: limitMemory,
				StoragePerNode: storagePerNode,
				MaxUnavailable: maxUnavailableNodes,
			},
		},
	}

	objs := []runtime.Object{crdb}

	s := scheme.Scheme
	s.AddKnownTypes(dbv1alpha1.SchemeGroupVersion, crdb)

	cl = fake.NewFakeClient(objs...)


	r = ReconcileCockroachDB{client: cl, scheme: s}

	req = reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}

	reqNoNamespace = reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name: name,
		},
	}
	os.Exit(m.Run())
}

func TestCockroachDBController(t *testing.T) {
	t.Run("TestServiceAccount",      ServiceAccount)
	t.Run("TestRole",                Role)
	t.Run("TestRoleBinding",         RoleBinding)
	t.Run("TestClusterRole",         ClusterRole)
	t.Run("TestClusterRoleBinding",  ClusterRoleBinding)
	t.Run("TestPublicService",       PublicService)
	t.Run("TestService",             Service)
	t.Run("TestPodDisruptionBudget", PodDisruptionBudget)
	t.Run("TestStatefulSet",         StatefulSet)
}

func ServiceAccount(t *testing.T) {
	res, err := r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}
	// Check the result of reconciliation to make sure it has the desired state.
	if !res.Requeue {
		t.Error("reconcile did not requeue request as expected")
	}

	// Check if deployment has been created and has the correct size.
	sa := &corev1.ServiceAccount{}
	err = cl.Get(context.TODO(), req.NamespacedName, sa)
	if err != nil {
		t.Fatalf("get serviceaccount: (%+v)", err)
	}
}

// Test if Role is created
func Role(t *testing.T) {
	res, err := r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}
	// Check the result of reconciliation to make sure it has the desired state.
	if !res.Requeue {
		t.Error("reconcile did not requeue request as expected")
	}

	// Check if role has been created
	role := &rbacv1.Role{}
	err = cl.Get(context.TODO(), req.NamespacedName, role)
	if err != nil {
		t.Fatalf("get role: (%+v)", err)
	}
}

// Test if RoleBinding is created
func RoleBinding(t *testing.T) {
	res, err := r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}
	// Check the result of reconciliation to make sure it has the desired state.
	if !res.Requeue {
		t.Error("reconcile did not requeue request as expected")
	}

	// Check if rolebinding has been created
	role := &rbacv1.RoleBinding{}
	err = cl.Get(context.TODO(), req.NamespacedName, role)
	if err != nil {
		t.Fatalf("get rolebinding: (%+v)", err)
	}
}

// Test if ClusterRole is created
func ClusterRole(t *testing.T) {
	res, err := r.Reconcile(reqNoNamespace)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}
	// Check the result of reconciliation to make sure it has the desired state.
	if !res.Requeue {
		t.Error("reconcile did not requeue request as expected")
	}

	// Check if clusterrole has been created
	role := &rbacv1.ClusterRole{}
	err = cl.Get(context.TODO(), reqNoNamespace.NamespacedName, role)
	if err != nil {
		t.Fatalf("get clusterrole: (%+v)", err)
	}
}

// Test if ClusterRoleBinding is created
func ClusterRoleBinding(t *testing.T) {
	res, err := r.Reconcile(reqNoNamespace)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}
	// Check the result of reconciliation to make sure it has the desired state.
	if !res.Requeue {
		t.Error("reconcile did not requeue request as expected")
	}

	// Check if clusterrolebinding has been created
	role := &rbacv1.ClusterRoleBinding{}
	err = cl.Get(context.TODO(), reqNoNamespace.NamespacedName, role)
	if err != nil {
		t.Fatalf("get clusterrolebinding: (%+v)", err)
	}
}

// Test if PublicService is created
func PublicService(t *testing.T) {
	customReq := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name + "-public",
			Namespace: namespace},
	}
	res, err := r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}
	// Check the result of reconciliation to make sure it has the desired state.
	if !res.Requeue {
		t.Error("reconcile did not requeue request as expected")
	}

	// Check if publicservice has been created
	role := &corev1.Service{}
	err = cl.Get(context.TODO(), customReq.NamespacedName, role)
	if err != nil {
		t.Fatalf("get publicservice: (%+v)", err)
	}
}

// Test if Service is created
func Service(t *testing.T) {
	res, err := r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}
	// Check the result of reconciliation to make sure it has the desired state.
	if !res.Requeue {
		t.Error("reconcile did not requeue request as expected")
	}

	// Check if service has been created
	role := &corev1.Service{}
	err = cl.Get(context.TODO(), req.NamespacedName, role)
	if err != nil {
		t.Fatalf("get service: (%+v)", err)
	}
}

// Test if PodDisruptionBudget is created
func PodDisruptionBudget(t *testing.T) {
	res, err := r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}
	// Check the result of reconciliation to make sure it has the desired state.
	if !res.Requeue {
		t.Error("reconcile did not requeue request as expected")
	}

	// Check if service has been created
	obj := &v1beta1.PodDisruptionBudget{}
	err = cl.Get(context.TODO(), req.NamespacedName, obj)
	if err != nil {
		t.Fatalf("get poddisruptionbudget: (%+v)", err)
	}

	if obj.Spec.MaxUnavailable.IntValue() != maxUnavailableNodes {
		t.Fatalf("expected PodDisruptionBudget MaxUnavailable to be %d, got %d", maxUnavailableNodes, obj.Spec.MaxUnavailable.IntValue())
	}
}
