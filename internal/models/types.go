package models

import (
	"net"
	"time"
)

// ConfidenceLevel represents the threat confidence level
type ConfidenceLevel string

const (
	ConfidenceNone     ConfidenceLevel = "NONE"
	ConfidenceLow      ConfidenceLevel = "LOW"
	ConfidenceMedium   ConfidenceLevel = "MEDIUM"
	ConfidenceHigh     ConfidenceLevel = "HIGH"
	ConfidenceCritical ConfidenceLevel = "CRITICAL"
)

// IPRange represents an IP address range with associated metadata
type IPRange struct {
	StartIP     net.IP
	EndIP       net.IP
	StartIPInt  uint32
	EndIPInt    uint32
	ASN         uint32
	CountryCode string
	Description string
}

// MaliciousIP represents a malicious IP with its score
type MaliciousIP struct {
	IP          string
	Score       float64
	ASN         uint32
	CountryCode string
	FirstSeen   time.Time
	LastSeen    time.Time
}

// ClassificationRequest represents a request to classify an IP/ASN/Country
type ClassificationRequest struct {
	IP      string `json:"ip,omitempty"`
	ASN     uint32 `json:"asn,omitempty"`
	Country string `json:"country,omitempty"`
}

// ClassificationResponse represents the classification result
type ClassificationResponse struct {
	Input           string          `json:"input"`
	ConfidenceLevel ConfidenceLevel `json:"confidence_level"`
	Score           float64         `json:"score"`
	ASN             uint32          `json:"asn,omitempty"`
	CountryCode     string          `json:"country_code,omitempty"`
	Description     string          `json:"description,omitempty"`
	Reason          string          `json:"reason,omitempty"`
	MaliciousCount  int             `json:"malicious_count,omitempty"`
	TotalIPs        int             `json:"total_ips,omitempty"`
}

// ASNStats holds statistics for an ASN
type ASNStats struct {
	ASN            uint32
	MaliciousCount int
	TotalSubnets   int
	AvgScore       float64
	MaxScore       float64
}

// CountryStats holds statistics for a country
type CountryStats struct {
	CountryCode    string
	MaliciousCount int
	TotalSubnets   int
	AvgScore       float64
	MaxScore       float64
}

// SubnetStats holds statistics for a subnet
type SubnetStats struct {
	StartIP        net.IP
	EndIP          net.IP
	MaliciousCount int
	TotalIPs       int
	AvgScore       float64
	MaxScore       float64
	MaliciousIPs   []string
}
