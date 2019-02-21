package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// CockroachDBClusterSpec defines the desired state of the CockroachDB Cluster
type CockroachDBClusterSpec struct {
	Image string `json:"image,omitempty"`
	// Size is the number of nodes the cluster should have
	Size           int32  `json:"size,omitempty"`
	RequestMemory  string `json:"requestMemory,omitempty"`
	LimitMemory    string `json:"limitMemory,omitempty"`
	StoragePerNode string `json:"storagePerNode,omitempty"`
	MaxUnavailable int    `json:"maxUnavailable,omitempty"`
}

// CockroachDBClientSpec defines the desired state of CockroachDB client
type CockroachDBClientSpec struct {
	Enabled bool `json:"enable,omitempty"`
}

// CockroachDBDashboardSpec defines the desired state of CockroachDB Dashboard
type CockroachDBDashboardSpec struct {
	Enabled  bool  `json:"enable,omitempty"`
	NodePort int32 `json:"nodePort,omitempty"`
}

// CockroachDBSpec defines the desired state of CockroachDB
type CockroachDBSpec struct {
	Cluster   CockroachDBClusterSpec   `json:"cluster,omitempty"`
	Client    CockroachDBClientSpec    `json:"client,omitempty"`
	Dashboard CockroachDBDashboardSpec `json:"dashboard,omitempty"`
}

// CockroachDBStatus defines the observed state of CockroachDB
type CockroachDBStatus struct {
	// Nodes are the names of the cockroachdb pods
	Nodes []CockroachDBNode `json:"nodes,omitempty"`
	State string `json:"state,omitempty"`
}

// CockroachDBNodeList
type CockroachDBNode struct {
	Name    string `json:"name"`
	Ready   bool   `json:"ready"`
	Serving bool   `json:"serving"`
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CockroachDB is the Schema for the cockroachdbs API
type CockroachDB struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   CockroachDBSpec   `json:"spec,omitempty"`
	Status CockroachDBStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// CockroachDBList contains a list of CockroachDB
type CockroachDBList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []CockroachDB `json:"items"`
}

func init() {
	SchemeBuilder.Register(&CockroachDB{}, &CockroachDBList{})
}
