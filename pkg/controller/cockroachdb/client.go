package cockroachdb

import (
	dbv1alpha1 "go.smartmachine.io/crdb-operator/pkg/apis/db/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const ClientHandler Name = 1200

func init() {
	// Register Client
	info := NewInfo(ClientPodForCockroachDB, createIfNotExist, &corev1.Pod{})
	info.Postfix = "-client"
	info.SpecConditional = "Spec.Client.Enabled"
	info.SuppressStatus = true
	err := Register(ClientHandler, info)
	if err != nil {
		panic(err.Error())
	}
}

// clientPodForCockroachDB returns a cockroachdb ClientPod object
func ClientPodForCockroachDB(r *ReconcileCockroachDB, m *dbv1alpha1.CockroachDB) interface{} {

	reqLogger := log.WithValues("CockroachDB.Meta.Name", m.ObjectMeta.Name, "CockroachDB.Meta.Namespace", m.ObjectMeta.Namespace)
	reqLogger.Info("Reconciling CockroachDB")

	ls := labelsForCockroachDB(m.Name + "-client")
	terminationGracePeriodSeconds := int64(0)

	dep := &corev1.Pod{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "v1",
			Kind:       "Pod",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name + "-client",
			Namespace: m.Namespace,
			Labels:    ls,
		},
		Spec: corev1.PodSpec{
			ServiceAccountName: m.Name,
			InitContainers: []corev1.Container{{
				Name: "init-certs",
				Image: "smartmachine/cockroach-k8s-request-cert:0.3",
				ImagePullPolicy: corev1.PullIfNotPresent,
				Command: []string{
					"/bin/ash",
					"-ecx",
					"/request-cert -namespace=${POD_NAMESPACE} -certs-dir=/cockroach-certs " +
						"-type=client -user=root -symlink-ca-from=/var/run/secrets/kubernetes.io/serviceaccount/ca.crt " +
						"-cluster=" + m.Name,
				},
				Env: []corev1.EnvVar{{
					Name: "POD_NAMESPACE",
					ValueFrom: &corev1.EnvVarSource{
						FieldRef: &corev1.ObjectFieldSelector{
							FieldPath: "metadata.namespace",
						},
					},
				}},
				VolumeMounts: []corev1.VolumeMount{{
					Name: "client-certs",
					MountPath: "/cockroach-certs",
				}},
			}},
			Containers: []corev1.Container{{
				Name: "cockroachdb-client",
				Image: m.Spec.Cluster.Image,
				ImagePullPolicy: corev1.PullIfNotPresent,
				VolumeMounts: []corev1.VolumeMount{{
					Name: "client-certs",
					MountPath: "/cockroach-certs",
				}},
				Command: []string{
					"sleep",
					"2147483648",
				},
			}},
			TerminationGracePeriodSeconds: &terminationGracePeriodSeconds,
			Volumes: []corev1.Volume{{
				Name: "client-certs",
				VolumeSource: corev1.VolumeSource{
					EmptyDir: &corev1.EmptyDirVolumeSource{},
				},
			}},
		},
	}

	// Set CockroachDB instance as the owner and controller
	err := controllerutil.SetControllerReference(m, dep, r.scheme)
	if err != nil {
		reqLogger.Error(err, "Failed to set Controller Reference", "m", m, "dep", dep, "r.scheme", r.scheme)
	}
	return dep
}