/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// PodSleuthSpec defines the desired state of PodSleuth
type PodSleuthSpec struct {
	// ReconcileInterval is the duration for periodic reconciliation.
	// Default: 5 minutes
	// +optional
	ReconcileInterval *metav1.Duration `json:"reconcileInterval,omitempty"`

	// PodLabelSelector is a label selector to filter pods across all namespaces.
	// If not specified, monitors all pods in all namespaces.
	// +optional
	PodLabelSelector *metav1.LabelSelector `json:"podLabelSelector,omitempty"`
}

// ContainerError contains detailed error information for a specific container
type ContainerError struct {
	// ContainerName is the name of the container
	ContainerName string `json:"containerName"`

	// Type indicates whether this is a regular container or init container
	Type string `json:"type"` // "container" or "initContainer"

	// State is the current state of the container (waiting, terminated, running)
	State string `json:"state"`

	// Reason is the error reason (CrashLoopBackOff, ImagePullBackOff, etc.)
	Reason string `json:"reason"`

	// Message is the detailed error message
	Message string `json:"message"`

	// ExitCode is the exit code if the container terminated
	// +optional
	ExitCode *int32 `json:"exitCode,omitempty"`

	// RestartCount is the number of times the container has restarted
	RestartCount int32 `json:"restartCount"`

	// Ready indicates if the container is ready
	Ready bool `json:"ready"`
}

// PodCondition represents a pod condition status
type PodCondition struct {
	// Type is the type of condition
	Type string `json:"type"`

	// Status is the status of the condition (True, False, Unknown)
	Status string `json:"status"`

	// Reason is the reason for the condition status
	// +optional
	Reason string `json:"reason,omitempty"`

	// Message is the message describing the condition
	// +optional
	Message string `json:"message,omitempty"`
}

// NonReadyPodInfo contains information about a non-ready pod
type NonReadyPodInfo struct {
	// Name is the name of the pod
	Name string `json:"name"`

	// Namespace is the namespace of the pod
	Namespace string `json:"namespace"`

	// Phase is the current phase of the pod (Pending, Running, Failed, etc.)
	Phase string `json:"phase"`

	// OwnerKind is the kind of the owner (Deployment or StatefulSet)
	// +optional
	OwnerKind string `json:"ownerKind,omitempty"`

	// OwnerName is the name of the owner
	// +optional
	OwnerName string `json:"ownerName,omitempty"`

	// Reason is the primary reason why the pod is not ready (from container status investigation)
	// +optional
	Reason string `json:"reason,omitempty"`

	// Message is the detailed message explaining why the pod is not ready
	// +optional
	Message string `json:"message,omitempty"`

	// ContainerErrors contains detailed error information for each unready container
	// +optional
	ContainerErrors []ContainerError `json:"containerErrors,omitempty"`

	// PodConditions contains all pod conditions for comprehensive status
	// +optional
	PodConditions []PodCondition `json:"podConditions,omitempty"`
}

// PodSleuthStatus defines the observed state of PodSleuth
type PodSleuthStatus struct {
	// NonReadyPods is a dynamic list of non-ready pods
	// +optional
	NonReadyPods []NonReadyPodInfo `json:"nonReadyPods,omitempty"`

	// conditions represent the current state of the PodSleuth resource.
	// Each condition has a unique type and reflects the status of a specific aspect of the resource.
	//
	// Standard condition types include:
	// - "Available": the resource is fully functional
	// - "Progressing": the resource is being created or updated
	// - "Degraded": the resource failed to reach or maintain its desired state
	//
	// The status of each condition is one of True, False, or Unknown.
	// +listType=map
	// +listMapKey=type
	// +optional
	Conditions []metav1.Condition `json:"conditions,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:scope=Cluster

// PodSleuth is the Schema for the podsleuths API
type PodSleuth struct {
	metav1.TypeMeta `json:",inline"`

	// metadata is a standard object metadata
	// +optional
	metav1.ObjectMeta `json:"metadata,omitzero"`

	// spec defines the desired state of PodSleuth
	// +required
	Spec PodSleuthSpec `json:"spec"`

	// status defines the observed state of PodSleuth
	// +optional
	Status PodSleuthStatus `json:"status,omitzero"`
}

// +kubebuilder:object:root=true

// PodSleuthList contains a list of PodSleuth
type PodSleuthList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitzero"`
	Items           []PodSleuth `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PodSleuth{}, &PodSleuthList{})
}
