package cockroachdb

import (
	dbv1alpha1 "go.smartmachine.io/crdb-operator/pkg/apis/db/v1alpha1"
	rbacv1 "k8s.io/api/rbac/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const (
	RoleHandler        Name = 200
	RoleBindingHandler Name = 300
)

func init() {
	// Register Role
	info := NewInfo(RoleForCockroachDB, createIfNotExist, &rbacv1.Role{})
	err := Register(RoleHandler, info)
	if err != nil {
		panic(err.Error())
	}

	// Register RoleBinding
	info = NewInfo(RoleBindingForCockroachDB, createIfNotExist, &rbacv1.RoleBinding{})
	err = Register(RoleBindingHandler, info)
	if err != nil {
		panic(err.Error())
	}
}

// roleForCockroachDB returns a cockroachdb Role object
func RoleForCockroachDB(r *ReconcileCockroachDB, m *dbv1alpha1.CockroachDB) interface{} {

	reqLogger := log.WithValues("CockroachDB.Meta.Name", m.ObjectMeta.Name, "CockroachDB.Meta.Namespace", m.ObjectMeta.Namespace)
	reqLogger.Info("Reconciling CockroachDB")

	ls := labelsForCockroachDB(m.Name)

	dep := &rbacv1.Role{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "Role",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
			Labels:    ls,
		},
		Rules: []rbacv1.PolicyRule{{
			APIGroups: []string{""},
			Resources: []string{"secrets"},
			Verbs:     []string{"create", "get"},
		}},
	}

	// Set CockroachDB instance as the owner and controller
	err := controllerutil.SetControllerReference(m, dep, r.scheme)
	if err != nil {
		reqLogger.Error(err, "Failed to set Controller Reference", "m", m, "dep", dep, "r.scheme", r.scheme)
	}
	return dep
}

// roleBindingForCockroachDB returns a cockroachdb RoleBinding object
func RoleBindingForCockroachDB(r *ReconcileCockroachDB, m *dbv1alpha1.CockroachDB) interface{} {

	reqLogger := log.WithValues("CockroachDB.Meta.Name", m.ObjectMeta.Name, "CockroachDB.Meta.Namespace", m.ObjectMeta.Namespace)
	reqLogger.Info("Reconciling CockroachDB")

	ls := labelsForCockroachDB(m.Name)

	dep := &rbacv1.RoleBinding{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "rbac.authorization.k8s.io/v1",
			Kind:       "RoleBinding",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
			Labels:    ls,
		},
		RoleRef: rbacv1.RoleRef{
			APIGroup: "rbac.authorization.k8s.io",
			Kind:     "Role",
			Name:     m.Name,
		},
		Subjects: []rbacv1.Subject{{
			Kind:      "ServiceAccount",
			Name:      m.Name,
			Namespace: m.Namespace,
		}},
	}

	// Set CockroachDB instance as the owner and controller
	err := controllerutil.SetControllerReference(m, dep, r.scheme)
	if err != nil {
		reqLogger.Error(err, "Failed to set Controller Reference", "m", m, "dep", dep, "r.scheme", r.scheme)
	}
	return dep
}
