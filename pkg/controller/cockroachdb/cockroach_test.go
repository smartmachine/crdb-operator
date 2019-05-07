package cockroachdb

import (
	"context"
	dbv1alpha1 "go.smartmachine.io/crdb-operator/pkg/apis/db/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
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

	cl client.Client
	r  ReconcileCockroachDB
)

func TestMain(m *testing.M) {
	crdb := &dbv1alpha1.CockroachDB{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
		TypeMeta: metav1.TypeMeta{
			APIVersion: "db.smartmachine.io/v1alpha1",
			Kind: "CockroachDB",
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
		Status: dbv1alpha1.CockroachDBStatus{
			State: "Cluster Serving",
			Nodes: []dbv1alpha1.CockroachDBNode{
				{
					Name: name + "-1",
					Ready: false,
					Serving: true,
				},
				{
					Name: name + "-2",
					Ready: false,
					Serving: true,
				},
				{
					Name: name + "-3",
					Ready: false,
					Serving: true,
				},
			},
		},
	}

	pod1 := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind: "Pod",
			APIVersion: "core/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name + "-1",
			Namespace: namespace,
			Labels: map[string]string{
				"app": name,
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Name: "cockroachdb",
				Image: "cockroachdb/cockroach:2.1.6",
			}},
		},
		Status: corev1.PodStatus{
			ContainerStatuses: []corev1.ContainerStatus{{
				Name: "cockroachdb",
				Ready: true,
			}},
		},
	}

	pod2 := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind: "Pod",
			APIVersion: "core/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name + "-2",
			Namespace: namespace,
			Labels: map[string]string{
				"app": name,
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Name: "cockroachdb",
				Image: "cockroachdb/cockroach:2.1.6",
			}},
		},
		Status: corev1.PodStatus{
			ContainerStatuses: []corev1.ContainerStatus{{
				Name: "cockroachdb",
				Ready: true,
			}},
		},
	}

	pod3 := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			Kind: "Pod",
			APIVersion: "core/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name: name + "-3",
			Namespace: namespace,
			Labels: map[string]string{
				"app": name,
			},
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{{
				Name: "cockroachdb",
				Image: "cockroachdb/cockroach:2.1.6",
			}},
		},
		Status: corev1.PodStatus{
			ContainerStatuses: []corev1.ContainerStatus{{
				Name: "cockroachdb",
				Ready: true,
			}},
		},
	}

	objs := []runtime.Object{crdb, pod1, pod2, pod3}

	s := scheme.Scheme
	s.AddKnownTypes(dbv1alpha1.SchemeGroupVersion, crdb)
	s.AddKnownTypes(corev1.SchemeGroupVersion, pod1, pod2, pod3)

	cl = fake.NewFakeClient(objs...)

	r = ReconcileCockroachDB{client: cl, scheme: s}

	os.Exit(m.Run())
}

func TestCockroachDBController(t *testing.T) {

	var tests = []struct {
		description string
		name        string
		namePostfix string
		namespace   string
		object      runtime.Object
	}{
		{description: "TestServiceAccount",      name: name, namespace: namespace, object: &corev1.ServiceAccount{}},
		{description: "TestRole",                name: name, namespace: namespace, object: &rbacv1.Role{}},
		{description: "TestRoleBinding",         name: name, namespace: namespace, object: &rbacv1.RoleBinding{}},
		{description: "TestClusterRole",         name: name, namespace: "",        object: &rbacv1.ClusterRole{}},
		{description: "TestClusterRoleBinding",  name: name, namespace: "",        object: &rbacv1.ClusterRoleBinding{}},
		{description: "TestPublicService",       name: name, namespace: namespace, object: &corev1.Service{}, namePostfix: "-public"},
		{description: "TestService",             name: name, namespace: namespace, object: &corev1.Service{}},
		{description: "TestPodDisruptionBudget", name: name, namespace: namespace, object: &policyv1beta1.PodDisruptionBudget{}},
		{description: "TestStatefulSession",     name: name, namespace: namespace, object: &appsv1.StatefulSet{}},
		{description: "TestBatchJob",            name: name, namespace: namespace, object: &batchv1.Job{}},
	}

	for _, test := range tests {

		t.Run(test.description, func(t *testing.T) {

			crNamespacedName := types.NamespacedName{
				Name:      test.name,
				Namespace: test.namespace,
			}

			objNamespacedName := types.NamespacedName{
				Name:      test.name + test.namePostfix,
				Namespace: test.namespace,
			}

			req := reconcile.Request{
				NamespacedName: crNamespacedName,
			}

			res, err := r.Reconcile(req)
			if err != nil {
				t.Fatalf("reconcile: (%+v)", err)
			}
			// Check the result of reconciliation to make sure it has the desired state.
			if !res.Requeue {
				t.Error("reconcile did not requeue request as expected")
			}

			// Check if object has been created
			err = cl.Get(context.TODO(), objNamespacedName, test.object)
			if err != nil {
				t.Fatalf("get %T: (%+v)", test.object, err)
			}

		})
	}
}
