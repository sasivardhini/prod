package classifier

import (
	"fmt"
	"net"
	"strings"

	"github.com/onefirewall/classifier/internal/data"
	"github.com/onefirewall/classifier/internal/models"
)

// Classifier performs IP/ASN/Country classification
type Classifier struct {
	ipDB      *data.IPDB
	ofClient  *data.OneFirewallClient
	scorer    *Scorer
}

// NewClassifier creates a new classifier instance
func NewClassifier(ipDB *data.IPDB, ofClient *data.OneFirewallClient) *Classifier {
	return &Classifier{
		ipDB:     ipDB,
		ofClient: ofClient,
		scorer:   NewScorer(),
	}
}

// ClassifyIP classifies an IP address
func (c *Classifier) ClassifyIP(ipStr string) (*models.ClassificationResponse, error) {
	ip := net.ParseIP(ipStr)
	if ip == nil {
		return nil, fmt.Errorf("invalid IP address: %s", ipStr)
	}

	// Lookup IP range
	ipRange, err := c.ipDB.LookupIP(ipStr)
	if err != nil {
		// IP not found in database - return NONE
		return &models.ClassificationResponse{
			Input:           ipStr,
			ConfidenceLevel: models.ConfidenceNone,
			Score:           0,
			Reason:          "IP not found in routing database",
		}, nil
	}

	// Check if ASN is 0 (not-routed)
	if ipRange.ASN == 0 {
		return &models.ClassificationResponse{
			Input:           ipStr,
			ConfidenceLevel: models.ConfidenceNone,
			Score:           0,
			ASN:             0,
			CountryCode:     ipRange.CountryCode,
			Description:     "Not routed (AS=0)",
			Reason:          "IP is marked as not-routed (ASN=0)",
		}, nil
	}

	// Check if IP is network or broadcast address
	if data.IsNetworkAddress(ip, ipRange) {
		return &models.ClassificationResponse{
			Input:           ipStr,
			ConfidenceLevel: models.ConfidenceNone,
			Score:           0,
			ASN:             ipRange.ASN,
			CountryCode:     ipRange.CountryCode,
			Description:     ipRange.Description,
			Reason:          "IP is network address (not routable)",
		}, nil
	}

	if data.IsBroadcastAddress(ip, ipRange) {
		return &models.ClassificationResponse{
			Input:           ipStr,
			ConfidenceLevel: models.ConfidenceNone,
			Score:           0,
			ASN:             ipRange.ASN,
			CountryCode:     ipRange.CountryCode,
			Description:     ipRange.Description,
			Reason:          "IP is broadcast address (not routable)",
		}, nil
	}

	// Check if this specific IP is known to be malicious
	if isMal, score := c.ofClient.IsMalicious(ipStr); isMal {
		confidence := c.scorer.ScoreToConfidence(score)
		return &models.ClassificationResponse{
			Input:           ipStr,
			ConfidenceLevel: confidence,
			Score:           score,
			ASN:             ipRange.ASN,
			CountryCode:     ipRange.CountryCode,
			Description:     ipRange.Description,
			Reason:          fmt.Sprintf("IP is known malicious with score %.2f", score),
			MaliciousCount:  1,
		}, nil
	}

	// Analyze subnet
	subnetStats := c.analyzeSubnet(ipRange)

	// Calculate confidence based on subnet statistics
	confidence, score := c.scorer.CalculateSubnetConfidence(subnetStats)

	return &models.ClassificationResponse{
		Input:           ipStr,
		ConfidenceLevel: confidence,
		Score:           score,
		ASN:             ipRange.ASN,
		CountryCode:     ipRange.CountryCode,
		Description:     ipRange.Description,
		Reason:          fmt.Sprintf("Based on subnet analysis: %d/%d IPs malicious, avg score: %.2f", subnetStats.MaliciousCount, subnetStats.TotalIPs, subnetStats.AvgScore),
		MaliciousCount:  subnetStats.MaliciousCount,
		TotalIPs:        subnetStats.TotalIPs,
	}, nil
}

// ClassifyASN classifies an entire ASN
func (c *Classifier) ClassifyASN(asn uint32) (*models.ClassificationResponse, error) {
	if asn == 0 {
		return &models.ClassificationResponse{
			Input:           fmt.Sprintf("AS%d", asn),
			ConfidenceLevel: models.ConfidenceNone,
			Score:           0,
			ASN:             asn,
			Reason:          "ASN 0 is not-routed",
		}, nil
	}

	// Get all ranges for this ASN
	ranges := c.ipDB.GetRangesByASN(asn)
	if len(ranges) == 0 {
		return nil, fmt.Errorf("ASN not found: AS%d", asn)
	}

	// Analyze ASN
	asnStats := c.analyzeASN(asn, ranges)

	// Calculate confidence based on ASN statistics
	confidence, score := c.scorer.CalculateASNConfidence(asnStats)

	return &models.ClassificationResponse{
		Input:           fmt.Sprintf("AS%d", asn),
		ConfidenceLevel: confidence,
		Score:           score,
		ASN:             asn,
		Reason:          fmt.Sprintf("ASN analysis: %d malicious IPs, avg score: %.2f, max score: %.2f", asnStats.MaliciousCount, asnStats.AvgScore, asnStats.MaxScore),
		MaliciousCount:  asnStats.MaliciousCount,
	}, nil
}

// ClassifyCountry classifies a country
func (c *Classifier) ClassifyCountry(countryCode string) (*models.ClassificationResponse, error) {
	countryCode = strings.ToUpper(countryCode)

	// Get all ranges for this country
	ranges := c.ipDB.GetRangesByCountry(countryCode)
	if len(ranges) == 0 {
		return nil, fmt.Errorf("country not found: %s", countryCode)
	}

	// Analyze country
	countryStats := c.analyzeCountry(countryCode, ranges)

	// Calculate confidence based on country statistics
	confidence, score := c.scorer.CalculateCountryConfidence(countryStats)

	return &models.ClassificationResponse{
		Input:           countryCode,
		ConfidenceLevel: confidence,
		Score:           score,
		CountryCode:     countryCode,
		Reason:          fmt.Sprintf("Country analysis: %d malicious IPs, avg score: %.2f, max score: %.2f", countryStats.MaliciousCount, countryStats.AvgScore, countryStats.MaxScore),
		MaliciousCount:  countryStats.MaliciousCount,
	}, nil
}

// analyzeSubnet analyzes malicious activity within a subnet
func (c *Classifier) analyzeSubnet(ipRange *models.IPRange) *models.SubnetStats {
	stats := &models.SubnetStats{
		StartIP:        ipRange.StartIP,
		EndIP:          ipRange.EndIP,
		MaliciousCount: 0,
		TotalIPs:       int(data.GetSubnetSize(ipRange)),
		AvgScore:       0,
		MaxScore:       0,
		MaliciousIPs:   make([]string, 0),
	}

	// Get all malicious IPs
	allMalicious := c.ofClient.GetCachedMaliciousIPs()

	totalScore := 0.0
	for _, malIP := range allMalicious {
		ip := net.ParseIP(malIP.IP)
		if ip == nil {
			continue
		}

		ipInt := ipToUint32(ip)
		if ipInt >= ipRange.StartIPInt && ipInt <= ipRange.EndIPInt {
			stats.MaliciousCount++
			stats.MaliciousIPs = append(stats.MaliciousIPs, malIP.IP)
			totalScore += malIP.Score

			if malIP.Score > stats.MaxScore {
				stats.MaxScore = malIP.Score
			}
		}
	}

	if stats.MaliciousCount > 0 {
		stats.AvgScore = totalScore / float64(stats.MaliciousCount)
	}

	return stats
}

// analyzeASN analyzes malicious activity within an ASN
func (c *Classifier) analyzeASN(asn uint32, ranges []models.IPRange) *models.ASNStats {
	stats := &models.ASNStats{
		ASN:            asn,
		MaliciousCount: 0,
		TotalSubnets:   len(ranges),
		AvgScore:       0,
		MaxScore:       0,
	}

	maliciousIPs := c.ofClient.GetMaliciousIPsByASN(asn)

	totalScore := 0.0
	for _, malIP := range maliciousIPs {
		stats.MaliciousCount++
		totalScore += malIP.Score

		if malIP.Score > stats.MaxScore {
			stats.MaxScore = malIP.Score
		}
	}

	if stats.MaliciousCount > 0 {
		stats.AvgScore = totalScore / float64(stats.MaliciousCount)
	}

	return stats
}

// analyzeCountry analyzes malicious activity within a country
func (c *Classifier) analyzeCountry(countryCode string, ranges []models.IPRange) *models.CountryStats {
	stats := &models.CountryStats{
		CountryCode:    countryCode,
		MaliciousCount: 0,
		TotalSubnets:   len(ranges),
		AvgScore:       0,
		MaxScore:       0,
	}

	maliciousIPs := c.ofClient.GetMaliciousIPsByCountry(countryCode)

	totalScore := 0.0
	for _, malIP := range maliciousIPs {
		stats.MaliciousCount++
		totalScore += malIP.Score

		if malIP.Score > stats.MaxScore {
			stats.MaxScore = malIP.Score
		}
	}

	if stats.MaliciousCount > 0 {
		stats.AvgScore = totalScore / float64(stats.MaliciousCount)
	}

	return stats
}

// ipToUint32 converts an IP address to uint32 for comparison
func ipToUint32(ip net.IP) uint32 {
	ip = ip.To4()
	if ip == nil {
		return 0
	}
	return uint32(ip[0])<<24 | uint32(ip[1])<<16 | uint32(ip[2])<<8 | uint32(ip[3])
}
