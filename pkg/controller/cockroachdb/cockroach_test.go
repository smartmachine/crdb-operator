package cockroachdb

import (
	dbv1alpha1 "go.smartmachine.io/crdb-operator/pkg/apis/db/v1alpha1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"testing"
)

func TestCockroachDBController(t *testing.T) {

	crdb := &dbv1alpha1.CockroachDB{
		ObjectMeta: v1.ObjectMeta{
			Name:      "crdb-test",
			Namespace: "crdb",
		},
	}

	objs := []runtime.Object{crdb}

	s := scheme.Scheme
	s.AddKnownTypes(dbv1alpha1.SchemeGroupVersion, crdb)

	client := fake.NewFakeClient(objs...)

	r := ReconcileCockroachDB{client: client, scheme: s}

	req := reconcile.Request{
		NamespacedName: types.NamespacedName{
			Name:      "crdb-test",
			Namespace: "crdb",
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

}
