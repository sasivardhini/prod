package api

import (
	"net/http"

	"github.com/gorilla/mux"
	log "github.com/sirupsen/logrus"
)

// SetupRouter creates and configures the API router
func SetupRouter(handler *Handler) *mux.Router {
	router := mux.NewRouter()

	// Middleware
	router.Use(loggingMiddleware)
	router.Use(corsMiddleware)

	// API v1 routes
	api := router.PathPrefix("/api/v1").Subrouter()

	// Health check
	api.HandleFunc("/health", handler.HealthCheck).Methods("GET")

	// Classification endpoints
	api.HandleFunc("/classify/ip/{ip}", handler.ClassifyIP).Methods("GET")
	api.HandleFunc("/classify/asn/{asn}", handler.ClassifyASN).Methods("GET")
	api.HandleFunc("/classify/country/{country}", handler.ClassifyCountry).Methods("GET")
	api.HandleFunc("/classify/batch", handler.ClassifyBatch).Methods("POST")

	// Root health check
	router.HandleFunc("/health", handler.HealthCheck).Methods("GET")
	router.HandleFunc("/", handler.HealthCheck).Methods("GET")

	return router
}

// loggingMiddleware logs all incoming requests
func loggingMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		log.Infof("%s %s %s", r.RemoteAddr, r.Method, r.URL.Path)
		next.ServeHTTP(w, r)
	})
}

// corsMiddleware adds CORS headers
func corsMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if r.Method == "OPTIONS" {
			w.WriteHeader(http.StatusOK)
			return
		}

		next.ServeHTTP(w, r)
	})
}
