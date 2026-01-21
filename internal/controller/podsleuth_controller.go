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
	"strings"
	"sync"
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

// CachedAnalysisResult represents a cached log analysis result for a pod
type CachedAnalysisResult struct {
	PodUID       types.UID
	PodNamespace string
	PodName      string
	RestartCount int32
	Result       *infrav1alpha1.LogAnalysisResult
	CachedAt     time.Time
	ExpiresAt    time.Time
}

// PodSleuthReconciler reconciles a PodSleuth object
type PodSleuthReconciler struct {
	client.Client
	Scheme    *runtime.Scheme
	K8sClient kubernetes.Interface

	// Cache for log analysis results
	analysisCache    map[string]*CachedAnalysisResult
	analysisCacheMux sync.RWMutex

	OperatorStartTime time.Time
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

	// Check for force-refresh annotations
	globalForceRefresh := false
	targetForcePod := ""
	if podSleuth.Annotations != nil {
		if _, exists := podSleuth.Annotations["kubesleuth.io/force-refresh"]; exists {
			globalForceRefresh = true
			logger.Info("force-refresh annotation detected - bypassing cache for all pods")
		}
		if v, exists := podSleuth.Annotations["kubesleuth.io/force-refresh-pod"]; exists {
			targetForcePod = strings.TrimSpace(v)
			logger.Info("force-refresh annotation detected for specific pod", "pod", targetForcePod)
		}
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

		// Perform log analysis if enabled and pod is not ready
		if podSleuth.Spec.LogAnalysis != nil && podSleuth.Spec.LogAnalysis.Enabled {
			// Run analysis for any non-ready pod except Succeeded (which is already finished)
			if pod.Status.Phase != corev1.PodSucceeded {
				// Get cache configuration
				cacheTTL := 5 * time.Minute // default
				if podSleuth.Spec.LogAnalysis.CacheTTL != nil {
					cacheTTL = podSleuth.Spec.LogAnalysis.CacheTTL.Duration
				}

				cacheEnabled := true
				if podSleuth.Spec.LogAnalysis.CacheEnabled != nil {
					cacheEnabled = *podSleuth.Spec.LogAnalysis.CacheEnabled
				}

				var logAnalysisResult *infrav1alpha1.LogAnalysisResult

				// Use global or pod-specific force refresh flag
				podKey := fmt.Sprintf("%s/%s", pod.Namespace, pod.Name)
				forceRefresh := globalForceRefresh || (targetForcePod != "" && targetForcePod == podKey)
				if targetForcePod != "" {
					logger.Info("checking force refresh for pod", "currentPod", podKey, "targetPod", targetForcePod, "match", targetForcePod == podKey, "forceRefresh", forceRefresh)
				}

				// Try to get cached result if caching is enabled (but skip cache on first reconcile or force refresh)
				if cacheEnabled && !forceRefresh {
					logAnalysisResult = r.getCachedAnalysis(&pod, cacheTTL)
					if logAnalysisResult != nil {
						logger.Info("using cached log analysis", "pod", pod.Name, "namespace", pod.Namespace, "cachedAt", logAnalysisResult.CachedAt)
					}
				}

				if logAnalysisResult == nil {
					if forceRefresh {
						logger.Info("force refresh requested - running log analysis immediately", "pod", pod.Name, "namespace", pod.Namespace)
						// Ensure at least 1 second passes to guarantee a new timestamp for the dashboard to detect
						time.Sleep(1100 * time.Millisecond)
					}

					result, err := analyzeLogs(ctx, r.Client, r.K8sClient, &pod, podSleuth.Spec.LogAnalysis)
					if err != nil {
						logger.Info("log analysis failed", "pod", pod.Name, "namespace", pod.Namespace, "error", err)
						// Create failure result so the dashboard polling detects completion
						result = &infrav1alpha1.LogAnalysisResult{
							RootCause:  fmt.Sprintf("Analysis Failed: %v", err),
							Methods:    []string{"failed"},
							AnalyzedAt: metav1.Now(),
							Confidence: 0,
						}
					}

					if result != nil {

						logger.Info("log analysis successful", "pod", pod.Name, "newAnalyzedAt", result.AnalyzedAt, "timestamp", result.AnalyzedAt.Time.Unix())
						logAnalysisResult = result
						// Cache the result if caching is enabled
						if cacheEnabled {
							r.setCachedAnalysis(&pod, result, cacheTTL)
							logger.Info("log analysis completed and cached", "pod", pod.Name, "namespace", pod.Namespace)
						} else {
							logger.Info("log analysis completed (no cache)", "pod", pod.Name, "namespace", pod.Namespace)
						}
					}
				}

				// Use the analysis result (cached or fresh)
				if logAnalysisResult != nil {
					podInfo.LogAnalysis = logAnalysisResult

					// Append log analysis findings to the message
					if logAnalysisResult.RootCause != "" {
						if podInfo.Message != "" {
							podInfo.Message = podInfo.Message + ". Log analysis: " + logAnalysisResult.RootCause
						} else {
							podInfo.Message = "Log analysis: " + logAnalysisResult.RootCause
						}
					}
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

	// Clean up cache for pods that are no longer in the non-ready list
	currentPods := make(map[string]bool)
	for _, pod := range podList.Items {
		if !isPodReady(&pod) {
			currentPods[getCacheKey(&pod)] = true
		}
	}
	r.cleanupCache(currentPods)

	// Update status
	podSleuth.Status.NonReadyPods = nonReadyPods
	if err := r.Status().Update(ctx, &podSleuth); err != nil {
		logger.Error(err, "unable to update PodSleuth status")
		return ctrl.Result{}, err
	}

	// If force refresh was active and status update succeeded, remove the annotations
	if globalForceRefresh || targetForcePod != "" {
		// Fetch latest version to avoid conflict
		if err := r.Get(ctx, req.NamespacedName, &podSleuth); err == nil {
			changed := false
			if podSleuth.Annotations != nil {
				if _, exists := podSleuth.Annotations["kubesleuth.io/force-refresh"]; exists {
					delete(podSleuth.Annotations, "kubesleuth.io/force-refresh")
					changed = true
				}
				if _, exists := podSleuth.Annotations["kubesleuth.io/force-refresh-pod"]; exists {
					delete(podSleuth.Annotations, "kubesleuth.io/force-refresh-pod")
					changed = true
				}
			}

			if changed {
				if err := r.Update(ctx, &podSleuth); err != nil {
					logger.Error(err, "failed to remove force-refresh annotation(s) after analysis")
				} else {
					logger.Info("cleared force-refresh annotations after successful analysis")
				}
			}
		}
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

// getCacheKey generates a cache key for a pod based on UID and restart count
func getCacheKey(pod *corev1.Pod) string {
	// Get the highest restart count from all containers
	maxRestartCount := int32(0)
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.RestartCount > maxRestartCount {
			maxRestartCount = cs.RestartCount
		}
	}
	for _, cs := range pod.Status.InitContainerStatuses {
		if cs.RestartCount > maxRestartCount {
			maxRestartCount = cs.RestartCount
		}
	}

	return fmt.Sprintf("%s/%s/%s/%d", pod.Namespace, pod.Name, pod.UID, maxRestartCount)
}

// isPodReady checks if a pod is ready
func isPodReady(pod *corev1.Pod) bool {
	for _, condition := range pod.Status.Conditions {
		if condition.Type == corev1.PodReady {
			return condition.Status == corev1.ConditionTrue
		}
	}
	return false
}

// getCachedAnalysis retrieves a cached analysis result if it exists and hasn't expired
func (r *PodSleuthReconciler) getCachedAnalysis(pod *corev1.Pod, cacheTTL time.Duration) *infrav1alpha1.LogAnalysisResult {
	r.analysisCacheMux.RLock()
	defer r.analysisCacheMux.RUnlock()

	if r.analysisCache == nil {
		return nil
	}

	cacheKey := getCacheKey(pod)
	cached, exists := r.analysisCache[cacheKey]
	if !exists {
		return nil
	}

	// Check if cache has expired
	if time.Now().After(cached.ExpiresAt) {
		return nil
	}

	return cached.Result
}

// setCachedAnalysis stores an analysis result in the cache
func (r *PodSleuthReconciler) setCachedAnalysis(pod *corev1.Pod, result *infrav1alpha1.LogAnalysisResult, cacheTTL time.Duration) {
	r.analysisCacheMux.Lock()
	defer r.analysisCacheMux.Unlock()

	if r.analysisCache == nil {
		r.analysisCache = make(map[string]*CachedAnalysisResult)
	}

	// Get the highest restart count
	maxRestartCount := int32(0)
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.RestartCount > maxRestartCount {
			maxRestartCount = cs.RestartCount
		}
	}
	for _, cs := range pod.Status.InitContainerStatuses {
		if cs.RestartCount > maxRestartCount {
			maxRestartCount = cs.RestartCount
		}
	}

	cacheKey := getCacheKey(pod)
	now := time.Now()
	expiresAt := now.Add(cacheTTL)

	// Set CachedAt and CacheExpiresAt timestamps in the result
	result.CachedAt = metav1.NewTime(now)
	cacheExpiresAtTime := metav1.NewTime(expiresAt)
	result.CacheExpiresAt = &cacheExpiresAtTime

	r.analysisCache[cacheKey] = &CachedAnalysisResult{
		PodUID:       pod.UID,
		PodNamespace: pod.Namespace,
		PodName:      pod.Name,
		RestartCount: maxRestartCount,
		Result:       result,
		CachedAt:     now,
		ExpiresAt:    expiresAt,
	}
}

// cleanupCache removes stale cache entries for pods that no longer exist or are ready
func (r *PodSleuthReconciler) cleanupCache(currentPods map[string]bool) {
	r.analysisCacheMux.Lock()
	defer r.analysisCacheMux.Unlock()

	if r.analysisCache == nil {
		return
	}

	// Remove entries for pods that are no longer in the non-ready list
	for key := range r.analysisCache {
		if !currentPods[key] {
			delete(r.analysisCache, key)
		}
	}
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
