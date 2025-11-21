package data

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/onefirewall/classifier/internal/models"
	log "github.com/sirupsen/logrus"
)

// OneFirewallClient manages interactions with OneFirewall API
type OneFirewallClient struct {
	baseURL    string
	httpClient *http.Client
	cache      map[string]*models.MaliciousIP
	mu         sync.RWMutex
}

// OneFirewallResponse represents the API response structure
type OneFirewallResponse struct {
	IP          string  `json:"ip"`
	Score       float64 `json:"score"`
	ASN         uint32  `json:"asn"`
	Country     string  `json:"country"`
	FirstSeen   string  `json:"first_seen"`
	LastSeen    string  `json:"last_seen"`
	Description string  `json:"description"`
}

// NewOneFirewallClient creates a new OneFirewall API client
func NewOneFirewallClient(baseURL string) *OneFirewallClient {
	if baseURL == "" {
		baseURL = "https://app.onefirewall.com"
	}

	return &OneFirewallClient{
		baseURL: baseURL,
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

	// Make API request
	url := fmt.Sprintf("%s/api/search?ioc=%s", c.baseURL, ip)
	resp, err := c.httpClient.Get(url)
	if err != nil {
		log.Warnf("Failed to fetch IP info for %s: %v", ip, err)
		return nil, err
	}
	defer resp.Body.Close()

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
		return nil, err
	}

	firstSeen, _ := time.Parse(time.RFC3339, apiResp.FirstSeen)
	lastSeen, _ := time.Parse(time.RFC3339, apiResp.LastSeen)

	maliciousIP := &models.MaliciousIP{
		IP:          apiResp.IP,
		Score:       apiResp.Score,
		ASN:         apiResp.ASN,
		CountryCode: apiResp.Country,
		FirstSeen:   firstSeen,
		LastSeen:    lastSeen,
	}

	// Cache the result
	c.mu.Lock()
	c.cache[ip] = maliciousIP
	c.mu.Unlock()

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
func (c *OneFirewallClient) IsMalicious(ip string) (bool, float64) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if malIP, ok := c.cache[ip]; ok {
		return true, malIP.Score
	}

	return false, 0
}
