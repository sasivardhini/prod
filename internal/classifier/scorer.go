package classifier

import (
	"math"

	"github.com/onefirewall/classifier/internal/models"
)

// Scorer implements the scoring logic for classification
type Scorer struct {
	// Thresholds for OneFirewall scores
	ScoreThresholdLow      float64
	ScoreThresholdMedium   float64
	ScoreThresholdHigh     float64
	ScoreThresholdCritical float64

	// Weights for different factors
	SubnetDensityWeight float64
	AvgScoreWeight      float64
	MaxScoreWeight      float64
	SubnetSizeWeight    float64
}

// NewScorer creates a new scorer with default thresholds
func NewScorer() *Scorer {
	return &Scorer{
		// OneFirewall score thresholds
		ScoreThresholdLow:      3.0,
		ScoreThresholdMedium:   5.0,
		ScoreThresholdHigh:     7.5,
		ScoreThresholdCritical: 9.0,

		// Weights
		SubnetDensityWeight: 0.4,
		AvgScoreWeight:      0.3,
		MaxScoreWeight:      0.2,
		SubnetSizeWeight:    0.1,
	}
}

// ScoreToConfidence converts a numeric score to confidence level
func (s *Scorer) ScoreToConfidence(score float64) models.ConfidenceLevel {
	if score >= s.ScoreThresholdCritical {
		return models.ConfidenceCritical
	} else if score >= s.ScoreThresholdHigh {
		return models.ConfidenceHigh
	} else if score >= s.ScoreThresholdMedium {
		return models.ConfidenceMedium
	} else if score >= s.ScoreThresholdLow {
		return models.ConfidenceLow
	}
	return models.ConfidenceNone
}

// CalculateSubnetConfidence calculates confidence for a subnet
// Based on OneFirewall's crime score concept: more malicious IPs = higher confidence
func (s *Scorer) CalculateSubnetConfidence(stats *models.SubnetStats) (models.ConfidenceLevel, float64) {
	if stats.MaliciousCount == 0 {
		return models.ConfidenceNone, 0
	}

	// Calculate malicious density (percentage of malicious IPs)
	density := 0.0
	if stats.TotalIPs > 0 {
		density = float64(stats.MaliciousCount) / float64(stats.TotalIPs)
	}

	// Smaller subnets with malicious IPs are more concerning
	// Normalize subnet size (256 IPs = /24, 65536 IPs = /16)
	subnetSizeFactor := 1.0
	if stats.TotalIPs > 0 {
		// Inverse relationship: smaller subnet = higher factor
		// For small subnets (/24 or smaller), this gives high weight
		subnetSizeFactor = math.Max(0.1, 1.0-math.Log10(float64(stats.TotalIPs))/5.0)
	}

	// Normalize average score to 0-10 scale (OneFirewall crime scores)
	normalizedAvgScore := math.Min(10.0, stats.AvgScore)
	normalizedMaxScore := math.Min(10.0, stats.MaxScore)

	// Weighted score calculation
	// Priority: average score > density > max score > subnet size
	score := (normalizedAvgScore * s.AvgScoreWeight) +
		(density * s.SubnetDensityWeight * 10.0) +
		(normalizedMaxScore * s.MaxScoreWeight) +
		(subnetSizeFactor * s.SubnetSizeWeight * 10.0)

	// Apply multipliers based on malicious IP count and density
	// Logic: If subnet has many malicious IPs, it's more dangerous
	if stats.TotalIPs <= 256 { // /24 or smaller subnet
		if density > 0.20 { // More than 20% malicious
			score *= 1.8
		} else if density > 0.10 { // More than 10% malicious
			score *= 1.5
		} else if density > 0.05 { // More than 5% malicious
			score *= 1.3
		}
	} else if stats.TotalIPs <= 65536 { // /16 subnet
		if stats.MaliciousCount >= 100 {
			score *= 1.4
		} else if stats.MaliciousCount >= 50 {
			score *= 1.2
		}
	}

	// Boost score if we have high-severity malicious IPs (score >= 8)
	if stats.MaxScore >= 8.0 {
		score *= 1.15
	}

	// Cap the score at 10
	score = math.Min(10.0, score)

	confidence := s.ScoreToConfidence(score)
	return confidence, score
}

// CalculateASNConfidence calculates confidence for an ASN
func (s *Scorer) CalculateASNConfidence(stats *models.ASNStats) (models.ConfidenceLevel, float64) {
	if stats.MaliciousCount == 0 {
		return models.ConfidenceNone, 0
	}

	// For ASNs, we focus more on the quantity and severity of malicious IPs
	// rather than density (since ASNs can be very large)

	// Normalize counts (logarithmic scale for large numbers)
	countFactor := math.Log10(float64(stats.MaliciousCount) + 1)

	// Normalize scores
	normalizedAvgScore := math.Min(10.0, stats.AvgScore)
	normalizedMaxScore := math.Min(10.0, stats.MaxScore)

	// Calculate score
	score := (countFactor * 1.5) + // More weight on malicious count
		(normalizedAvgScore * 0.4) +
		(normalizedMaxScore * 0.3)

	// Apply multipliers based on thresholds
	if stats.MaliciousCount >= 100 {
		score *= 1.3
	} else if stats.MaliciousCount >= 50 {
		score *= 1.2
	} else if stats.MaliciousCount >= 10 {
		score *= 1.1
	}

	// Cap the score at 10
	score = math.Min(10.0, score)

	confidence := s.ScoreToConfidence(score)
	return confidence, score
}

// CalculateCountryConfidence calculates confidence for a country
func (s *Scorer) CalculateCountryConfidence(stats *models.CountryStats) (models.ConfidenceLevel, float64) {
	if stats.MaliciousCount == 0 {
		return models.ConfidenceNone, 0
	}

	// For countries, we use a more conservative approach
	// Countries are very large, so we focus on average severity

	// Normalize counts (logarithmic scale)
	countFactor := math.Log10(float64(stats.MaliciousCount) + 1)

	// Normalize scores
	normalizedAvgScore := math.Min(10.0, stats.AvgScore)
	normalizedMaxScore := math.Min(10.0, stats.MaxScore)

	// Calculate score with more weight on average score
	score := (countFactor * 1.0) +
		(normalizedAvgScore * 0.5) +
		(normalizedMaxScore * 0.2)

	// Conservative multipliers for countries
	if stats.MaliciousCount >= 1000 {
		score *= 1.2
	} else if stats.MaliciousCount >= 500 {
		score *= 1.1
	}

	// Cap the score at 10
	score = math.Min(10.0, score)

	confidence := s.ScoreToConfidence(score)
	return confidence, score
}

// CalculateDynamicScore calculates a score based on custom parameters
func (s *Scorer) CalculateDynamicScore(maliciousCount int, totalSize int, avgScore, maxScore float64) (models.ConfidenceLevel, float64) {
	if maliciousCount == 0 {
		return models.ConfidenceNone, 0
	}

	density := 0.0
	if totalSize > 0 {
		density = float64(maliciousCount) / float64(totalSize)
	}

	normalizedAvgScore := math.Min(10.0, avgScore)
	normalizedMaxScore := math.Min(10.0, maxScore)

	score := (density * 4.0) +
		(normalizedAvgScore * 0.4) +
		(normalizedMaxScore * 0.3) +
		(math.Log10(float64(maliciousCount)+1) * 0.3)

	score = math.Min(10.0, score)

	confidence := s.ScoreToConfidence(score)
	return confidence, score
}
