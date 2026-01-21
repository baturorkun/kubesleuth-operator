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

package web

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"sigs.k8s.io/controller-runtime/pkg/client"
	log "sigs.k8s.io/controller-runtime/pkg/log"

	infrav1alpha1 "github.com/baturorkun/kubebuilder-demo-operator/api/v1alpha1"
)

// Server handles web dashboard requests
type Server struct {
	client client.Client
	port   string
}

// NewServer creates a new web server
func NewServer(client client.Client, port string) *Server {
	return &Server{
		client: client,
		port:   port,
	}
}

// Start starts the HTTP server
func (s *Server) Start(ctx context.Context) error {
	mux := http.NewServeMux()

	// Dashboard HTML
	mux.HandleFunc("/", s.handleDashboard)

	// API endpoints
	mux.HandleFunc("/api/podsleuths", s.handleListPodSleuths)
	mux.HandleFunc("/api/podsleuths/", s.handleGetPodSleuth)
	mux.HandleFunc("/api/force-refresh", s.handleForceRefresh) // Restored for manual analysis trigger

	server := &http.Server{
		Addr:    s.port,
		Handler: mux,
	}

	logger := log.Log.WithName("web")
	logger.Info("Starting dashboard server", "port", s.port)

	go func() {
		<-ctx.Done()
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error(err, "Error shutting down dashboard server")
		}
	}()

	if err := server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("dashboard server error: %w", err)
	}

	return nil
}

// handleDashboard serves the HTML dashboard
func (s *Server) handleDashboard(w http.ResponseWriter, r *http.Request) {
	// Prevent browser caching - always serve fresh dashboard
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate, max-age=0")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "Thu, 01 Jan 1970 00:00:00 GMT")
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, dashboardHTML)
}

// handleListPodSleuths returns all PodSleuth resources as JSON
func (s *Server) handleListPodSleuths(w http.ResponseWriter, r *http.Request) {
	// Prevent browser caching for API calls
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate, max-age=0")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
	w.Header().Set("Content-Type", "application/json")

	var podSleuthList infrav1alpha1.PodSleuthList
	if err := s.client.List(r.Context(), &podSleuthList); err != nil {
		http.Error(w, fmt.Sprintf("Error listing PodSleuth: %v", err), http.StatusInternalServerError)
		return
	}

	// Prevent caching of API responses
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(podSleuthList)
}

// handleGetPodSleuth returns a specific PodSleuth resource as JSON
func (s *Server) handleGetPodSleuth(w http.ResponseWriter, r *http.Request) {
	name := r.URL.Path[len("/api/podsleuths/"):]
	if name == "" {
		http.Error(w, "PodSleuth name required", http.StatusBadRequest)
		return
	}

	var podSleuth infrav1alpha1.PodSleuth
	if err := s.client.Get(r.Context(), client.ObjectKey{Name: name}, &podSleuth); err != nil {
		http.Error(w, fmt.Sprintf("Error getting PodSleuth: %v", err), http.StatusNotFound)
		return
	}

	// Prevent caching of API responses
	w.Header().Set("Cache-Control", "no-cache, no-store, must-revalidate")
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(podSleuth)
}

// handleForceRefresh forces cache refresh by adding annotation to PodSleuths
// Accepts optional JSON body: {"podName":"...","podNamespace":"..."}
// If provided, only that pod will bypass cache; otherwise all pods are refreshed.
type forceRefreshRequest struct {
	PodName      string `json:"podName"`
	PodNamespace string `json:"podNamespace"`
}

func (s *Server) handleForceRefresh(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	var reqBody forceRefreshRequest
	_ = json.NewDecoder(r.Body).Decode(&reqBody) // best-effort; ignore errors for empty body
	targetPod := ""
	if reqBody.PodName != "" && reqBody.PodNamespace != "" {
		targetPod = fmt.Sprintf("%s/%s", strings.TrimSpace(reqBody.PodNamespace), strings.TrimSpace(reqBody.PodName))
	}

	var podSleuthList infrav1alpha1.PodSleuthList
	if err := s.client.List(r.Context(), &podSleuthList); err != nil {
		http.Error(w, fmt.Sprintf("Error listing PodSleuth: %v", err), http.StatusInternalServerError)
		return
	}

	log.Log.Info("force-refresh request received", "targetPod", targetPod)

	updatedCount := 0
	for i := range podSleuthList.Items {
		ps := &podSleuthList.Items[i]

		if ps.Annotations == nil {
			ps.Annotations = make(map[string]string)
		}

		if targetPod != "" {
			ps.Annotations["kubesleuth.io/force-refresh-pod"] = targetPod
		} else {
			ps.Annotations["kubesleuth.io/force-refresh"] = time.Now().Format(time.RFC3339)
		}

		if err := s.client.Update(r.Context(), ps); err != nil {
			log.Log.Error(err, "Failed to update PodSleuth with force-refresh annotation", "name", ps.Name)
			continue
		}
		updatedCount++

		log.Log.Info("force-refresh annotation applied", "podSleuth", ps.Name, "targetPod", targetPod)
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(map[string]interface{}{
		"success":   true,
		"message":   fmt.Sprintf("Force refresh triggered for %d PodSleuth resources", updatedCount),
		"count":     updatedCount,
		"targetPod": targetPod,
	})
}
