package cockroachdb

import (
	"errors"
	"fmt"
	dbv1alpha1 "go.smartmachine.io/crdb-operator/pkg/apis/db/v1alpha1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sort"
)


type ResourceHandler interface {
	CallHandler(r *ReconcileCockroachDB, m *dbv1alpha1.CockroachDB) interface{}
}

type ResourceType func(r *ReconcileCockroachDB, m *dbv1alpha1.CockroachDB) interface{}

type ReconcileHandler interface {
	CallHandler(info *Info, db *dbv1alpha1.CockroachDB, r *ReconcileCockroachDB) (bool, reconcile.Result, error)
}

type ReconcileType func(info *Info, db *dbv1alpha1.CockroachDB, r *ReconcileCockroachDB) (bool, reconcile.Result, error)

type Name int
type Info struct {
	Resource    ResourceType
	Reconcile   ReconcileType
	Object      runtime.Object
	Postfix     string
	SpecConditional string
	NoNamespace bool
	SuppressStatus bool
}
type Map map[Name]*Info

type Config struct {
	Handlers Map
	keys     []int
}


var hc = &Config{
	Handlers: make(Map),
	keys: []int{},
}

func (h ResourceType) CallHandler(r *ReconcileCockroachDB, m *dbv1alpha1.CockroachDB) interface{} {
	return h(r, m)
}

func (h ReconcileType) CallHandler(info *Info, db *dbv1alpha1.CockroachDB, r *ReconcileCockroachDB) (bool, reconcile.Result, error) {
	return h(info, db, r)
}

func NewInfo(resource interface{}, recon interface{}, obj runtime.Object) *Info {
	info := &Info{
		Reconcile: ReconcileType(recon.(func(info *Info, db *dbv1alpha1.CockroachDB, r *ReconcileCockroachDB) (bool, reconcile.Result, error))),
		Object: obj,
	}
	if resource != nil {
		info.Resource = ResourceType(resource.(func(r *ReconcileCockroachDB, m *dbv1alpha1.CockroachDB) interface{}))
	}
	return info
}

func Register(h Name, f *Info) error {
	if _, ok := hc.Handlers[h]; ok {
		return errors.New("a ResourceHandler with that id already exists")
	}
	hc.Handlers[h] = f
	log.Info(fmt.Sprintf("Adding Resource with priority %d and type %T", h, f.Object))
	hc.keys = append(hc.keys, int(h))
	sort.Ints(hc.keys)
	return nil
}

func Enumerate() []Info {
	items := make([]Info, len(hc.keys))
	for i, key := range hc.keys {
		items[i] = *hc.Handlers[Name(key)]
	}
	return items
}

