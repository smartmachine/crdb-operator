package e2e

import (
	"context"
	framework "github.com/operator-framework/operator-sdk/pkg/test"
	"github.com/operator-framework/operator-sdk/pkg/test/e2eutil"
	dbv1alpha1 "go.smartmachine.io/crdb-operator/pkg/apis/db/v1alpha1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/apimachinery/pkg/util/wait"
	"k8s.io/client-go/kubernetes"
	"testing"
	"time"
)

var (
	retryInterval        = time.Second * 5
	timeout              = time.Second * 300
	cleanupRetryInterval = time.Second * 1
	cleanupTimeout       = time.Second * 5
)

func TestCRDB(t *testing.T) {

	crdbCRD := &dbv1alpha1.CockroachDB{
		TypeMeta: metav1.TypeMeta{
			Kind: "CockroachDB",
			APIVersion: "db.smartmachine.io/v1alpha1",
		},
	}

	err := framework.AddToFrameworkScheme(dbv1alpha1.AddToScheme, crdbCRD)
	if err != nil {
		t.Fatalf("failed to add custom resource scheme to framework: %v", err)
	}

	ctx := framework.NewTestCtx(t)
	defer ctx.Cleanup()
	err = ctx.InitializeClusterResources(&framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		t.Fatalf("failed to initialize cluster resources: %v", err)
	}
	t.Log("Initialized cluster resources")

	namespace, err := ctx.GetNamespace()
	if err != nil {
		t.Fatal(err)
	}
	// get global framework variables
	f := framework.Global

	err = e2eutil.WaitForOperatorDeployment(t, f.KubeClient, namespace, "crdb-operator", 1, retryInterval, timeout)
	if err != nil {
		t.Fatal(err)
	}

	crdb := &dbv1alpha1.CockroachDB{
		TypeMeta: metav1.TypeMeta{
			Kind:       "CockroachDB",
			APIVersion: "db.smartmachine.io/v1alpha1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      "crdb-test",
			Namespace: "crdb",
		},
		Spec: dbv1alpha1.CockroachDBSpec{
			Cluster: dbv1alpha1.CockroachDBClusterSpec{
				Image:          "cockroachdb/cockroach:v2.1.6",
				Size:           3,
				RequestMemory:  "300Mi",
				LimitMemory:    "500Mi",
				StoragePerNode: "100Gi",
				MaxUnavailable: 1,
			},
			Client: dbv1alpha1.CockroachDBClientSpec{
				Enabled: false,
			},
			Dashboard: dbv1alpha1.CockroachDBDashboardSpec{
				Enabled: false,
			},
		},
	}

	err = f.Client.Create(context.TODO(), crdb, &framework.CleanupOptions{TestContext: ctx, Timeout: cleanupTimeout, RetryInterval: cleanupRetryInterval})
	if err != nil {
		t.Fatal(err)
	}

	err = waitForStatefulSet(t, f.KubeClient, namespace, "crdb-test", 3, retryInterval, timeout)
	if err != nil {
		t.Fatal(err)
	}

	err = f.Client.Get(context.TODO(), types.NamespacedName{Name: "crdb-test", Namespace: namespace}, crdb)
	if err != nil {
		t.Fatal(err)
	}
	crdb.Spec.Cluster.Size = 4
	err = f.Client.Update(context.TODO(), crdb)
	if err != nil {
		t.Fatal(err)
	}

	err = waitForStatefulSet(t, f.KubeClient, namespace, "crdb-test", 4, retryInterval, timeout)
	if err != nil {
		t.Fatal(err)
	}

}

func waitForStatefulSet(t *testing.T, kubeclient kubernetes.Interface, namespace, name string, replicas int, retryInterval, timeout time.Duration) error {
	err := wait.Poll(retryInterval, timeout, func() (done bool, err error) {
		statefulset, err := kubeclient.AppsV1().StatefulSets(namespace).Get(name, metav1.GetOptions{IncludeUninitialized: true})


		if err != nil {
			if apierrors.IsNotFound(err) {
				t.Logf("Waiting for availability of %s stateful set\n", name)
				return false, nil
			}
			return false, err
		}

		if int(statefulset.Status.Replicas) == replicas {
			return true, nil
		}
		t.Logf("Waiting for full availability of %s stateful set (%d/%d)\n", name, statefulset.Status.Replicas, replicas)
		return false, nil
	})
	if err != nil {
		return err
	}
	t.Logf("StatefulSet available (%d/%d)\n", replicas, replicas)
	return nil
}
