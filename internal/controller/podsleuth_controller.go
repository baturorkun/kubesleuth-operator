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

package controller

import (
	"context"
	"fmt"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	log "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"

	infrav1alpha1 "github.com/baturorkun/kubebuilder-demo-operator/api/v1alpha1"
)

// PodSleuthReconciler reconciles a PodSleuth object
type PodSleuthReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	K8sClient kubernetes.Interface
}

// +kubebuilder:rbac:groups=apps.ops.dev,resources=podsleuths,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps.ops.dev,resources=podsleuths/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=apps.ops.dev,resources=podsleuths/finalizers,verbs=update
// +kubebuilder:rbac:groups="",resources=pods,verbs=get;list;watch
// +kubebuilder:rbac:groups="",resources=pods/log,verbs=get;list
// +kubebuilder:rbac:groups="",resources=secrets,verbs=get;list
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list
// +kubebuilder:rbac:groups=apps,resources=replicasets,verbs=get;list

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
func (r *PodSleuthReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	// Create a simple logger without controller-runtime context to avoid verbose fields
	logger := log.Log

	// Fetch the PodSleuth resource
	var podSleuth infrav1alpha1.PodSleuth
	if err := r.Get(ctx, req.NamespacedName, &podSleuth); err != nil {
		logger.Error(err, "unable to fetch PodSleuth")
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// List all pods across all namespaces
	var podList corev1.PodList
	listOptions := []client.ListOption{}

	// Apply pod label selector if specified
	if podSleuth.Spec.PodLabelSelector != nil {
		selector, err := metav1.LabelSelectorAsSelector(podSleuth.Spec.PodLabelSelector)
		if err != nil {
			logger.Error(err, "invalid pod label selector")
			return ctrl.Result{}, err
		}
		listOptions = append(listOptions, client.MatchingLabelsSelector{Selector: selector})
	}

	if err := r.List(ctx, &podList, listOptions...); err != nil {
		logger.Error(err, "unable to list pods")
		return ctrl.Result{}, err
	}

	// Filter non-ready pods and collect information
	var nonReadyPods []infrav1alpha1.NonReadyPodInfo
	for _, pod := range podList.Items {
		// Check if pod is ready
		isReady := false
		for _, condition := range pod.Status.Conditions {
			if condition.Type == corev1.PodReady {
				if condition.Status == corev1.ConditionTrue {
					isReady = true
				}
				break
			}
		}

		// Skip ready pods
		if isReady {
			continue
		}

		// Get owner information
		ownerKind, ownerName := r.getPodOwner(ctx, &pod)

		// Perform comprehensive investigation
		reason, message, containerErrors, conditions := r.investigatePodFailure(&pod)

		// Create NonReadyPodInfo with comprehensive investigation results
		podInfo := infrav1alpha1.NonReadyPodInfo{
			Name:            pod.Name,
			Namespace:       pod.Namespace,
			Phase:           string(pod.Status.Phase),
			OwnerKind:       ownerKind,
			OwnerName:       ownerName,
			Reason:          reason,
			Message:         message,
			ContainerErrors: containerErrors,
			PodConditions:   conditions,
		}

		// Perform log analysis if enabled and pod is running but not ready
		if podSleuth.Spec.LogAnalysis != nil && podSleuth.Spec.LogAnalysis.Enabled {
			if pod.Status.Phase == corev1.PodRunning {
				// Only analyze logs for running pods that are not ready
				logAnalysisResult, err := analyzeLogs(ctx, r.Client, r.K8sClient, &pod, podSleuth.Spec.LogAnalysis)
				if err != nil {
					logger.Info("log analysis failed", "pod", pod.Name, "namespace", pod.Namespace, "error", err)
				} else if logAnalysisResult != nil {
					podInfo.LogAnalysis = logAnalysisResult
					// Append log analysis findings to the message
					if logAnalysisResult.RootCause != "" {
						if podInfo.Message != "" {
							podInfo.Message = podInfo.Message + ". Log analysis: " + logAnalysisResult.RootCause
						} else {
							podInfo.Message = "Log analysis: " + logAnalysisResult.RootCause
						}
					}
					logger.Info("log analysis completed", "pod", pod.Name, "namespace", pod.Namespace, "rootCause", logAnalysisResult.RootCause, "method", logAnalysisResult.Method, "confidence", logAnalysisResult.Confidence, "errorLines", len(logAnalysisResult.ErrorLines))
				} else {
					logger.Info("log analysis returned no results", "pod", pod.Name, "namespace", pod.Namespace)
				}
			}
		}

		nonReadyPods = append(nonReadyPods, podInfo)

		// Log the non-ready pod with detailed information
		logger.Info("Non-ready pod detected",
			"pod", pod.Name,
			"namespace", pod.Namespace,
			"phase", pod.Status.Phase,
			"ownerKind", ownerKind,
			"ownerName", ownerName,
			"reason", podInfo.Reason,
			"message", podInfo.Message,
			"containerErrors", len(podInfo.ContainerErrors),
		)
	}

	// Update status
	podSleuth.Status.NonReadyPods = nonReadyPods
	if err := r.Status().Update(ctx, &podSleuth); err != nil {
		logger.Error(err, "unable to update PodSleuth status")
		return ctrl.Result{}, err
	}

	// Determine reconcile interval
	reconcileInterval := 5 * time.Minute // default
	if podSleuth.Spec.ReconcileInterval != nil {
		reconcileInterval = podSleuth.Spec.ReconcileInterval.Duration
	}

	return ctrl.Result{RequeueAfter: reconcileInterval}, nil
}

// investigatePodFailure performs comprehensive investigation of why a pod is not ready
func (r *PodSleuthReconciler) investigatePodFailure(pod *corev1.Pod) (string, string, []infrav1alpha1.ContainerError, []infrav1alpha1.PodCondition) {
	var containerErrors []infrav1alpha1.ContainerError
	var primaryReason, primaryMessage string

	// Investigate regular containers
	// Check ALL containers, not just unready ones, to catch terminated containers
	for _, containerStatus := range pod.Status.ContainerStatuses {
		// Include containers that are not ready OR are in a failed state (terminated with error)
		shouldInvestigate := !containerStatus.Ready
		if containerStatus.State.Terminated != nil {
			// Always investigate terminated containers, especially if they exited with error
			if containerStatus.State.Terminated.ExitCode != 0 || containerStatus.State.Terminated.Reason == "Error" {
				shouldInvestigate = true
			}
		}

		if shouldInvestigate {
			err := r.investigateContainerStatus(containerStatus, "container")
			containerErrors = append(containerErrors, err)

			// Set primary reason/message from first problematic container
			// Prioritize waiting/terminated states over running but not ready
			// Also prioritize containers with actual error reasons over generic ones
			if primaryReason == "" {
				primaryReason = err.Reason
				primaryMessage = err.Message
			} else if err.State != "running" && err.Reason != "" {
				// Update if we have a more specific error (waiting/terminated) than current
				// Prefer waiting state errors (ImagePullBackOff, ErrImagePull) over terminated
				if err.State == "waiting" || (err.State == "terminated" && primaryReason == "ReadinessProbeFailed") {
					primaryReason = err.Reason
					primaryMessage = err.Message
				}
			}
		}
	}

	// Investigate init containers
	for _, initStatus := range pod.Status.InitContainerStatuses {
		if !initStatus.Ready {
			err := r.investigateContainerStatus(initStatus, "initContainer")
			containerErrors = append(containerErrors, err)

			// Init container failures are critical
			if primaryReason == "" {
				primaryReason = err.Reason
				primaryMessage = err.Message
			}
		}
	}

	// Collect all pod conditions
	var conditions []infrav1alpha1.PodCondition
	for _, condition := range pod.Status.Conditions {
		conditions = append(conditions, infrav1alpha1.PodCondition{
			Type:    string(condition.Type),
			Status:  string(condition.Status),
			Reason:  condition.Reason,
			Message: condition.Message,
		})
	}

	// Fallback: if no container errors found but pod is not ready, use Ready condition
	if primaryReason == "" {
		for _, condition := range pod.Status.Conditions {
			if condition.Type == corev1.PodReady && condition.Status == corev1.ConditionFalse {
				primaryReason = condition.Reason
				primaryMessage = condition.Message
				break
			}
		}
	}

	return primaryReason, primaryMessage, containerErrors, conditions
}

// investigateContainerStatus extracts detailed error information from container status
func (r *PodSleuthReconciler) investigateContainerStatus(containerStatus corev1.ContainerStatus, containerType string) infrav1alpha1.ContainerError {
	err := infrav1alpha1.ContainerError{
		ContainerName: containerStatus.Name,
		Type:          containerType,
		Ready:         containerStatus.Ready,
		RestartCount:  containerStatus.RestartCount,
	}

	// Check current state
	if containerStatus.State.Waiting != nil {
		err.State = "waiting"
		err.Reason = containerStatus.State.Waiting.Reason
		err.Message = containerStatus.State.Waiting.Message
	} else if containerStatus.State.Terminated != nil {
		err.State = "terminated"
		err.Reason = containerStatus.State.Terminated.Reason
		err.Message = containerStatus.State.Terminated.Message
		err.ExitCode = &containerStatus.State.Terminated.ExitCode

		// If no message but we have reason and exit code, generate a meaningful message
		if err.Message == "" {
			if err.ExitCode != nil {
				err.Message = fmt.Sprintf("Container terminated with reason '%s' (exit code: %d)",
					err.Reason, *err.ExitCode)
			} else {
				err.Message = fmt.Sprintf("Container terminated with reason '%s'", err.Reason)
			}
		}
	} else if containerStatus.State.Running != nil {
		err.State = "running"
		// Container is running but not ready - likely readiness probe failure
		if !containerStatus.Ready {
			err.Reason = "ReadinessProbeFailed"
			err.Message = "Container is running but readiness probe is failing"
		}
	}

	// Check last termination state (for crash loops)
	if containerStatus.LastTerminationState.Terminated != nil {
		if err.Reason == "" || err.Reason == "CrashLoopBackOff" {
			// Use last termination info if current state doesn't have details
			if err.Message == "" {
				err.Message = fmt.Sprintf("Container exited with code %d: %s",
					containerStatus.LastTerminationState.Terminated.ExitCode,
					containerStatus.LastTerminationState.Terminated.Reason)
			}
			if err.ExitCode == nil {
				exitCode := containerStatus.LastTerminationState.Terminated.ExitCode
				err.ExitCode = &exitCode
			}
		}
	}

	// Final fallback: if we still don't have a reason or message, generate one
	if err.Reason == "" {
		if err.State == "terminated" && err.ExitCode != nil {
			err.Reason = "ContainerTerminated"
			if err.Message == "" {
				err.Message = fmt.Sprintf("Container terminated with exit code %d", *err.ExitCode)
			}
		} else if err.State == "waiting" {
			err.Reason = "ContainerWaiting"
			if err.Message == "" {
				err.Message = "Container is waiting to start"
			}
		} else if !err.Ready {
			err.Reason = "ContainerNotReady"
			if err.Message == "" {
				err.Message = "Container is not ready"
			}
		}
	}

	return err
}

// getPodOwner finds the owner Deployment or StatefulSet for a pod
func (r *PodSleuthReconciler) getPodOwner(ctx context.Context, pod *corev1.Pod) (string, string) {
	for _, ownerRef := range pod.OwnerReferences {
		if ownerRef.Kind == "ReplicaSet" {
			// Get ReplicaSet to find its owner Deployment
			var rs appsv1.ReplicaSet
			if err := r.Get(ctx, types.NamespacedName{
				Name:      ownerRef.Name,
				Namespace: pod.Namespace,
			}, &rs); err != nil {
				continue
			}

			// Check ReplicaSet's owner
			for _, rsOwnerRef := range rs.OwnerReferences {
				if rsOwnerRef.Kind == "Deployment" {
					return "Deployment", rsOwnerRef.Name
				}
			}
		} else if ownerRef.Kind == "StatefulSet" {
			return "StatefulSet", ownerRef.Name
		} else if ownerRef.Kind == "Deployment" {
			// Direct Deployment owner (uncommon but possible)
			return "Deployment", ownerRef.Name
		}
	}

	return "", ""
}

// findObjectsForPod maps pod changes to PodSleuth resources
func (r *PodSleuthReconciler) findObjectsForPod(ctx context.Context, pod client.Object) []reconcile.Request {
	var podSleuthList infrav1alpha1.PodSleuthList
	if err := r.List(ctx, &podSleuthList); err != nil {
		return []reconcile.Request{}
	}

	var requests []reconcile.Request
	for _, podSleuth := range podSleuthList.Items {
		// Check if pod matches the label selector if specified
		if podSleuth.Spec.PodLabelSelector != nil {
			selector, err := metav1.LabelSelectorAsSelector(podSleuth.Spec.PodLabelSelector)
			if err != nil {
				continue
			}
			if !selector.Matches(labels.Set(pod.GetLabels())) {
				continue
			}
		}

		requests = append(requests, reconcile.Request{
			NamespacedName: client.ObjectKey{
				Name: podSleuth.Name,
			},
		})
	}

	return requests
}

// SetupWithManager sets up the controller with the Manager.
func (r *PodSleuthReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrav1alpha1.PodSleuth{}).
		Watches(
			&corev1.Pod{},
			handler.EnqueueRequestsFromMapFunc(r.findObjectsForPod),
		).
		Complete(r)
}
