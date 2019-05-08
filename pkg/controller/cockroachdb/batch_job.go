package cockroachdb

import (
	dbv1alpha1 "go.smartmachine.io/crdb-operator/pkg/apis/db/v1alpha1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// jobForCockroachDB returns a cockroachdb Job object
func batchJob(r *ReconcileCockroachDB, m *dbv1alpha1.CockroachDB) runtime.Object {

	reqLogger := log.WithValues("CockroachDB.Meta.Name", m.ObjectMeta.Name, "CockroachDB.Meta.Namespace", m.ObjectMeta.Namespace)
	reqLogger.Info("Reconciling CockroachDB Batch Job")

	ls := labelsForCockroachDB(m.Name)

	one := int32(1)

	dep := &batchv1.Job{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "batch/v1",
			Kind:       "Job",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
			Labels:    ls,
		},
		Spec: batchv1.JobSpec{
			Template: corev1.PodTemplateSpec{
				Spec: corev1.PodSpec{
					ServiceAccountName: m.Name,
					InitContainers: []corev1.Container{{
						Name: "init-certs",
						Image: "smartmachine/cockroach-k8s-request-cert:0.3",
						ImagePullPolicy: corev1.PullIfNotPresent,
						Command: []string{
							"/bin/ash",
							"-ecx",
							"/request-cert -namespace=${POD_NAMESPACE} -certs-dir=/cockroach-certs -type=client " +
								"-user=root -symlink-ca-from=/var/run/secrets/kubernetes.io/serviceaccount/ca.crt " +
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
						Name: "cluster-init",
						Image: m.Spec.Cluster.Image,
						ImagePullPolicy: corev1.PullIfNotPresent,
						VolumeMounts: []corev1.VolumeMount{{
							Name: "client-certs",
							MountPath: "/cockroach-certs",
						}},
						Command: []string{
							"/cockroach/cockroach",
							"init",
							"--certs-dir=/cockroach-certs",
							"--host=" + m.Name + "-0." + m.Name,
						},
					}},
					RestartPolicy: corev1.RestartPolicyNever,
					Volumes: []corev1.Volume{{
						Name: "client-certs",
						VolumeSource: corev1.VolumeSource{
							EmptyDir: &corev1.EmptyDirVolumeSource{},
						},
					}},
				},
			},
			Parallelism: &one,
			Completions: &one,
		},
	}

	// Set CockroachDB instance as the owner and controller
	err := controllerutil.SetControllerReference(m, dep, r.scheme)
	if err != nil {
		reqLogger.Error(err, "Failed to set Controller Reference", "m", m, "dep", dep, "r.scheme", r.scheme)
	}
	return dep
}