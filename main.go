package main

import (
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/onefirewall/classifier/internal/api"
	"github.com/onefirewall/classifier/internal/classifier"
	"github.com/onefirewall/classifier/internal/data"
	"github.com/onefirewall/classifier/internal/models"
	log "github.com/sirupsen/logrus"
)

func main() {
	// Command line flags
	port := flag.String("port", "8080", "Server port")
	ipdbFile := flag.String("ipdb", "", "Path to IP database file (TSV format)")
	maliciousFile := flag.String("malicious", "", "Path to malicious IPs file (JSON format)")
	logLevel := flag.String("log-level", "info", "Log level (debug, info, warn, error)")
	apiKey := flag.String("api-key", "", "OneFirewall API key (can also use ONEFIREWALL_API_KEY env var)")
	baseURL := flag.String("onefirewall-url", "https://app.onefirewall.com", "OneFirewall base URL")
	flag.Parse()

	// Configure logging
	level, err := log.ParseLevel(*logLevel)
	if err != nil {
		level = log.InfoLevel
	}
	log.SetLevel(level)
	log.SetFormatter(&log.TextFormatter{
		FullTimestamp: true,
	})

	log.Info("Starting OneFirewall Classifier Service...")

	// Get API key from environment if not provided via flag
	oneFirewallAPIKey := *apiKey
	if oneFirewallAPIKey == "" {
		oneFirewallAPIKey = os.Getenv("ONEFIREWALL_API_KEY")
	}

	if oneFirewallAPIKey != "" {
		log.Info("OneFirewall API key configured - real-time API integration enabled")
	} else {
		log.Warn("No OneFirewall API key provided - will use cached data only")
	}

	// Initialize IP database
	ipDB := data.NewIPDB()

	if *ipdbFile != "" {
		log.Infof("Loading IP database from: %s", *ipdbFile)
		if err := ipDB.LoadFromFile(*ipdbFile); err != nil {
			log.Fatalf("Failed to load IP database: %v", err)
		}
	} else {
		log.Warn("No IP database file specified. Using sample data...")
		loadSampleData(ipDB)
	}

	// Initialize OneFirewall client with API key
	ofClient := data.NewOneFirewallClient(*baseURL, oneFirewallAPIKey)

	// Load malicious IPs if provided
	if *maliciousFile != "" {
		log.Infof("Loading malicious IPs from: %s", *maliciousFile)
		if err := ofClient.LoadMaliciousIPsFromFile(*maliciousFile); err != nil {
			log.Errorf("Failed to load malicious IPs from file: %v", err)
			log.Warn("Falling back to sample malicious data...")
			loadSampleMaliciousIPs(ofClient, ipDB)
		}
	} else {
		if oneFirewallAPIKey == "" {
			log.Warn("No malicious IPs file specified and no API key. Using sample malicious data...")
			loadSampleMaliciousIPs(ofClient, ipDB)
		} else {
			log.Info("No malicious IPs file specified. Will fetch data from OneFirewall API on-demand.")
		}
	}

	// Initialize classifier
	classifierInstance := classifier.NewClassifier(ipDB, ofClient)

	// Setup API
	handler := api.NewHandler(classifierInstance)
	router := api.SetupRouter(handler)

	// Start server
	addr := fmt.Sprintf(":%s", *port)
	server := &http.Server{
		Addr:         addr,
		Handler:      router,
		ReadTimeout:  15 * time.Second,
		WriteTimeout: 15 * time.Second,
		IdleTimeout:  60 * time.Second,
	}

	log.Infof("Server starting on %s", addr)
	log.Info("API endpoints:")
	log.Info("  GET  /health")
	log.Info("  GET  /api/v1/health")
	log.Info("  GET  /api/v1/classify/ip/{ip}")
	log.Info("  GET  /api/v1/classify/asn/{asn}")
	log.Info("  GET  /api/v1/classify/country/{country}")
	log.Info("  POST /api/v1/classify/batch")

	if err := server.ListenAndServe(); err != nil {
		log.Fatalf("Server failed to start: %v", err)
	}
}

// loadSampleData loads sample IP ranges for testing
func loadSampleData(ipDB *data.IPDB) {
	sampleData := `1.0.0.0	1.0.0.255	13335	US	CLOUDFLARENET
1.0.1.0	1.0.3.255	13335	US	CLOUDFLARENET
8.8.8.0	8.8.8.255	15169	US	GOOGLE
8.8.4.0	8.8.4.255	15169	US	GOOGLE
80.20.29.0	80.20.92.255	12345	GB	Example ISP
112.118.109.0	112.118.109.255	9808	CN	China Mobile
192.168.0.0	192.168.255.255	0	ZZ	Private Network
10.0.0.0	10.255.255.255	0	ZZ	Private Network
172.16.0.0	172.31.255.255	0	ZZ	Private Network
45.32.0.0	45.32.255.255	20473	US	AS-CHOOPA
185.220.100.0	185.220.103.255	209	NL	ASNL
195.123.220.0	195.123.223.255	44050	GB	PONYNET`

	if err := ipDB.LoadFromTSVData(sampleData); err != nil {
		log.Errorf("Failed to load sample data: %v", err)
	} else {
		log.Info("Sample IP database loaded successfully")
	}
}

// loadSampleMaliciousIPs loads sample malicious IPs for testing
func loadSampleMaliciousIPs(ofClient *data.OneFirewallClient, ipDB *data.IPDB) {
	// Sample malicious IPs based on the example from OneFirewall
	sampleMalicious := []struct {
		ip    string
		score float64
	}{
		{"112.118.109.197", 8.5},
		{"112.118.109.45", 7.2},
		{"112.118.109.88", 6.8},
		{"112.118.109.123", 9.1},
		{"80.20.30.15", 5.5},
		{"80.20.45.88", 6.0},
		{"80.20.60.123", 5.8},
		{"80.20.75.200", 7.5},
		{"45.32.15.88", 8.0},
		{"45.32.45.123", 7.8},
		{"185.220.101.5", 9.5},
		{"185.220.101.15", 9.2},
		{"185.220.101.25", 9.8},
		{"195.123.221.10", 6.5},
	}

	maliciousIPs := make([]models.MaliciousIP, 0, len(sampleMalicious))

	for _, sm := range sampleMalicious {
		// Lookup ASN and country for each IP
		ipRange, err := ipDB.LookupIP(sm.ip)
		var asn uint32
		var country string

		if err == nil {
			asn = ipRange.ASN
			country = ipRange.CountryCode
		}

		maliciousIPs = append(maliciousIPs, models.MaliciousIP{
			IP:          sm.ip,
			Score:       sm.score,
			ASN:         asn,
			CountryCode: country,
			FirstSeen:   time.Now().Add(-30 * 24 * time.Hour),
			LastSeen:    time.Now().Add(-2 * time.Hour),
		})
	}

	ofClient.LoadMaliciousIPs(maliciousIPs)
	log.Infof("Loaded %d sample malicious IPs", len(maliciousIPs))
}
