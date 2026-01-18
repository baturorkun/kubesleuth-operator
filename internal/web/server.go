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
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	fmt.Fprint(w, dashboardHTML)
}

// handleListPodSleuths returns all PodSleuth resources as JSON
func (s *Server) handleListPodSleuths(w http.ResponseWriter, r *http.Request) {
	var podSleuthList infrav1alpha1.PodSleuthList
	if err := s.client.List(r.Context(), &podSleuthList); err != nil {
		http.Error(w, fmt.Sprintf("Error listing PodSleuth: %v", err), http.StatusInternalServerError)
		return
	}

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

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(podSleuth)
}
