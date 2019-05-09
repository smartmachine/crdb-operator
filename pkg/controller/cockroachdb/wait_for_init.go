package cockroachdb

import (
	corev1 "k8s.io/api/core/v1"
)

const WaitForInitHandler Name = 1000

func init() {
	// Register PodList
	info := NewInfo(nil, waitForInit, &corev1.PodList{})
	err := Register(WaitForInitHandler, info)
	if err != nil {
		panic(err.Error())
	}
}