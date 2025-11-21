package data

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
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
type OneFirewallResponse struct {
	IP          string  `json:"ip"`
	Score       float64 `json:"score"`
	CrimeScore  float64 `json:"crime_score"`
	ASN         uint32  `json:"asn"`
	Country     string  `json:"country"`
	CountryCode string  `json:"country_code"`
	FirstSeen   string  `json:"first_seen"`
	LastSeen    string  `json:"last_seen"`
	Description string  `json:"description"`
	Events      int     `json:"events"`
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

	// Make API request
	url := fmt.Sprintf("%s/api/search?ioc=%s", c.baseURL, ip)
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
		log.Warnf("OneFirewall API returned status %d for IP %s", resp.StatusCode, ip)
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

	firstSeen, _ := time.Parse(time.RFC3339, apiResp.FirstSeen)
	lastSeen, _ := time.Parse(time.RFC3339, apiResp.LastSeen)

	// Use CrimeScore if available, otherwise fall back to Score
	score := apiResp.Score
	if apiResp.CrimeScore > 0 {
		score = apiResp.CrimeScore
	}

	// Get country code
	countryCode := apiResp.CountryCode
	if countryCode == "" {
		countryCode = apiResp.Country
	}

	maliciousIP := &models.MaliciousIP{
		IP:          apiResp.IP,
		Score:       score,
		ASN:         apiResp.ASN,
		CountryCode: countryCode,
		FirstSeen:   firstSeen,
		LastSeen:    lastSeen,
	}

	// Cache the result
	c.mu.Lock()
	c.cache[ip] = maliciousIP
	c.mu.Unlock()

	log.Infof("Fetched IP %s from OneFirewall: score=%.2f, ASN=%d, country=%s", ip, score, apiResp.ASN, countryCode)

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
