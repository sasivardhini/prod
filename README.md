# OneFirewall Classification Model

A high-performance IP classification service for OneFirewall that provides confidence levels (NONE, LOW, MEDIUM, HIGH, CRITICAL) for IP addresses based on ASN, Country, and Network intelligence.

## Overview

This service analyzes IP addresses, ASNs, and countries to determine the likelihood of involvement in cyber attacks. It combines network topology data with historical malicious activity to provide real-time threat assessments.

## Features

- **Fast IP Classification**: Instantly classify any IPv4 address
- **ASN-Level Analysis**: Evaluate entire Autonomous Systems
- **Country-Level Intelligence**: Assess country-wide threat patterns
- **Subnet Analysis**: Deep analysis of network ranges
- **RESTful API**: Easy-to-use HTTP API endpoints
- **Scoring Algorithm**: Sophisticated multi-factor scoring system
- **Sample Data**: Built-in sample data for quick testing

## Architecture

The system consists of several key components:

1. **IP Database (IPDB)**: Manages IP-to-ASN and IP-to-Country mappings from iptoasn.com
2. **OneFirewall Client**: Integrates with OneFirewall's malicious IP database
3. **Classifier**: Core classification engine with multi-factor analysis
4. **Scorer**: Advanced scoring algorithm considering:
   - Malicious IP density within subnets
   - Average and maximum severity scores
   - Subnet size (smaller subnets = higher confidence)
   - ASN and country-level patterns
5. **REST API**: HTTP endpoints for classification requests

## Classification Logic

### Basic Rules

1. **AS=0 (Not-Routed)** → NONE
2. **Network Address** (first IP) → NONE
3. **Broadcast Address** (last IP) → NONE

### Intelligence-Based Classification

For all other IPs, the system analyzes:

- **Subnet Level**: How many IPs in the same subnet are malicious? Smaller subnets with malicious activity are more concerning.
- **ASN Level**: Aggregate malicious activity across all subnets in an ASN
- **Country Level**: Country-wide patterns (uses conservative thresholds)

### Confidence Levels

- **NONE**: Score < 3.0 or not routable
- **LOW**: Score 3.0 - 4.9
- **MEDIUM**: Score 5.0 - 7.4
- **HIGH**: Score 7.5 - 8.9
- **CRITICAL**: Score ≥ 9.0

## API Endpoints

### Health Check
```
GET /health
GET /api/v1/health
```

### Classify IP
```
GET /api/v1/classify/ip/{ip}

Example: GET /api/v1/classify/ip/112.118.109.197

Response:
{
  "input": "112.118.109.197",
  "confidence_level": "HIGH",
  "score": 8.5,
  "asn": 9808,
  "country_code": "CN",
  "description": "China Mobile",
  "reason": "IP is known malicious with score 8.50",
  "malicious_count": 1
}
```

### Classify ASN
```
GET /api/v1/classify/asn/{asn}

Example: GET /api/v1/classify/asn/9808
         GET /api/v1/classify/asn/AS9808

Response:
{
  "input": "AS9808",
  "confidence_level": "HIGH",
  "score": 7.8,
  "asn": 9808,
  "reason": "ASN analysis: 4 malicious IPs, avg score: 7.90, max score: 9.10",
  "malicious_count": 4
}
```

### Classify Country
```
GET /api/v1/classify/country/{country}

Example: GET /api/v1/classify/country/CN

Response:
{
  "input": "CN",
  "confidence_level": "HIGH",
  "score": 7.5,
  "country_code": "CN",
  "reason": "Country analysis: 4 malicious IPs, avg score: 7.90, max score: 9.10",
  "malicious_count": 4
}
```

### Batch Classification
```
POST /api/v1/classify/batch

Request Body:
[
  {"ip": "112.118.109.197"},
  {"asn": 9808},
  {"country": "CN"}
]

Response:
{
  "results": [...],
  "total": 3
}
```

## Installation

### Prerequisites

- Go 1.21 or higher
- Make (optional, for using Makefile)

### Build from Source

```bash
# Clone the repository
git clone <repository-url>
cd onefirewall-classifier

# Install dependencies
go mod download

# Build the application
make build

# Or use go directly
go build -o build/onefirewall-classifier cmd/server/main.go
```

## Usage

### Running with Sample Data

```bash
# Using Make
make run

# Or directly
./build/onefirewall-classifier
```

The server will start on port 8080 with built-in sample data.

### Running with Custom Data

```bash
# With IP database from iptoasn.com
./build/onefirewall-classifier -ipdb data/ip2asn-v4.tsv

# With custom port
./build/onefirewall-classifier -port 9090

# With debug logging
./build/onefirewall-classifier -log-level debug

# All options
./build/onefirewall-classifier \
  -port 8080 \
  -ipdb data/ip2asn-v4.tsv \
  -malicious data/malicious-ips.json \
  -log-level info
```

### Command Line Options

- `-port`: Server port (default: 8080)
- `-ipdb`: Path to IP database file in TSV format
- `-malicious`: Path to malicious IPs file (optional)
- `-log-level`: Log level - debug, info, warn, error (default: info)

## Data Sources

### IP to ASN Database

Download the latest IP to ASN mappings from [iptoasn.com](https://iptoasn.com/):

```bash
# Download IP to ASN database
wget https://iptoasn.com/data/ip2asn-v4.tsv.gz
gunzip ip2asn-v4.tsv.gz
mv ip2asn-v4.tsv data/
```

**Format**: TSV with columns: `start_ip`, `end_ip`, `asn`, `country_code`, `description`

Example:
```
1.0.0.0	1.0.0.255	13335	US	CLOUDFLARENET
1.0.1.0	1.0.3.255	13335	US	CLOUDFLARENET
```

### OneFirewall Malicious IPs

The service can integrate with OneFirewall's API to fetch real-time malicious IP data. For testing, sample malicious IPs are included.

## Development

### Project Structure

```
.
├── cmd/
│   └── server/
│       └── main.go           # Application entry point
├── internal/
│   ├── api/
│   │   ├── handler.go        # HTTP request handlers
│   │   └── router.go         # Route configuration
│   ├── classifier/
│   │   ├── classifier.go     # Core classification logic
│   │   └── scorer.go         # Scoring algorithms
│   ├── data/
│   │   ├── ipdb.go          # IP database management
│   │   └── onefirewall.go   # OneFirewall API client
│   └── models/
│       └── types.go          # Data models
├── data/                     # Data files
├── Makefile                  # Build automation
├── go.mod                    # Go dependencies
└── README.md                 # This file
```

### Running Tests

```bash
# Run all tests
make test

# Run with coverage
make test-coverage

# View coverage report
open coverage.html
```

### Code Formatting

```bash
make fmt
```

### Building for Multiple Platforms

```bash
make build-all
```

This creates binaries for Linux, macOS, and Windows.

## Example Usage

### Test with cURL

```bash
# Health check
curl http://localhost:8080/health

# Classify an IP
curl http://localhost:8080/api/v1/classify/ip/112.118.109.197

# Classify an ASN
curl http://localhost:8080/api/v1/classify/asn/9808

# Classify a country
curl http://localhost:8080/api/v1/classify/country/CN

# Batch classification
curl -X POST http://localhost:8080/api/v1/classify/batch \
  -H "Content-Type: application/json" \
  -d '[
    {"ip": "112.118.109.197"},
    {"asn": 9808},
    {"country": "CN"}
  ]'
```

### Test with Python

```python
import requests

# Classify an IP
response = requests.get('http://localhost:8080/api/v1/classify/ip/112.118.109.197')
result = response.json()
print(f"Confidence: {result['confidence_level']}, Score: {result['score']}")

# Batch classification
batch_data = [
    {"ip": "112.118.109.197"},
    {"asn": 9808},
    {"country": "CN"}
]
response = requests.post('http://localhost:8080/api/v1/classify/batch', json=batch_data)
results = response.json()
print(f"Processed {results['total']} items")
```

## Scoring Algorithm Details

### Subnet Confidence

The subnet scoring considers:

1. **Density**: Percentage of malicious IPs in the subnet
2. **Subnet Size**: Smaller subnets receive higher weight (inverse logarithmic scale)
3. **Average Score**: Mean of all malicious IP scores in the subnet
4. **Max Score**: Highest individual IP score in the subnet

Formula:
```
score = (density * 0.4 * 10) + (avg_score * 0.3) + (max_score * 0.2) + (subnet_factor * 0.1 * 10)
```

Additional multiplier applied if >5% of a small subnet (<256 IPs) is malicious.

### ASN Confidence

ASN scoring focuses on:

1. **Malicious Count**: Logarithmic scale of malicious IP count
2. **Average Score**: Mean severity across the ASN
3. **Max Score**: Highest severity in the ASN

Multipliers applied at thresholds: 10, 50, 100 malicious IPs.

### Country Confidence

Country scoring uses a conservative approach:

1. **Malicious Count**: Logarithmic scale
2. **Average Score**: Heavily weighted (0.5)
3. **Max Score**: Lesser weight (0.2)

Conservative multipliers at 500 and 1000 malicious IPs.

## Performance

- **Classification Speed**: < 1ms per IP (with in-memory database)
- **Memory Usage**: ~100MB for 1M IP ranges
- **Concurrent Requests**: Thread-safe, supports high concurrency
- **API Response Time**: < 5ms average (excluding network latency)

## Deployment

### Docker Deployment

Create a `Dockerfile`:

```dockerfile
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o onefirewall-classifier cmd/server/main.go

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/onefirewall-classifier .
EXPOSE 8080
CMD ["./onefirewall-classifier"]
```

Build and run:

```bash
docker build -t onefirewall-classifier .
docker run -p 8080:8080 onefirewall-classifier
```

### Production Deployment

For production:

1. Download and mount real IP database from iptoasn.com
2. Configure OneFirewall API credentials
3. Use environment variables for configuration
4. Set up monitoring and logging
5. Deploy behind a load balancer
6. Enable rate limiting

## Contributing

1. Fork the repository
2. Create a feature branch
3. Make your changes
4. Run tests and formatting
5. Submit a pull request

## License

[Your License Here]

## Contact

For questions or support, please contact the OneFirewall team.

## Acknowledgments

- IP to ASN data from [iptoasn.com](https://iptoasn.com/)
- OneFirewall team for threat intelligence
