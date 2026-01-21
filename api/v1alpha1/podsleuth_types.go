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
	corev1 "k8s.io/api/core/v1"
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

	// LogAnalysis enables log analysis for running but not ready pods
	// +optional
	LogAnalysis *LogAnalysisConfig `json:"logAnalysis,omitempty"`
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

// LogAnalysisConfig defines configuration for log analysis
type LogAnalysisConfig struct {
	// Enabled enables log analysis for non-ready pods
	Enabled bool `json:"enabled"`

	// Method specifies the analysis method: "pattern" or "ai"
	// Deprecated: Use Methods instead for multiple method support
	// Default: "pattern"
	// +optional
	Method string `json:"method,omitempty"`

	// Methods specifies the analysis methods to run in order (e.g., ["pattern", "ai"])
	// Methods are executed sequentially in the order specified
	// Example: ["pattern", "ai"] runs pattern first, then AI
	// Default: ["pattern"]
	// +optional
	Methods []string `json:"methods,omitempty"`

	// CacheEnabled enables caching of analysis results to avoid re-analyzing on every reconcile
	// Results are cached per pod (keyed by UID + restart count)
	// Cache is invalidated when pod restarts or after TTL expires
	// Default: true
	// +optional
	CacheEnabled *bool `json:"cacheEnabled,omitempty"`

	// CacheTTL is the duration to cache analysis results before re-analyzing
	// Default: 5m
	// +optional
	CacheTTL *metav1.Duration `json:"cacheTTL,omitempty"`

	// LinesToAnalyze is the number of recent log lines to fetch and analyze
	// Default: 100
	// +optional
	LinesToAnalyze *int32 `json:"linesToAnalyze,omitempty"`

	// FilterErrorsOnly if true, filters error/warning lines from the last LinesToAnalyze lines
	// Process: 1) Fetch last LinesToAnalyze lines, 2) Filter for errors/warnings, 3) Analyze filtered lines
	// Default: true
	// +optional
	FilterErrorsOnly *bool `json:"filterErrorsOnly,omitempty"`

	// Patterns defines custom error patterns for pattern matching method
	// Deprecated: Use MethodConfigs with PatternConfig instead for cleaner structure
	// If not specified, default patterns will be used (connection errors, service unavailable, etc.)
	// Each pattern should be a regex that matches error messages
	// +optional
	Patterns []ErrorPattern `json:"patterns,omitempty"`

	// AIEndpoint is the URL endpoint for AI analysis (required if "ai" in methods)
	// Deprecated: Use MethodConfigs with AIConfig instead for cleaner structure
	// Examples:
	//   - OpenAI: "https://api.openai.com/v1/chat/completions"
	//   - Ollama: "http://localhost:11434/api/generate"
	// +optional
	AIEndpoint string `json:"aiEndpoint,omitempty"`

	// AIFormat specifies the API format to use: "openai", "anthropic", "ollama", or "generic"
	// Deprecated: Use MethodConfigs with AIConfig instead
	// Default: "openai"
	// +optional
	AIFormat string `json:"aiFormat,omitempty"`

	// AIModel specifies the model name to use for AI analysis
	// Deprecated: Use MethodConfigs with AIConfig instead
	// Examples: "gpt-4", "qwen3:8b", "claude-3-opus"
	// +optional
	AIModel string `json:"aiModel,omitempty"`

	// AIAPIKey is the API key for AI analysis
	// Deprecated: Use MethodConfigs with AIConfig instead
	// Should be stored as a Kubernetes Secret reference
	// +optional
	AIAPIKey *corev1.SecretKeySelector `json:"aiApiKey,omitempty"`

	// AIAuthHeader specifies the HTTP header name for authentication
	// Deprecated: Use MethodConfigs with AIConfig instead
	// Default: "Authorization"
	// +optional
	AIAuthHeader string `json:"aiAuthHeader,omitempty"`

	// AIAuthPrefix specifies the prefix for the auth header value (e.g., "Bearer", "ApiKey")
	// Deprecated: Use MethodConfigs with AIConfig instead
	// Default: "Bearer"
	// +optional
	AIAuthPrefix string `json:"aiAuthPrefix,omitempty"`

	// MethodConfigs defines method-specific configurations in order of execution
	// This is the NEW PREFERRED way to configure log analysis methods.
	// Each method has its own configuration block, avoiding parameter mixing.
	// Example:
	//   methodConfigs:
	//     - type: "pattern"
	//       patternConfig:
	//         patterns: [...]
	//     - type: "ai"
	//       aiConfig:
	//         endpoint: "..."
	//         model: "..."
	// +optional
	MethodConfigs []MethodConfig `json:"methodConfigs,omitempty"`
}

// MethodConfig defines configuration for a specific analysis method
type MethodConfig struct {
	// Type specifies the analysis method type: "pattern" or "ai"
	// +kubebuilder:validation:Enum=pattern;ai
	Type string `json:"type"`

	// PatternConfig contains pattern-specific configuration (used when type is "pattern")
	// +optional
	PatternConfig *PatternConfig `json:"patternConfig,omitempty"`

	// AIConfig contains AI-specific configuration (used when type is "ai")
	// +optional
	AIConfig *AIConfig `json:"aiConfig,omitempty"`
}

// PatternConfig defines configuration for pattern-based analysis
type PatternConfig struct {
	// Patterns defines custom error patterns for pattern matching
	// If not specified, default patterns will be used
	// +optional
	Patterns []ErrorPattern `json:"patterns,omitempty"`
}

// AIConfig defines configuration for AI-based analysis
type AIConfig struct {
	// Endpoint is the URL endpoint for AI analysis
	// Examples:
	//   - OpenAI: "https://api.openai.com/v1/chat/completions"
	//   - Ollama: "http://localhost:11434/api/generate"
	Endpoint string `json:"endpoint"`

	// Format specifies the API format: "openai", "anthropic", "ollama", or "generic"
	// Default: "openai"
	// +optional
	Format string `json:"format,omitempty"`

	// Model specifies the model name to use
	// Examples: "gpt-4", "qwen3:8b", "claude-3-opus"
	// +optional
	Model string `json:"model,omitempty"`

	// APIKeySecretRef references a secret containing the API key
	// +optional
	APIKeySecretRef *corev1.SecretKeySelector `json:"apiKeySecretRef,omitempty"`

	// AuthHeader specifies the HTTP header name for authentication
	// Default: "Authorization"
	// +optional
	AuthHeader string `json:"authHeader,omitempty"`

	// AuthPrefix specifies the prefix for the auth header value
	// Default: "Bearer"
	// +optional
	AuthPrefix string `json:"authPrefix,omitempty"`

	// Timeout specifies the timeout for the AI request
	// Default: 60s
	// +optional
	Timeout *metav1.Duration `json:"timeout,omitempty"`
}

// ErrorPattern defines a pattern to match error messages in logs
type ErrorPattern struct {
	// Name is a descriptive name for this pattern (e.g., "KafkaConnectionError")
	Name string `json:"name"`

	// Pattern is the regex pattern to match against log lines
	Pattern string `json:"pattern"`

	// RootCause is the root cause message to report when this pattern matches
	// If empty, the matched log line will be used as the root cause
	// +optional
	RootCause string `json:"rootCause,omitempty"`

	// Priority determines which pattern to use if multiple patterns match
	// Higher priority patterns are preferred. Default: 0
	// +optional
	Priority int32 `json:"priority,omitempty"`
}

// PatternError contains detailed error information for a specific pattern
type PatternError struct {
	// PatternName is the name of the pattern
	PatternName string `json:"patternName"`

	// MatchedLine is the log line that matched the pattern
	MatchedLine string `json:"matchedLine"`

	// Reason is the reason why the pattern matched
	Reason string `json:"reason"`

	// Message is the detailed message explaining the pattern match
	Message string `json:"message"`
}

// PatternAnalysisResult contains pattern-specific analysis results
type PatternAnalysisResult struct {
	// MatchedPattern is the name of the pattern that matched
	MatchedPattern string `json:"matchedPattern,omitempty"`

	// Priority is the priority of the matched pattern
	Priority int32 `json:"priority,omitempty"`

	// RootCause is the root cause from pattern matching
	RootCause string `json:"rootCause,omitempty"`

	// Confidence is the confidence level (0-100) of the pattern match
	Confidence int32 `json:"confidence,omitempty"`

	// Error contains any error message if pattern analysis failed
	// +optional
	Error string `json:"error,omitempty"`
}

// AIAnalysisResult contains AI-specific analysis results
type AIAnalysisResult struct {
	// Model is the AI model used for analysis
	Model string `json:"model,omitempty"`

	// RootCause is the root cause identified by AI
	RootCause string `json:"rootCause,omitempty"`

	// Confidence is the confidence level (0-100) from AI analysis
	Confidence int32 `json:"confidence,omitempty"`

	// Error contains any error message if AI analysis failed
	// +optional
	Error string `json:"error,omitempty"`
}

// LogAnalysisResult contains results from log analysis
type LogAnalysisResult struct {
	// RootCause is the identified root cause from log analysis (merged from all methods)
	RootCause string `json:"rootCause,omitempty"`

	// Confidence is the confidence level (0-100) of the analysis (merged from all methods)
	Confidence int32 `json:"confidence,omitempty"`

	// Method used for analysis: "pattern" or "ai"
	// Deprecated: Use Methods instead for multiple method support
	Method string `json:"method,omitempty"`

	// Methods used for analysis in execution order (e.g., ["pattern", "ai"])
	Methods []string `json:"methods,omitempty"`

	// MatchedPattern is the name of the pattern that matched (for pattern analysis)
	// Used internally, prefer PatternResult.MatchedPattern
	// +optional
	MatchedPattern string `json:"matchedPattern,omitempty"`

	// Priority is the priority of the matched pattern (for pattern analysis)
	// Used internally, prefer PatternResult.Priority
	// +optional
	Priority int32 `json:"priority,omitempty"`

	// Model is the AI model used (for AI analysis)
	// Used internally, prefer AIResult.Model
	// +optional
	Model string `json:"model,omitempty"`

	// PatternResult contains pattern-specific analysis details
	// +optional
	PatternResult *PatternAnalysisResult `json:"patternResult,omitempty"`

	// AIResult contains AI-specific analysis details
	// +optional
	AIResult *AIAnalysisResult `json:"aiResult,omitempty"`

	// ErrorLines contains the error lines that led to this conclusion
	ErrorLines []string `json:"errorLines,omitempty"`

	// AnalyzedAt is when the analysis was performed
	AnalyzedAt metav1.Time `json:"analyzedAt,omitempty"`

	// CachedAt is when the result was cached (if caching is enabled)
	// +optional
	CachedAt metav1.Time `json:"cachedAt,omitempty"`

	// CacheExpiresAt is when the cached result will expire (if caching is enabled)
	// +optional
	CacheExpiresAt *metav1.Time `json:"cacheExpiresAt,omitempty"`
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

	// LogAnalysis contains results from log analysis if enabled
	// +optional
	LogAnalysis *LogAnalysisResult `json:"logAnalysis,omitempty"`
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
