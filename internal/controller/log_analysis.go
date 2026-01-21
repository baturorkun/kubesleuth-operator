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
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sort"
	"strings"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes"
	"sigs.k8s.io/controller-runtime/pkg/client"
	log "sigs.k8s.io/controller-runtime/pkg/log"

	infrav1alpha1 "github.com/baturorkun/kubebuilder-demo-operator/api/v1alpha1"
)

// DefaultPattern defines a built-in error pattern
type DefaultPattern struct {
	Name      string
	Pattern   *regexp.Regexp
	RootCause string
	Priority  int32
}

// getDefaultPatterns returns built-in error patterns
func getDefaultPatterns() []DefaultPattern {
	patterns := []DefaultPattern{
		{
			Name:      "ConnectionRefused",
			Pattern:   regexp.MustCompile(`(?i)(connection refused|connection reset|connection closed)`),
			RootCause: "Connection refused - service may be down or unreachable",
			Priority:  10,
		},
		{
			Name:      "ConnectionTimeout",
			Pattern:   regexp.MustCompile(`(?i)(connection timeout|timeout|timed out)`),
			RootCause: "Connection timeout - service may be slow or unreachable",
			Priority:  10,
		},
		{
			Name:      "DialTCP",
			Pattern:   regexp.MustCompile(`(?i)(dial tcp|failed to connect)`),
			RootCause: "Network connection failed - unable to reach service",
			Priority:  10,
		},
		{
			Name:      "ServiceUnavailable",
			Pattern:   regexp.MustCompile(`(?i)(service unavailable|503|503 service unavailable)`),
			RootCause: "Service unavailable - backend service is down or overloaded",
			Priority:  10,
		},
		{
			Name:      "DNSError",
			Pattern:   regexp.MustCompile(`(?i)(no such host|name resolution failed|dns error|unknown host)`),
			RootCause: "DNS resolution failed - service name cannot be resolved",
			Priority:  10,
		},
		{
			Name:      "KafkaBrokerError",
			Pattern:   regexp.MustCompile(`(?i)(broker not available|leader not available|connection to node)`),
			RootCause: "Kafka service is down or unreachable",
			Priority:  15,
		},
		{
			Name:      "KafkaConnectionError",
			Pattern:   regexp.MustCompile(`(?i)(kafka.*connection|kafka.*timeout|kafka.*error)`),
			RootCause: "Kafka connection error - broker may be down",
			Priority:  12,
		},
		{
			Name:      "DatabaseConnectionError",
			Pattern:   regexp.MustCompile(`(?i)(connection pool exhausted|too many connections|database.*connection.*failed)`),
			RootCause: "Database connection failed - connection pool may be exhausted",
			Priority:  10,
		},
		{
			Name:      "HTTP502",
			Pattern:   regexp.MustCompile(`(?i)(502 bad gateway|bad gateway)`),
			RootCause: "502 Bad Gateway - upstream service is unavailable",
			Priority:  10,
		},
		{
			Name:      "HTTP503",
			Pattern:   regexp.MustCompile(`(?i)(503 service unavailable)`),
			RootCause: "503 Service Unavailable - service is temporarily unavailable",
			Priority:  10,
		},
	}
	return patterns
}

// analyzeLogs performs log analysis using the configured method(s)
func analyzeLogs(ctx context.Context, client client.Client, k8sClient kubernetes.Interface, pod *corev1.Pod, config *infrav1alpha1.LogAnalysisConfig) (*infrav1alpha1.LogAnalysisResult, error) {
	if config == nil || !config.Enabled {
		return nil, nil
	}

	// Determine methods to use (backward compatibility)
	methods := config.Methods
	if len(methods) == 0 && config.Method != "" {
		// Support deprecated single Method field
		methods = []string{config.Method}
	}
	if len(methods) == 0 {
		// Default to pattern method
		methods = []string{"pattern"}
	}

	// Get log lines once (shared by all methods)
	logLines, err := getPodLogs(ctx, k8sClient, pod, config)
	if err != nil {
		return nil, fmt.Errorf("failed to get pod logs: %w", err)
	}

	if len(logLines) == 0 {
		return nil, nil
	}

	logger := log.Log.WithName("log-analysis")
	logger.Info("starting multi-method log analysis", "pod", pod.Name, "namespace", pod.Namespace, "methods", methods, "logLines", len(logLines))

	var patternResult *infrav1alpha1.PatternAnalysisResult
	var aiResult *infrav1alpha1.AIAnalysisResult
	var errorLines []string

	// Run each method in order
	for i, method := range methods {
		logger.Info("running analysis method", "method", method, "order", i+1, "total", len(methods))

		switch method {
		case "pattern":
			result, err := analyzeWithPatterns(logLines, config)
			if err != nil {
				logger.Error(err, "pattern analysis failed")
				// Store error in result for UI display
				patternResult = &infrav1alpha1.PatternAnalysisResult{
					Error: fmt.Sprintf("Pattern analysis failed: %v", err),
				}
			} else if result != nil {
				patternResult = &infrav1alpha1.PatternAnalysisResult{
					MatchedPattern: result.MatchedPattern,
					Priority:       result.Priority,
					RootCause:      result.RootCause,
					Confidence:     result.Confidence,
				}
				// Collect error lines
				errorLines = append(errorLines, result.ErrorLines...)
				logger.Info("pattern analysis completed", "matchedPattern", result.MatchedPattern, "confidence", result.Confidence)
			}

		case "ai":
			result, err := analyzeWithAI(ctx, client, logLines, pod, config)
			if err != nil {
				logger.Error(err, "AI analysis failed")
				// Store error in result for UI display
				aiResult = &infrav1alpha1.AIAnalysisResult{
					Error: fmt.Sprintf("AI analysis failed: %v", err),
				}
			} else if result != nil {
				aiResult = &infrav1alpha1.AIAnalysisResult{
					Model:      result.Model,
					RootCause:  result.RootCause,
					Confidence: result.Confidence,
				}
				// Collect error lines
				errorLines = append(errorLines, result.ErrorLines...)
				logger.Info("AI analysis completed", "model", result.Model, "confidence", result.Confidence)
			}

		default:
			logger.Info("unknown analysis method, skipping", "method", method)
		}
	}

	// Merge results from all methods
	finalResult := mergeAnalysisResults(patternResult, aiResult, methods, errorLines)
	if finalResult != nil {
		finalResult.AnalyzedAt = metav1.Now()
		logger.Info("multi-method analysis completed", "methods", finalResult.Methods, "rootCause", finalResult.RootCause, "confidence", finalResult.Confidence)
	}

	return finalResult, nil
}

// mergeAnalysisResults combines results from multiple analysis methods
func mergeAnalysisResults(patternResult *infrav1alpha1.PatternAnalysisResult, aiResult *infrav1alpha1.AIAnalysisResult, methods []string, errorLines []string) *infrav1alpha1.LogAnalysisResult {
	result := &infrav1alpha1.LogAnalysisResult{
		Methods:       methods,
		PatternResult: patternResult,
		AIResult:      aiResult,
		ErrorLines:    deduplicateLines(errorLines),
	}

	// Determine primary root cause and confidence based on available results
	if aiResult != nil && patternResult != nil {
		// Both methods ran
		if aiResult.Confidence > 80 {
			// High AI confidence - use AI as primary
			result.RootCause = aiResult.RootCause
			result.Confidence = aiResult.Confidence
			result.Method = "ai" // For backward compatibility
		} else if aiResult.Confidence < 50 {
			// Low AI confidence - use pattern as primary
			result.RootCause = patternResult.RootCause
			result.Confidence = patternResult.Confidence
			result.Method = "pattern" // For backward compatibility
		} else {
			// Medium AI confidence - combine both
			result.RootCause = fmt.Sprintf("[Pattern] %s | [AI] %s", patternResult.RootCause, aiResult.RootCause)
			result.Confidence = (patternResult.Confidence + aiResult.Confidence) / 2
			result.Method = "pattern+ai" // For backward compatibility
		}
	} else if aiResult != nil {
		// Only AI ran
		result.RootCause = aiResult.RootCause
		result.Confidence = aiResult.Confidence
		result.Method = "ai" // For backward compatibility
	} else if patternResult != nil {
		// Only pattern ran
		result.RootCause = patternResult.RootCause
		result.Confidence = patternResult.Confidence
		result.Method = "pattern" // For backward compatibility
	} else {
		// No results
		return nil
	}

	return result
}

// deduplicateLines removes duplicate lines from a slice
func deduplicateLines(lines []string) []string {
	seen := make(map[string]bool)
	result := []string{}
	for _, line := range lines {
		if !seen[line] {
			seen[line] = true
			result = append(result, line)
		}
	}
	return result
}

// getPodLogs retrieves logs from a pod container
func getPodLogs(ctx context.Context, k8sClient kubernetes.Interface, pod *corev1.Pod, config *infrav1alpha1.LogAnalysisConfig) ([]string, error) {
	// Determine which container to analyze
	// Priority: 1) First non-ready container, 2) Container with errors (waiting/terminated), 3) First container
	containerName := ""
	var containerWithError string

	for _, containerStatus := range pod.Status.ContainerStatuses {
		// Check if container is not ready
		if !containerStatus.Ready {
			// Prefer containers with actual errors (waiting or terminated states)
			if containerStatus.State.Waiting != nil || containerStatus.State.Terminated != nil {
				if containerWithError == "" {
					containerWithError = containerStatus.Name
				}
			}
			// Use first non-ready container if we haven't found one yet
			if containerName == "" {
				containerName = containerStatus.Name
			}
		}
	}

	// Prefer container with errors over just non-ready
	if containerWithError != "" {
		containerName = containerWithError
	}

	// Fallback to first container if all are ready (shouldn't happen for non-ready pods, but just in case)
	if containerName == "" && len(pod.Spec.Containers) > 0 {
		containerName = pod.Spec.Containers[0].Name
	}

	if containerName == "" {
		return nil, fmt.Errorf("no container found to analyze for pod %s/%s", pod.Namespace, pod.Name)
	}

	logger := log.Log.WithName("log-analysis")
	logger.Info("analyzing logs", "pod", pod.Name, "namespace", pod.Namespace, "container", containerName)
	if containerWithError != "" {
		logger.V(1).Info("selected container with error state", "container", containerName)
	} else if containerName != "" {
		logger.V(1).Info("selected non-ready container", "container", containerName)
	}

	// Get lines to analyze (default 100)
	linesToAnalyze := int64(100)
	if config.LinesToAnalyze != nil {
		linesToAnalyze = int64(*config.LinesToAnalyze)
	}

	// Get logs from Kubernetes API
	req := k8sClient.CoreV1().Pods(pod.Namespace).GetLogs(pod.Name, &corev1.PodLogOptions{
		Container: containerName,
		TailLines: &linesToAnalyze,
	})

	logStream, err := req.Stream(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to stream logs: %w", err)
	}
	defer logStream.Close()

	// Read all lines
	var allLines []string
	scanner := bufio.NewScanner(logStream)
	for scanner.Scan() {
		allLines = append(allLines, scanner.Text())
	}
	if err := scanner.Err(); err != nil && err != io.EOF {
		return nil, fmt.Errorf("failed to read log stream: %w", err)
	}

	logger.Info("retrieved log lines", "totalLines", len(allLines))

	// Filter for errors if configured (default true)
	filterErrorsOnly := true
	if config.FilterErrorsOnly != nil {
		filterErrorsOnly = *config.FilterErrorsOnly
	}

	if filterErrorsOnly {
		filteredLines := filterErrorLines(allLines)
		logger.Info("filtered error lines", "originalLines", len(allLines), "errorLines", len(filteredLines))
		return filteredLines, nil
	}

	return allLines, nil
}

// filterErrorLines filters log lines for errors and warnings
func filterErrorLines(lines []string) []string {
	var errorLines []string
	errorKeywords := []string{
		"error", "err", "failed", "failure", "fatal", "panic",
		"exception", "warning", "warn", "critical", "alert",
	}

	for _, line := range lines {
		lowerLine := strings.ToLower(line)
		for _, keyword := range errorKeywords {
			if strings.Contains(lowerLine, keyword) {
				errorLines = append(errorLines, line)
				break
			}
		}
	}

	return errorLines
}

// analyzeWithPatterns analyzes logs using pattern matching
func analyzeWithPatterns(logLines []string, config *infrav1alpha1.LogAnalysisConfig) (*infrav1alpha1.LogAnalysisResult, error) {
	var patterns []PatternMatch

	// Use custom patterns if provided, otherwise use defaults
	logger := log.Log.WithName("log-analysis")
	if len(config.Patterns) > 0 {
		logger.Info("using custom patterns", "count", len(config.Patterns))
		for _, customPattern := range config.Patterns {
			regex, err := regexp.Compile(customPattern.Pattern)
			if err != nil {
				logger.Info("failed to compile pattern", "name", customPattern.Name, "pattern", customPattern.Pattern, "error", err)
				continue // Skip invalid patterns
			}
			patterns = append(patterns, PatternMatch{
				Name:      customPattern.Name,
				Pattern:   regex,
				RootCause: customPattern.RootCause,
				Priority:  customPattern.Priority,
			})
			logger.Info("pattern compiled successfully", "name", customPattern.Name, "priority", customPattern.Priority, "rootCause", customPattern.RootCause)
		}
		if len(patterns) == 0 {
			logger.Info("no valid custom patterns found, falling back to defaults")
			// Fall back to default patterns if all custom patterns failed
			defaultPatterns := getDefaultPatterns()
			for _, dp := range defaultPatterns {
				patterns = append(patterns, PatternMatch{
					Name:      dp.Name,
					Pattern:   dp.Pattern,
					RootCause: dp.RootCause,
					Priority:  dp.Priority,
				})
			}
		}
	} else {
		// Use default patterns
		logger.V(2).Info("using default patterns")
		defaultPatterns := getDefaultPatterns()
		for _, dp := range defaultPatterns {
			patterns = append(patterns, PatternMatch{
				Name:      dp.Name,
				Pattern:   dp.Pattern,
				RootCause: dp.RootCause,
				Priority:  dp.Priority,
			})
		}
	}

	// Sort patterns by priority (highest first)
	sort.Slice(patterns, func(i, j int) bool {
		return patterns[i].Priority > patterns[j].Priority
	})

	// Match patterns against log lines
	var matchedLines []string
	var bestMatch *PatternMatch

	logger.V(2).Info("matching patterns", "logLines", len(logLines), "patterns", len(patterns))
	if len(logLines) > 0 {
		logger.V(2).Info("sample log lines", "firstLine", logLines[0], "lastLine", logLines[len(logLines)-1])
	}
	for _, line := range logLines {
		for i := range patterns {
			if patterns[i].Pattern.MatchString(line) {
				matchedLines = append(matchedLines, line)
				// Only log the first match to avoid spam (we already have the matched lines in the result)
				if bestMatch == nil {
					logger.Info("pattern matched", "pattern", patterns[i].Name, "line", line, "rootCause", patterns[i].RootCause)
					bestMatch = &patterns[i]
				}
				break // Only match once per line
			}
		}
	}

	// Log summary of matches instead of individual matches
	if len(matchedLines) > 0 && bestMatch != nil {
		logger.Info("pattern matching summary", "pattern", bestMatch.Name, "matchedLines", len(matchedLines), "rootCause", bestMatch.RootCause)
	} else if len(logLines) > 0 {
		logger.Info("no patterns matched", "logLines", len(logLines), "patterns", len(patterns), "sampleLine", logLines[0])
	}

	if bestMatch == nil {
		// No pattern matched, return generic result
		if len(logLines) > 0 {
			return &infrav1alpha1.LogAnalysisResult{
				RootCause:      "Unknown error detected in logs",
				Confidence:     30,
				ErrorLines:     logLines[:min(10, len(logLines))], // Return first 10 lines
				MatchedPattern: "",
				Priority:       0,
			}, nil
		}
		return nil, nil
	}

	// Determine root cause message
	rootCause := bestMatch.RootCause
	if rootCause == "" && len(matchedLines) > 0 {
		rootCause = matchedLines[0] // Use first matched line as root cause
	}

	// Calculate confidence based on number of matches
	confidence := int32(50)
	if len(matchedLines) >= 3 {
		confidence = 80
	} else if len(matchedLines) >= 2 {
		confidence = 65
	}

	return &infrav1alpha1.LogAnalysisResult{
		RootCause:      rootCause,
		Confidence:     confidence,
		ErrorLines:     matchedLines,
		MatchedPattern: bestMatch.Name,
		Priority:       bestMatch.Priority,
	}, nil
}

// PatternMatch represents a pattern with its match information
type PatternMatch struct {
	Name      string
	Pattern   *regexp.Regexp
	RootCause string
	Priority  int32
}

// getAPIKeyFromSecret retrieves the API key from a Kubernetes Secret
func getAPIKeyFromSecret(ctx context.Context, k8sClient client.Client, secretRef *corev1.SecretKeySelector, namespace string) (string, error) {
	if secretRef == nil {
		return "", fmt.Errorf("secret reference is nil")
	}

	// SecretKeySelector doesn't have a Namespace field, so we use the provided namespace
	// Secrets are typically in the same namespace as the pod
	secretNamespace := namespace

	var secret corev1.Secret
	secretKey := types.NamespacedName{
		Namespace: secretNamespace,
		Name:      secretRef.Name,
	}

	if err := k8sClient.Get(ctx, secretKey, &secret); err != nil {
		return "", fmt.Errorf("failed to get secret %s/%s: %w", secretNamespace, secretRef.Name, err)
	}

	keyName := secretRef.Key
	if keyName == "" {
		keyName = "api-key" // Default key name
	}

	apiKeyBytes, ok := secret.Data[keyName]
	if !ok {
		return "", fmt.Errorf("key %s not found in secret %s/%s", keyName, secretNamespace, secretRef.Name)
	}

	return string(apiKeyBytes), nil
}

// analyzeWithAI analyzes logs using AI endpoint
func analyzeWithAI(ctx context.Context, k8sClient client.Client, logLines []string, pod *corev1.Pod, config *infrav1alpha1.LogAnalysisConfig) (*infrav1alpha1.LogAnalysisResult, error) {
	if config.AIEndpoint == "" {
		return nil, fmt.Errorf("aiEndpoint is required for AI analysis")
	}

	// Get API key if configured
	var apiKey string
	var err error
	if config.AIAPIKey != nil {
		apiKey, err = getAPIKeyFromSecret(ctx, k8sClient, config.AIAPIKey, pod.Namespace)
		if err != nil {
			return nil, fmt.Errorf("failed to get API key: %w", err)
		}
	}

	// Determine request format based on endpoint and format setting
	requestBody, err := buildAIRequest(config, logLines, pod)
	if err != nil {
		return nil, fmt.Errorf("failed to build AI request: %w", err)
	}

	// Create HTTP request
	req, err := http.NewRequestWithContext(ctx, "POST", config.AIEndpoint, bytes.NewBuffer(requestBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")

	// Add authentication header if API key is provided
	if apiKey != "" {
		authHeader := config.AIAuthHeader
		if authHeader == "" {
			authHeader = "Authorization"
		}

		authPrefix := config.AIAuthPrefix
		if authPrefix == "" {
			authPrefix = "Bearer"
		}

		authValue := apiKey
		if authPrefix != "" {
			authValue = authPrefix + " " + apiKey
		}

		req.Header.Set(authHeader, authValue)
	}

	// Make HTTP request with timeout
	httpClient := &http.Client{
		Timeout: 30 * time.Second,
	}

	resp, err := httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make AI request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("AI endpoint returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	// Parse response
	result, err := parseAIResponse(resp.Body, config.AIEndpoint, config.AIFormat)
	if err != nil {
		return nil, fmt.Errorf("failed to parse AI response: %w", err)
	}

	// Add error lines to result
	result.ErrorLines = logLines[:min(20, len(logLines))]

	return result, nil
}

// buildAIRequest builds the request body based on endpoint type and format setting
func buildAIRequest(config *infrav1alpha1.LogAnalysisConfig, logLines []string, pod *corev1.Pod) ([]byte, error) {
	logsText := strings.Join(logLines, "\n")
	prompt := fmt.Sprintf(`Analyze these Kubernetes pod logs and identify the root cause why the pod is not ready.

Pod: %s/%s
Phase: %s

Logs:
%s

Provide a concise root cause analysis. Focus on the primary issue.`, pod.Namespace, pod.Name, pod.Status.Phase, logsText)

	var requestBody map[string]interface{}

	// Determine format: use explicit format if set, otherwise auto-detect from endpoint
	apiFormat := config.AIFormat
	if apiFormat == "" {
		// Auto-detect based on endpoint URL
		if strings.Contains(config.AIEndpoint, "openai.com") {
			apiFormat = "openai"
		} else if strings.Contains(config.AIEndpoint, "anthropic.com") {
			apiFormat = "anthropic"
		} else if strings.Contains(config.AIEndpoint, "ollama") || strings.Contains(config.AIEndpoint, ":11434") {
			apiFormat = "ollama"
		} else {
			// Default to OpenAI format for unknown endpoints (most compatible)
			apiFormat = "openai"
		}
	}

	// Determine model: use explicit model if set, otherwise use defaults
	model := config.AIModel
	if model == "" {
		// Use defaults based on format
		switch apiFormat {
		case "openai":
			model = "gpt-3.5-turbo"
		case "anthropic":
			model = "claude-3-haiku-20240307"
		case "ollama":
			model = "llama2"
		default:
			model = "" // Generic format doesn't require model
		}
	}

	// Build request based on format
	switch apiFormat {
	case "openai":
		// OpenAI format (also works for OpenAI-compatible services like Together AI, Groq, LocalAI, vLLM, etc.)
		requestBody = map[string]interface{}{
			"model": model,
			"messages": []map[string]string{
				{
					"role":    "system",
					"content": "You are a Kubernetes troubleshooting expert. Analyze pod logs and identify root causes.",
				},
				{
					"role":    "user",
					"content": prompt,
				},
			},
			"max_tokens":  200,
			"temperature": 0.3,
		}
	case "anthropic":
		// Anthropic format
		requestBody = map[string]interface{}{
			"model":      model,
			"max_tokens": 200,
			"messages": []map[string]string{
				{
					"role":    "user",
					"content": prompt,
				},
			},
		}
	case "ollama":
		// Ollama format
		requestBody = map[string]interface{}{
			"model":  model,
			"prompt": prompt,
			"stream": false,
		}
	default:
		// Generic format
		requestBody = map[string]interface{}{
			"prompt":     prompt,
			"max_tokens": 200,
		}
		if model != "" {
			requestBody["model"] = model
		}
	}

	return json.Marshal(requestBody)
}

// parseAIResponse parses the AI response based on endpoint type and format setting
func parseAIResponse(body io.Reader, endpoint string, format string) (*infrav1alpha1.LogAnalysisResult, error) {
	bodyBytes, err := io.ReadAll(body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	var response map[string]interface{}
	if err := json.Unmarshal(bodyBytes, &response); err != nil {
		return nil, fmt.Errorf("failed to parse JSON response: %w", err)
	}

	var rootCause string
	var confidence int32

	// Determine format: use explicit format if set, otherwise auto-detect from endpoint
	apiFormat := format
	if apiFormat == "" {
		// Auto-detect based on endpoint URL
		if strings.Contains(endpoint, "openai.com") {
			apiFormat = "openai"
		} else if strings.Contains(endpoint, "anthropic.com") {
			apiFormat = "anthropic"
		} else if strings.Contains(endpoint, "ollama") || strings.Contains(endpoint, ":11434") {
			apiFormat = "ollama"
		} else {
			// Default to OpenAI format for unknown endpoints (most compatible)
			apiFormat = "openai"
		}
	}

	// Parse based on format
	switch apiFormat {
	case "openai":
		// OpenAI format: {"choices": [{"message": {"content": "..."}}]}
		// Works for OpenAI and OpenAI-compatible services (Together AI, Groq, LocalAI, vLLM, etc.)
		if choices, ok := response["choices"].([]interface{}); ok && len(choices) > 0 {
			if choice, ok := choices[0].(map[string]interface{}); ok {
				if message, ok := choice["message"].(map[string]interface{}); ok {
					if content, ok := message["content"].(string); ok {
						rootCause = strings.TrimSpace(content)
					}
				}
			}
		}
	case "anthropic":
		// Anthropic format: {"content": [{"text": "..."}]}
		if content, ok := response["content"].([]interface{}); ok && len(content) > 0 {
			if block, ok := content[0].(map[string]interface{}); ok {
				if text, ok := block["text"].(string); ok {
					rootCause = strings.TrimSpace(text)
				}
			}
		}
	case "ollama":
		// Ollama format: {"response": "..."}
		if responseText, ok := response["response"].(string); ok {
			rootCause = strings.TrimSpace(responseText)
		}
	default:
		// Generic format: try common fields
		if text, ok := response["text"].(string); ok {
			rootCause = strings.TrimSpace(text)
		} else if answer, ok := response["answer"].(string); ok {
			rootCause = strings.TrimSpace(answer)
		} else if result, ok := response["result"].(string); ok {
			rootCause = strings.TrimSpace(result)
		} else if content, ok := response["content"].(string); ok {
			rootCause = strings.TrimSpace(content)
		}
	}

	if rootCause == "" {
		// Fallback: return raw response as string
		rootCause = fmt.Sprintf("AI analysis completed (response format not recognized): %s", string(bodyBytes))
		confidence = 50
	} else {
		// Calculate dynamic confidence based on response quality
		confidence = calculateAIConfidence(rootCause)
	}

	// Try to extract model from response
	model := ""
	if modelField, ok := response["model"].(string); ok {
		model = modelField
	}

	return &infrav1alpha1.LogAnalysisResult{
		RootCause:  rootCause,
		Confidence: confidence,
		Model:      model,
	}, nil
}

// calculateAIConfidence calculates confidence score based on AI response quality
func calculateAIConfidence(rootCause string) int32 {
	confidence := int32(60) // Base confidence

	// Factor 1: Response length (detailed responses are more confident)
	length := len(rootCause)
	if length > 200 {
		confidence += 20 // Very detailed
	} else if length > 100 {
		confidence += 15 // Detailed
	} else if length > 50 {
		confidence += 10 // Moderate detail
	} else if length < 20 {
		confidence -= 20 // Too short, likely incomplete
	}

	// Factor 2: Check for uncertainty indicators (reduce confidence)
	uncertainWords := []string{
		"might", "maybe", "possibly", "perhaps", "unclear",
		"not sure", "could be", "may be", "unsure", "uncertain",
		"difficult to determine", "hard to say",
	}
	lowerRootCause := strings.ToLower(rootCause)
	for _, word := range uncertainWords {
		if strings.Contains(lowerRootCause, word) {
			confidence -= 15
			break // Only penalize once for uncertainty
		}
	}

	// Factor 3: Check for certainty indicators (increase confidence)
	certaintyWords := []string{
		"error:", "failed:", "exception:", "timeout:", "connection refused",
		"out of memory", "disk full", "permission denied", "not found",
		"crashed", "terminated", "killed", "panic:", "fatal:",
	}
	certaintyCount := 0
	for _, word := range certaintyWords {
		if strings.Contains(lowerRootCause, word) {
			certaintyCount++
		}
	}
	if certaintyCount > 0 {
		confidence += int32(certaintyCount * 5) // +5 per certainty indicator, max +15
		if confidence > 100 {
			confidence = 100
		}
	}

	// Factor 4: Structured format (bullet points, line numbers, clear structure)
	structureIndicators := []string{"\n-", "\n*", "\n1.", "\n2.", "line ", "at line", "error at"}
	for _, indicator := range structureIndicators {
		if strings.Contains(lowerRootCause, indicator) {
			confidence += 8
			break // Only add once for structure
		}
	}

	// Factor 5: Check if response contains actual analysis vs generic statements
	genericPhrases := []string{
		"i cannot", "i can't", "i don't have", "no information",
		"please provide", "need more", "insufficient",
	}
	for _, phrase := range genericPhrases {
		if strings.Contains(lowerRootCause, phrase) {
			confidence -= 25 // Significant reduction for non-answers
			break
		}
	}

	// Ensure confidence stays within valid range (0-100)
	if confidence > 100 {
		confidence = 100
	} else if confidence < 10 {
		confidence = 10 // Minimum confidence (always give some credit)
	}

	return confidence
}

// min returns the minimum of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
