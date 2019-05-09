package cockroachdb

import (
	corev1 "k8s.io/api/core/v1"
)

const TagAsServingHandler Name = 1200

func init() {
	// Register PodList
	info := NewInfo(nil, waitForServing, &corev1.PodList{})
	err := Register(TagAsServingHandler, info)
	if err != nil {
		panic(err.Error())
	}
}