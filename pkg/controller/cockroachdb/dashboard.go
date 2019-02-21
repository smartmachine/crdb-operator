package cockroachdb

import (
	dbv1alpha1 "go.smartmachine.io/crdb-operator/pkg/apis/db/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const DashboardHandler Name = 1400

func init() {
	// Register Client
	info := NewInfo(DashboardServiceForCockroachDB, createIfNotExist, &corev1.Service{})
	info.Postfix = "-dashboard"
	info.SpecConditional = "Spec.Dashboard.Enabled"
	info.SuppressStatus = true
	err := Register(DashboardHandler, info)
	if err != nil {
		panic(err.Error())
	}
}

// dashboardServiceForCockroachDB returns a cockroachdb DashboardService object
func DashboardServiceForCockroachDB(r *ReconcileCockroachDB, m *dbv1alpha1.CockroachDB) interface{} {

	reqLogger := log.WithValues("CockroachDB.Meta.Name", m.ObjectMeta.Name, "CockroachDB.Meta.Namespace", m.ObjectMeta.Namespace)
	reqLogger.Info("Reconciling CockroachDB")

	ls := labelsForCockroachDB(m.Name)

	dep := &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Service",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name + "-dashboard",
			Namespace: m.Namespace,
			Labels:    ls,
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{{
				Port:       8080,
				TargetPort: intstr.FromInt(8080),
				Name:       "http",
				NodePort:   m.Spec.Dashboard.NodePort,
			}},
			Type: corev1.ServiceTypeNodePort,
			Selector: map[string]string{
				"app": m.Name,
			},
		},
	}

	// Set CockroachDB instance as the owner and controller
	err := controllerutil.SetControllerReference(m, dep, r.scheme)
	if err != nil {
		reqLogger.Error(err, "Failed to set Controller Reference", "m", m, "dep", dep, "r.scheme", r.scheme)
	}
	return dep
}