package data

import (
	"encoding/json"
	"fmt"
	"io"
	"math"
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/onefirewall/classifier/internal/models"
	log "github.com/sirupsen/logrus"
)

// OneFirewallClient manages interactions with OneFirewall API
type OneFirewallClient struct {
	baseURL    string
	apiKey     string
	httpClient *http.Client
	cache      map[string]*models.MaliciousIP
	mu         sync.RWMutex
}

// OneFirewallResponse represents the API response structure
// Actual API returns: {"header": {...}, "body": [...]}
type OneFirewallResponse struct {
	Header OneFirewallHeader   `json:"header"`
	Body   []OneFirewallIPData `json:"body"`
}

// OneFirewallHeader contains metadata about the response
type OneFirewallHeader struct {
	PageSize int `json:"page_size"`
	Live     bool `json:"live"`
}

// OneFirewallIPData represents a single IP's data from the API
type OneFirewallIPData struct {
	IP      string                `json:"ip"`
	Score   float64               `json:"score"`   // Raw score (can be 0-1000+)
	EntryTS int64                 `json:"entry_ts"` // First seen timestamp
	TS      int64                 `json:"ts"`       // Last seen timestamp
	Tags    []string              `json:"tags"`
	IPInfo  OneFirewallIPInfo     `json:"ip_info"`
	Info    OneFirewallEventInfo  `json:"info"`
}

// OneFirewallIPInfo contains IP geographic and network info
type OneFirewallIPInfo struct {
	ASN          string `json:"asn"`           // Format: "AS4760"
	Country      string `json:"country"`       // "Hong Kong"
	CountryCode  string `json:"country_code"`  // "HK"
	ASName       string `json:"as_name"`       // "HKT Limited"
	ASDomain     string `json:"as_domain"`     // "netvigator.com"
	Continent    string `json:"continent"`     // "Asia"
	ContinentCode string `json:"continent_code"` // "AS"
}

// OneFirewallEventInfo contains event statistics
type OneFirewallEventInfo struct {
	Members int      `json:"members"` // Number of alliance members reporting
	Events  int      `json:"events"`  // Number of events
	Sources []string `json:"sources"` // Sources reporting this IP
	Notes   []string `json:"notes"`   // Additional notes
}

// NewOneFirewallClient creates a new OneFirewall API client
func NewOneFirewallClient(baseURL, apiKey string) *OneFirewallClient {
	if baseURL == "" {
		baseURL = "https://app.onefirewall.com"
	}

	return &OneFirewallClient{
		baseURL: baseURL,
		apiKey:  apiKey,
		httpClient: &http.Client{
			Timeout: 10 * time.Second,
		},
		cache: make(map[string]*models.MaliciousIP),
	}
}

// GetIPInfo fetches information about an IP from OneFirewall
func (c *OneFirewallClient) GetIPInfo(ip string) (*models.MaliciousIP, error) {
	// Check cache first
	c.mu.RLock()
	if cached, ok := c.cache[ip]; ok {
		c.mu.RUnlock()
		return cached, nil
	}
	c.mu.RUnlock()

	// Don't make API calls if no API key is configured
	if c.apiKey == "" {
		return nil, fmt.Errorf("no API key configured")
	}

	// Make API request - correct endpoint is /api/v1/ips
	url := fmt.Sprintf("%s/api/v1/ips?ip=%s", c.baseURL, ip)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	// Add authentication header
	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.apiKey))
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		log.Warnf("Failed to fetch IP info for %s: %v", ip, err)
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		// IP not found in OneFirewall database - this is not an error
		log.Debugf("IP %s not found in OneFirewall database", ip)
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		log.Warnf("OneFirewall API returned status %d for IP %s: %s", resp.StatusCode, ip, string(body))
		return nil, fmt.Errorf("API returned status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response: %w", err)
	}

	var apiResp OneFirewallResponse
	if err := json.Unmarshal(body, &apiResp); err != nil {
		log.Warnf("Failed to parse OneFirewall response for %s: %v", ip, err)
		log.Debugf("Response body: %s", string(body))
		return nil, err
	}

	// Check if body has data
	if len(apiResp.Body) == 0 {
		log.Debugf("IP %s not found in OneFirewall database (empty body)", ip)
		return nil, nil
	}

	// Get the first (and should be only) result
	ipData := apiResp.Body[0]

	// Parse timestamps
	firstSeen := time.Unix(ipData.EntryTS, 0)
	lastSeen := time.Unix(ipData.TS, 0)

	// Normalize score from raw (0-1000+) to 0-10 scale
	// Based on observation: score of 234 should be HIGH (8-9)
	// Using logarithmic scaling: normalized_score = min(10, log10(score + 1) * 3)
	normalizedScore := normalizeOneFirewallScore(ipData.Score)

	// Parse ASN from "AS4760" format to uint32
	asn := parseASN(ipData.IPInfo.ASN)

	maliciousIP := &models.MaliciousIP{
		IP:          ipData.IP,
		Score:       normalizedScore,
		ASN:         asn,
		CountryCode: ipData.IPInfo.CountryCode,
		FirstSeen:   firstSeen,
		LastSeen:    lastSeen,
	}

	// Cache the result
	c.mu.Lock()
	c.cache[ip] = maliciousIP
	c.mu.Unlock()

	log.Infof("Fetched IP %s from OneFirewall: raw_score=%.0f, normalized_score=%.2f, ASN=%d, country=%s, members=%d, events=%d",
		ip, ipData.Score, normalizedScore, asn, ipData.IPInfo.CountryCode, ipData.Info.Members, ipData.Info.Events)

	return maliciousIP, nil
}

// LoadMaliciousIPs loads a batch of malicious IPs into the cache
func (c *OneFirewallClient) LoadMaliciousIPs(ips []models.MaliciousIP) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for i := range ips {
		c.cache[ips[i].IP] = &ips[i]
	}

	log.Infof("Loaded %d malicious IPs into cache", len(ips))
}

// GetCachedMaliciousIPs returns all cached malicious IPs
func (c *OneFirewallClient) GetCachedMaliciousIPs() []models.MaliciousIP {
	c.mu.RLock()
	defer c.mu.RUnlock()

	result := make([]models.MaliciousIP, 0, len(c.cache))
	for _, ip := range c.cache {
		result = append(result, *ip)
	}

	return result
}

// GetMaliciousIPsByASN returns all cached malicious IPs for a given ASN
func (c *OneFirewallClient) GetMaliciousIPsByASN(asn uint32) []models.MaliciousIP {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var result []models.MaliciousIP
	for _, ip := range c.cache {
		if ip.ASN == asn {
			result = append(result, *ip)
		}
	}

	return result
}

// GetMaliciousIPsByCountry returns all cached malicious IPs for a given country
func (c *OneFirewallClient) GetMaliciousIPsByCountry(country string) []models.MaliciousIP {
	c.mu.RLock()
	defer c.mu.RUnlock()

	var result []models.MaliciousIP
	for _, ip := range c.cache {
		if ip.CountryCode == country {
			result = append(result, *ip)
		}
	}

	return result
}

// IsMalicious checks if an IP is in the malicious cache
// If not in cache and API key is available, fetches from OneFirewall API
func (c *OneFirewallClient) IsMalicious(ip string) (bool, float64) {
	c.mu.RLock()
	cached, ok := c.cache[ip]
	c.mu.RUnlock()

	if ok {
		return true, cached.Score
	}

	// Try to fetch from API if configured
	if c.apiKey != "" {
		malIP, err := c.GetIPInfo(ip)
		if err == nil && malIP != nil {
			return true, malIP.Score
		}
	}

	return false, 0
}

// LoadMaliciousIPsFromFile loads malicious IPs from a JSON file
func (c *OneFirewallClient) LoadMaliciousIPsFromFile(filepath string) error {
	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	var ips []models.MaliciousIP
	if err := json.NewDecoder(file).Decode(&ips); err != nil {
		return fmt.Errorf("failed to decode JSON: %w", err)
	}

	c.LoadMaliciousIPs(ips)
	return nil
}

// RefreshCache clears the cache to force fresh API calls
func (c *OneFirewallClient) RefreshCache() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.cache = make(map[string]*models.MaliciousIP)
	log.Info("OneFirewall cache cleared")
}

// GetCacheSize returns the number of cached malicious IPs
func (c *OneFirewallClient) GetCacheSize() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.cache)
}

// normalizeOneFirewallScore normalizes OneFirewall raw scores to 0-10 scale
// OneFirewall returns raw scores that can range from 0 to 1000+
// Example: score=234 should map to ~8.5 (HIGH)
//
// Score mapping based on observed data:
// - 0-10: LOW (0-3)
// - 10-50: LOW-MEDIUM (3-5)
// - 50-150: MEDIUM (5-7)
// - 150-300: HIGH (7-9)
// - 300+: CRITICAL (9-10)
func normalizeOneFirewallScore(rawScore float64) float64 {
	if rawScore <= 0 {
		return 0
	}

	// Using logarithmic scaling with calibration
	// score = min(10, (log10(rawScore + 1) / log10(1000)) * 10 + offset)
	//
	// Examples:
	// rawScore=10 → normalized≈3.0 (LOW)
	// rawScore=50 → normalized≈5.0 (MEDIUM)
	// rawScore=150 → normalized≈7.0 (MEDIUM-HIGH)
	// rawScore=234 → normalized≈8.5 (HIGH)
	// rawScore=500 → normalized≈9.0 (CRITICAL)

	// Simple piecewise linear mapping for better control
	var normalized float64

	if rawScore < 10 {
		// 0-10 → 0-3 (LOW)
		normalized = rawScore * 0.3
	} else if rawScore < 50 {
		// 10-50 → 3-5 (LOW-MEDIUM)
		normalized = 3.0 + ((rawScore - 10) / 40.0) * 2.0
	} else if rawScore < 150 {
		// 50-150 → 5-7 (MEDIUM)
		normalized = 5.0 + ((rawScore - 50) / 100.0) * 2.0
	} else if rawScore < 300 {
		// 150-300 → 7-9 (HIGH)
		normalized = 7.0 + ((rawScore - 150) / 150.0) * 2.0
	} else {
		// 300+ → 9-10 (CRITICAL)
		// Logarithmic scaling for very high scores
		normalized = 9.0 + math.Min(1.0, math.Log10(rawScore/300.0))
	}

	// Cap at 10
	if normalized > 10 {
		normalized = 10
	}

	return normalized
}

// parseASN parses ASN from "AS4760" format to uint32
// Returns 0 if parsing fails
func parseASN(asnString string) uint32 {
	// Remove "AS" prefix if present
	asnString = strings.TrimPrefix(asnString, "AS")
	asnString = strings.TrimPrefix(asnString, "as")
	asnString = strings.TrimSpace(asnString)

	// Parse to uint32
	asn, err := strconv.ParseUint(asnString, 10, 32)
	if err != nil {
		log.Warnf("Failed to parse ASN from '%s': %v", asnString, err)
		return 0
	}

	return uint32(asn)
}
