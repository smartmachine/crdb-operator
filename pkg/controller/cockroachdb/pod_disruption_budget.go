package cockroachdb

import (
	dbv1alpha1 "go.smartmachine.io/crdb-operator/pkg/apis/db/v1alpha1"
	policyv1beta1 "k8s.io/api/policy/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	PodDisruptionBudgetHandler Name = 800
)

func init() {
	// Register PodDisruptionBudget
	info := NewInfo(PodDisruptionBudgetForCockroachDB, createIfNotExist, &policyv1beta1.PodDisruptionBudget{})
	err := Register(PodDisruptionBudgetHandler, info)
	if err != nil {
		panic(err.Error())
	}
}

// podDisruptionBudgetForCockroachDB returns a cockroachdb PodDisruptionBudget object
func PodDisruptionBudgetForCockroachDB(r *ReconcileCockroachDB, m *dbv1alpha1.CockroachDB) interface{} {

	reqLogger := log.WithValues("CockroachDB.Meta.Name", m.ObjectMeta.Name, "CockroachDB.Meta.Namespace", m.ObjectMeta.Namespace)
	reqLogger.Info("Reconciling CockroachDB")

	ls := labelsForCockroachDB(m.Name)

	maxUnavailable := intstr.FromInt(m.Spec.Cluster.MaxUnavailable)
	selector := metav1.LabelSelector{
		MatchLabels: map[string]string{
			"app": m.Name,
		},
	}

	dep := &policyv1beta1.PodDisruptionBudget{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "policy/v1beta1",
			Kind:       "PodDisruptionBudget",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
			Labels:    ls,
		},
		Spec: policyv1beta1.PodDisruptionBudgetSpec{
			Selector:       &selector,
			MaxUnavailable: &maxUnavailable,
		},
	}

	// Set CockroachDB instance as the owner and controller
	err := controllerutil.SetControllerReference(m, dep, r.scheme)
	if err != nil {
		reqLogger.Error(err, "Failed to set Controller Reference", "m", m, "dep", dep, "r.scheme", r.scheme)
	}
	return dep
}