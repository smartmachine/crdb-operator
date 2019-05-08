package cockroachdb

import (
	dbv1alpha1 "go.smartmachine.io/crdb-operator/pkg/apis/db/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var resources = []Info{
	{Resource: serviceAccount,      Reconcile: createIfNotExist},
	{Resource: role,                Reconcile: createIfNotExist},
	{Resource: roleBinding,         Reconcile: createIfNotExist},
	{Resource: clusterRole,         Reconcile: createIfNotExist, NoNamespace: true},
	{Resource: clusterRoleBinding,  Reconcile: createIfNotExist, NoNamespace: true},
	{Resource: publicService,       Reconcile: createIfNotExist, Postfix: "-public"},
	{Resource: service,             Reconcile: createIfNotExist},
	{Resource: podDisruptionBudget, Reconcile: createIfNotExist},
	{Resource: statefulSet,         Reconcile: createIfNotExist},
	{Resource: nil,                 Reconcile: waitForInit},
	{Resource: batchJob,            Reconcile: initCluster},
	{Resource: nil,                 Reconcile: waitForServing},
	{Resource: crdbClient,          Reconcile: createIfNotExist, Postfix: "-client",    SpecConditional: "Spec.Client.Enabled"},
	{Resource: dashboard,           Reconcile: createIfNotExist, Postfix: "-dashboard", SpecConditional: "Spec.Dashboard.Enabled" },


}

type ResourceHandler interface {
	CallHandler(r *ReconcileCockroachDB, m *dbv1alpha1.CockroachDB) interface{}
}

type ResourceType func(r *ReconcileCockroachDB, m *dbv1alpha1.CockroachDB) runtime.Object

type ReconcileHandler interface {
	CallHandler(info *Info, db *dbv1alpha1.CockroachDB, r *ReconcileCockroachDB) (bool, reconcile.Result, error)
}

type ReconcileType func(info *Info, db *dbv1alpha1.CockroachDB, r *ReconcileCockroachDB) (bool, reconcile.Result, error)

type Name int
type Info struct {
	Resource    ResourceType
	Reconcile   ReconcileType
	Postfix     string
	SpecConditional string
	NoNamespace bool
}
type Map map[Name]*Info

type Config struct {
	Handlers Map
	keys     []int
}



func (h ResourceType) CallHandler(r *ReconcileCockroachDB, m *dbv1alpha1.CockroachDB) runtime.Object {
	return h(r, m)
}

func (h ReconcileType) CallHandler(info *Info, db *dbv1alpha1.CockroachDB, r *ReconcileCockroachDB) (bool, reconcile.Result, error) {
	return h(info, db, r)
}

