package cockroachdb

import (
	dbv1alpha1 "go.smartmachine.io/crdb-operator/pkg/apis/db/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

// statefulSetForCockroachDB returns a cockroachdb StatefulSet object
func statefulSet(r *ReconcileCockroachDB, m *dbv1alpha1.CockroachDB) runtime.Object {

	reqLogger := log.WithValues("CockroachDB.Meta.Name", m.ObjectMeta.Name, "CockroachDB.Meta.Namespace", m.ObjectMeta.Namespace)
	reqLogger.Info("Reconciling CockroachDB")

	requestMemory, _ := resource.ParseQuantity(m.Spec.Cluster.RequestMemory)
	limitMemory, _ := resource.ParseQuantity(m.Spec.Cluster.LimitMemory)

	storagePerNode, _ := resource.ParseQuantity(m.Spec.Cluster.StoragePerNode)

	terminationGrace := int64(60)

	dep := &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			APIVersion: "apps/v1",
			Kind:       "StatefulSet",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      m.Name,
			Namespace: m.Namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: labelsForCockroachDB(m.Name),
			},
			ServiceName: m.Name,
			Replicas:    &m.Spec.Cluster.Size,
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": m.Name,
					},
				},
				Spec: corev1.PodSpec{
					ServiceAccountName: m.Name,
					InitContainers: []corev1.Container{{
						Name:            "init-certs",
						Image:           "smartmachine/cockroach-k8s-request-cert:0.3",
						ImagePullPolicy: corev1.PullIfNotPresent,
						Command: []string{
							"/bin/ash",
							"-ecx",
							"/request-cert -namespace=${POD_NAMESPACE} -certs-dir=/cockroach-certs -type=node " +
								"-addresses=192.168.64.2,localhost,127.0.0.1,$(hostname -f)," +
								"$(hostname -f|cut -f 1-2 -d '.')," + m.Name + "-public," +
								m.Name + "-public.$(hostname -f|cut -f 3- -d '.') " +
								"-symlink-ca-from=/var/run/secrets/kubernetes.io/serviceaccount/ca.crt " +
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
							Name:      "certs",
							MountPath: "/cockroach-certs",
						}},
					}},
					Affinity: &corev1.Affinity{
						PodAntiAffinity: &corev1.PodAntiAffinity{
							PreferredDuringSchedulingIgnoredDuringExecution: []corev1.WeightedPodAffinityTerm{{
								Weight: 100,
								PodAffinityTerm: corev1.PodAffinityTerm{
									LabelSelector: &metav1.LabelSelector{
										MatchExpressions: []metav1.LabelSelectorRequirement{{
											Key:      "app",
											Operator: metav1.LabelSelectorOpIn,
											Values:   []string{m.Name},
										}},
									},
									TopologyKey: "kubernetes.io/hostname",
								},
							}},
						},
					},
					Containers: []corev1.Container{{
						Name:            "cockroachdb",
						Image:           m.Spec.Cluster.Image,
						ImagePullPolicy: corev1.PullIfNotPresent,
						Ports: []corev1.ContainerPort{
							{
								ContainerPort: 26257,
								Name:          "grpc",
							},
							{
								ContainerPort: 8080,
								Name:          "http",
							},
						},
						Resources: corev1.ResourceRequirements{
							Limits: corev1.ResourceList{
								corev1.ResourceMemory: limitMemory,
							},
							Requests: corev1.ResourceList{
								corev1.ResourceMemory: requestMemory,
							},
						},
						LivenessProbe: &corev1.Probe{
							Handler: corev1.Handler{
								HTTPGet: &corev1.HTTPGetAction{
									Path:   "/health",
									Port:   intstr.FromString("http"),
									Scheme: corev1.URISchemeHTTPS,
								},
							},
							InitialDelaySeconds: 30,
							PeriodSeconds:       5,
						},
						ReadinessProbe: &corev1.Probe{
							Handler: corev1.Handler{
								HTTPGet: &corev1.HTTPGetAction{
									Path:   "/health?ready=1",
									Port:   intstr.FromString("http"),
									Scheme: corev1.URISchemeHTTPS,
								},
							},
							InitialDelaySeconds: 10,
							PeriodSeconds:       5,
							FailureThreshold:    2,
						},
						VolumeMounts: []corev1.VolumeMount{
							{
								Name:      "datadir",
								MountPath: "/cockroach/cockroach-data",
							}, {
								Name:      "certs",
								MountPath: "/cockroach/cockroach-certs",
							},
						},
						Env: []corev1.EnvVar{{
							Name:  "COCKROACH_CHANNEL",
							Value: "kubernetes-secure",
						}},
						Command: []string{
							"/bin/bash",
							"-ecx",
							"exec /cockroach/cockroach start --logtostderr --certs-dir /cockroach/cockroach-certs " +
								"--advertise-host $(hostname -f) --http-host 0.0.0.0 " +
								"--join " + m.Name + "-0." + m.Name + "," + m.Name + "-1." + m.Name + "," + m.Name + "-2." + m.Name + " " +
								"--cache 25% --max-sql-memory 25%",
						},
					}},
					TerminationGracePeriodSeconds: &terminationGrace,
					Volumes: []corev1.Volume{
						{
							Name: "datadir",
							VolumeSource: corev1.VolumeSource{
								PersistentVolumeClaim: &corev1.PersistentVolumeClaimVolumeSource{
									ClaimName: "datadir",
								},
							},
						}, {
							Name: "certs",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
					},
				},
			},
			PodManagementPolicy: appsv1.ParallelPodManagement,
			UpdateStrategy: appsv1.StatefulSetUpdateStrategy{
				Type: appsv1.RollingUpdateStatefulSetStrategyType,
			},
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{{
				ObjectMeta: metav1.ObjectMeta{
					Name: "datadir",
				},
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{
						corev1.ReadWriteOnce,
					},
					Resources: corev1.ResourceRequirements{
						Requests: corev1.ResourceList{
							corev1.ResourceStorage: storagePerNode,
						},
					},
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