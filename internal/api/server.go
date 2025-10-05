package api

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/eshwanth/distributed-search-engine/internal/search"
	"github.com/eshwanth/distributed-search-engine/internal/telemetry"
)

// Server exposes HTTP handlers for querying the search index.
type Server struct {
	Search *search.Service
	Logger telemetry.Logger
}

// Start launches the HTTP server on the provided address.
func (s *Server) Start(addr string) error {
	mux := http.NewServeMux()
	mux.HandleFunc("/search", s.handleSearch)

	srv := &http.Server{
		Addr:              addr,
		Handler:           mux,
		ReadHeaderTimeout: 2 * time.Second,
		WriteTimeout:      5 * time.Second,
	}

	s.Logger.Info("search_api_listening", "addr", addr)
	return srv.ListenAndServe()
}

func (s *Server) handleSearch(w http.ResponseWriter, r *http.Request) {
	start := time.Now()
	query := r.URL.Query().Get("q")
	limit := 10
	results := s.Search.Search(query, limit)

	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(results); err != nil {
		s.Logger.Error("encode_response_failed", err)
		w.WriteHeader(http.StatusInternalServerError)
		telemetry.ObserveSearch("error", time.Since(start))
		return
	}

	s.Logger.Info("search", "q", query, "count", len(results), "latency_ms", time.Since(start).Milliseconds())
	telemetry.ObserveSearch("ok", time.Since(start))
}
