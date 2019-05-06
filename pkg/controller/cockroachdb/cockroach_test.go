package cockroachdb

import (
	"context"
	dbv1alpha1 "go.smartmachine.io/crdb-operator/pkg/apis/db/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"testing"
)

func TestCockroachDBController(t *testing.T) {

	var (
		name            = "crdb-test"
		namespace       = "crdb"
		replicas  int32 = 3
	)

	crdb := &dbv1alpha1.CockroachDB{
		ObjectMeta: metav1.ObjectMeta{
			Name:      name,
			Namespace: namespace,
		},
	}

	objs := []runtime.Object{crdb}

	s := scheme.Scheme
	s.AddKnownTypes(dbv1alpha1.SchemeGroupVersion, crdb)

	client := fake.NewFakeClient(objs...)

	r := ReconcileCockroachDB{client: client, scheme: s}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      name,
			Namespace: namespace,
		},
	}

	res, err := r.Reconcile(req)
	if err != nil {
		t.Fatalf("reconcile: (%v)", err)
	}
	// Check the result of reconciliation to make sure it has the desired state.
	if !res.Requeue {
		t.Error("reconcile did not requeue request as expected")
	}

	// Check if deployment has been created and has the correct size.
	dep := &appsv1.StatefulSet{}
	err = client.Get(context.TODO(), req.NamespacedName, dep)
	if err != nil {
		t.Fatalf("get deployment: (%v)", err)
	}
	dsize := *dep.Spec.Replicas
	if dsize != replicas {
		t.Errorf("dep size (%d) is not the expected size (%d)", dsize, replicas)
	}

}
