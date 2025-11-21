package api

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/gorilla/mux"
	"github.com/onefirewall/classifier/internal/classifier"
	"github.com/onefirewall/classifier/internal/models"
	log "github.com/sirupsen/logrus"
)

// Handler manages HTTP request handlers
type Handler struct {
	classifier *classifier.Classifier
}

// NewHandler creates a new API handler
func NewHandler(c *classifier.Classifier) *Handler {
	return &Handler{
		classifier: c,
	}
}

// HealthCheck handles health check requests
func (h *Handler) HealthCheck(w http.ResponseWriter, r *http.Request) {
	response := map[string]string{
		"status": "ok",
		"service": "onefirewall-classifier",
	}
	respondJSON(w, http.StatusOK, response)
}

// ClassifyIP handles IP classification requests
// GET /api/v1/classify/ip/{ip}
func (h *Handler) ClassifyIP(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	ip := vars["ip"]

	if ip == "" {
		respondError(w, http.StatusBadRequest, "IP address is required")
		return
	}

	log.Infof("Classifying IP: %s", ip)

	result, err := h.classifier.ClassifyIP(ip)
	if err != nil {
		log.Errorf("Failed to classify IP %s: %v", ip, err)
		respondError(w, http.StatusInternalServerError, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, result)
}

// ClassifyASN handles ASN classification requests
// GET /api/v1/classify/asn/{asn}
func (h *Handler) ClassifyASN(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	asnStr := vars["asn"]

	if asnStr == "" {
		respondError(w, http.StatusBadRequest, "ASN is required")
		return
	}

	// Parse ASN (handle both "12345" and "AS12345" formats)
	var asn uint64
	var err error

	if len(asnStr) > 2 && (asnStr[:2] == "AS" || asnStr[:2] == "as") {
		asn, err = strconv.ParseUint(asnStr[2:], 10, 32)
	} else {
		asn, err = strconv.ParseUint(asnStr, 10, 32)
	}

	if err != nil {
		respondError(w, http.StatusBadRequest, "Invalid ASN format")
		return
	}

	log.Infof("Classifying ASN: %d", asn)

	result, err := h.classifier.ClassifyASN(uint32(asn))
	if err != nil {
		log.Errorf("Failed to classify ASN %d: %v", asn, err)
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, result)
}

// ClassifyCountry handles country classification requests
// GET /api/v1/classify/country/{country}
func (h *Handler) ClassifyCountry(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	country := vars["country"]

	if country == "" {
		respondError(w, http.StatusBadRequest, "Country code is required")
		return
	}

	log.Infof("Classifying Country: %s", country)

	result, err := h.classifier.ClassifyCountry(country)
	if err != nil {
		log.Errorf("Failed to classify country %s: %v", country, err)
		respondError(w, http.StatusNotFound, err.Error())
		return
	}

	respondJSON(w, http.StatusOK, result)
}

// ClassifyBatch handles batch classification requests
// POST /api/v1/classify/batch
func (h *Handler) ClassifyBatch(w http.ResponseWriter, r *http.Request) {
	var requests []models.ClassificationRequest

	if err := json.NewDecoder(r.Body).Decode(&requests); err != nil {
		respondError(w, http.StatusBadRequest, "Invalid request body")
		return
	}

	responses := make([]*models.ClassificationResponse, 0, len(requests))

	for _, req := range requests {
		var result *models.ClassificationResponse
		var err error

		if req.IP != "" {
			result, err = h.classifier.ClassifyIP(req.IP)
		} else if req.ASN != 0 {
			result, err = h.classifier.ClassifyASN(req.ASN)
		} else if req.Country != "" {
			result, err = h.classifier.ClassifyCountry(req.Country)
		} else {
			continue
		}

		if err != nil {
			log.Warnf("Failed to classify request: %v", err)
			continue
		}

		responses = append(responses, result)
	}

	respondJSON(w, http.StatusOK, map[string]interface{}{
		"results": responses,
		"total":   len(responses),
	})
}

// respondJSON writes a JSON response
func respondJSON(w http.ResponseWriter, status int, data interface{}) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(data); err != nil {
		log.Errorf("Failed to encode JSON response: %v", err)
	}
}

// respondError writes an error JSON response
func respondError(w http.ResponseWriter, status int, message string) {
	respondJSON(w, status, map[string]string{
		"error": message,
	})
}
