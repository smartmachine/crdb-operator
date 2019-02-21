package cockroachdb

import (
	corev1 "k8s.io/api/core/v1"
)

const UpdateStatusHandler Name = 1000

func init() {
	// Register PodList
	info := NewInfo(nil, updateStatus, &corev1.PodList{})
	err := Register(UpdateStatusHandler, info)
	if err != nil {
		panic(err.Error())
	}
}