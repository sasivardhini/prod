package data

import (
	"bufio"
	"encoding/binary"
	"fmt"
	"net"
	"os"
	"strconv"
	"strings"
	"sync"

	"github.com/onefirewall/classifier/internal/models"
	log "github.com/sirupsen/logrus"
)

// IPDB manages the IP to ASN/Country database
type IPDB struct {
	ranges []models.IPRange
	mu     sync.RWMutex
}

// NewIPDB creates a new IP database instance
func NewIPDB() *IPDB {
	return &IPDB{
		ranges: make([]models.IPRange, 0),
	}
}

// LoadFromFile loads IP ranges from a TSV file (iptoasn.com format)
// Format: start_ip	end_ip	asn	country	description
func (db *IPDB) LoadFromFile(filepath string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	file, err := os.Open(filepath)
	if err != nil {
		return fmt.Errorf("failed to open file: %w", err)
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		// Skip empty lines and comments
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Split(line, "\t")
		if len(parts) < 4 {
			log.Warnf("Invalid line %d: %s", lineNum, line)
			continue
		}

		startIP := net.ParseIP(parts[0])
		endIP := net.ParseIP(parts[1])

		if startIP == nil || endIP == nil {
			log.Warnf("Invalid IP addresses at line %d", lineNum)
			continue
		}

		asn, err := strconv.ParseUint(parts[2], 10, 32)
		if err != nil {
			log.Warnf("Invalid ASN at line %d: %s", lineNum, parts[2])
			continue
		}

		countryCode := parts[3]
		description := ""
		if len(parts) > 4 {
			description = parts[4]
		}

		ipRange := models.IPRange{
			StartIP:     startIP,
			EndIP:       endIP,
			StartIPInt:  ipToUint32(startIP),
			EndIPInt:    ipToUint32(endIP),
			ASN:         uint32(asn),
			CountryCode: countryCode,
			Description: description,
		}

		db.ranges = append(db.ranges, ipRange)
	}

	if err := scanner.Err(); err != nil {
		return fmt.Errorf("error reading file: %w", err)
	}

	log.Infof("Loaded %d IP ranges", len(db.ranges))
	return nil
}

// LoadFromTSVData loads IP ranges from TSV data string
func (db *IPDB) LoadFromTSVData(data string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	scanner := bufio.NewScanner(strings.NewReader(data))
	lineNum := 0

	for scanner.Scan() {
		lineNum++
		line := strings.TrimSpace(scanner.Text())

		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}

		parts := strings.Split(line, "\t")
		if len(parts) < 4 {
			continue
		}

		startIP := net.ParseIP(parts[0])
		endIP := net.ParseIP(parts[1])

		if startIP == nil || endIP == nil {
			continue
		}

		asn, err := strconv.ParseUint(parts[2], 10, 32)
		if err != nil {
			continue
		}

		countryCode := parts[3]
		description := ""
		if len(parts) > 4 {
			description = parts[4]
		}

		ipRange := models.IPRange{
			StartIP:     startIP,
			EndIP:       endIP,
			StartIPInt:  ipToUint32(startIP),
			EndIPInt:    ipToUint32(endIP),
			ASN:         uint32(asn),
			CountryCode: countryCode,
			Description: description,
		}

		db.ranges = append(db.ranges, ipRange)
	}

	log.Infof("Loaded %d IP ranges from TSV data", len(db.ranges))
	return nil
}

// LookupIP finds the IP range containing the given IP address
func (db *IPDB) LookupIP(ipStr string) (*models.IPRange, error) {
	db.mu.RLock()
	defer db.mu.RUnlock()

	ip := net.ParseIP(ipStr)
	if ip == nil {
		return nil, fmt.Errorf("invalid IP address: %s", ipStr)
	}

	ipInt := ipToUint32(ip)

	// Binary search would be more efficient if ranges are sorted
	for i := range db.ranges {
		if ipInt >= db.ranges[i].StartIPInt && ipInt <= db.ranges[i].EndIPInt {
			return &db.ranges[i], nil
		}
	}

	return nil, fmt.Errorf("IP not found in database: %s", ipStr)
}

// GetRangesByASN returns all IP ranges for a given ASN
func (db *IPDB) GetRangesByASN(asn uint32) []models.IPRange {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var result []models.IPRange
	for _, r := range db.ranges {
		if r.ASN == asn {
			result = append(result, r)
		}
	}
	return result
}

// GetRangesByCountry returns all IP ranges for a given country
func (db *IPDB) GetRangesByCountry(countryCode string) []models.IPRange {
	db.mu.RLock()
	defer db.mu.RUnlock()

	var result []models.IPRange
	for _, r := range db.ranges {
		if strings.EqualFold(r.CountryCode, countryCode) {
			result = append(result, r)
		}
	}
	return result
}

// GetAllRanges returns all IP ranges (use with caution for large datasets)
func (db *IPDB) GetAllRanges() []models.IPRange {
	db.mu.RLock()
	defer db.mu.RUnlock()

	result := make([]models.IPRange, len(db.ranges))
	copy(result, db.ranges)
	return result
}

// ipToUint32 converts an IP address to uint32 for comparison
func ipToUint32(ip net.IP) uint32 {
	ip = ip.To4()
	if ip == nil {
		return 0
	}
	return binary.BigEndian.Uint32(ip)
}

// IsNetworkAddress checks if an IP is a network address (last octet is 0)
// Network addresses are not routable and should return NONE
func IsNetworkAddress(ip net.IP, ipRange *models.IPRange) bool {
	ipv4 := ip.To4()
	if ipv4 == nil {
		return false
	}

	// Check if last octet is 0 (network address)
	// This handles cases like x.x.x.0 within any subnet
	if ipv4[3] == 0 {
		ipInt := ipToUint32(ip)
		// Verify IP is within the given range
		if ipInt >= ipRange.StartIPInt && ipInt <= ipRange.EndIPInt {
			return true
		}
	}

	return false
}

// IsBroadcastAddress checks if an IP is a broadcast address (last octet is 255)
// Broadcast addresses are not routable and should return NONE
func IsBroadcastAddress(ip net.IP, ipRange *models.IPRange) bool {
	ipv4 := ip.To4()
	if ipv4 == nil {
		return false
	}

	// Check if last octet is 255 (broadcast address)
	// This handles cases like x.x.x.255 within any subnet
	if ipv4[3] == 255 {
		ipInt := ipToUint32(ip)
		// Verify IP is within the given range
		if ipInt >= ipRange.StartIPInt && ipInt <= ipRange.EndIPInt {
			return true
		}
	}

	return false
}

// GetSubnetSize calculates the number of IPs in a subnet
func GetSubnetSize(ipRange *models.IPRange) uint32 {
	if ipRange.EndIPInt >= ipRange.StartIPInt {
		return ipRange.EndIPInt - ipRange.StartIPInt + 1
	}
	return 0
}
