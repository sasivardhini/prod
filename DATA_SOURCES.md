# OneFirewall Classifier - Data Sources

## Overview

The OneFirewall Classifier uses **two primary databases** for IP classification:

1. **IP-to-ASN Database** (`ip2asn-v4.tsv`)
2. **IP-to-Country Database** (`ip2country-v4.tsv`)

Both databases are sourced from [iptoasn.com](https://iptoasn.com/) and provide comprehensive coverage of the IPv4 address space.

---

## Database 1: IP-to-ASN Mapping

**File:** `data/ip2asn-v4.tsv`
**Size:** ~28 MB
**Records:** 512,027 IP ranges
**Source:** https://iptoasn.com/

### Format
```
start_ip | end_ip | asn | country_code | description
```

### Example
```
1.0.0.0     1.0.0.255   13335   US   CLOUDFLARENET
8.8.8.0     8.8.8.255   15169   US   GOOGLE
112.118.109.0  112.118.109.255  4760  HK  HKTIMS-AP HKT Limited
```

### Purpose
- Maps IP addresses to Autonomous System Numbers (ASN)
- Identifies the network provider/organization
- Used for ASN-level threat aggregation
- Provides network infrastructure context

### Updates
- Download latest version from: https://iptoasn.com/
- Recommended update frequency: Monthly
- File format: TSV (Tab-Separated Values)

---

## Database 2: IP-to-Country Mapping

**File:** `data/ip2country-v4.tsv`
**Size:** ~16 MB
**Records:** 512,027 IP ranges
**Source:** Extracted from iptoasn.com data

### Format
```
start_ip | end_ip | country_code
```

### Example
```
1.0.0.0     1.0.0.255   US
8.8.8.0     8.8.8.255   US
112.118.109.0  112.118.109.255  HK
185.220.101.0  185.220.101.255  DE
```

### Purpose
- Maps IP addresses to country codes (ISO 3166-1 alpha-2)
- Enables country-level threat analysis
- Used for geographic threat distribution
- Supports country-based classification

### Country Codes
- **US** - United States
- **CN** - China
- **HK** - Hong Kong
- **DE** - Germany
- **RU** - Russia
- **None** - Not routed/Private

### Updates
- Regenerate from ip2asn-v4.tsv after updates:
  ```bash
  awk -F'\t' '{print $1"\t"$2"\t"$4}' ip2asn-v4.tsv > ip2country-v4.tsv
  ```

---

## Database 3: OneFirewall Threat Intelligence (API)

**Endpoint:** `https://app.onefirewall.com/api/v1/ips`
**Method:** Real-time API calls
**Authentication:** Bearer token (JWT)

### Purpose
- Provides real-time malicious IP data
- Returns crime scores (0-1000+ raw, normalized to 0-10)
- Includes threat event statistics
- Tracks first seen / last seen timestamps

### Response Format
```json
{
  "header": {
    "page_size": 1,
    "live": true
  },
  "body": [{
    "ip": "112.118.109.197",
    "score": 234,
    "entry_ts": 1760746505,
    "ts": 1763729619,
    "tags": ["honeypot", "Telnet", "cta"],
    "ip_info": {
      "asn": "AS4760",
      "country": "Hong Kong",
      "country_code": "HK",
      "as_name": "HKT Limited"
    },
    "info": {
      "members": 10,
      "events": 12,
      "sources": ["router", "pfsense", "CTA"]
    }
  }]
}
```

### Score Normalization
OneFirewall returns raw scores (0-1000+), which are normalized to 0-10:

| Raw Score | Normalized | Confidence |
|-----------|------------|------------|
| 0-10      | 0-3        | LOW        |
| 10-50     | 3-5        | LOW-MEDIUM |
| 50-150    | 5-7        | MEDIUM     |
| 150-300   | 7-9        | HIGH       |
| 300+      | 9-10       | CRITICAL   |

**Example:**
- Raw score 234 → Normalized 8.5 → HIGH confidence
- Raw score 13 → Normalized 3.15 → LOW confidence

### Configuration
Set environment variable:
```bash
export ONEFIREWALL_API_KEY="your-api-key-here"
```

Or add to `.env`:
```
ONEFIREWALL_API_KEY=your-api-key-here
```

---

## How the Databases Work Together

### Classification Flow

```
1. User queries IP: 112.118.109.197
   ↓
2. Lookup in ip2asn-v4.tsv
   → ASN: 4760
   → Country: HK
   → Provider: HKT Limited
   ↓
3. Check OneFirewall API
   → Score: 234 (raw)
   → Normalized: 8.5
   → Confidence: HIGH
   ↓
4. Analyze subnet (all IPs in 112.118.109.0/24)
   → Count malicious IPs
   → Calculate density
   → Apply scoring algorithm
   ↓
5. Return classification result
```

### ASN-Level Analysis
```
1. User queries ASN: AS4760
   ↓
2. Find all IP ranges in ip2asn-v4.tsv with ASN=4760
   ↓
3. Query OneFirewall for malicious IPs in those ranges
   ↓
4. Aggregate statistics:
   - Total malicious IPs
   - Average score
   - Max score
   ↓
5. Calculate ASN confidence level
```

### Country-Level Analysis
```
1. User queries Country: HK
   ↓
2. Find all IP ranges in ip2country-v4.tsv with country=HK
   ↓
3. Query OneFirewall for malicious IPs in Hong Kong ranges
   ↓
4. Aggregate statistics:
   - Total malicious IPs
   - Average score
   - Distribution across ASNs
   ↓
5. Calculate country confidence level
```

---

## Data Statistics

### Coverage
- **Total IPv4 ranges:** 512,027
- **Unique ASNs:** ~100,000
- **Countries covered:** 249
- **Address space:** Complete IPv4 (0.0.0.0 - 255.255.255.255)

### Top ASNs by Range Count
1. AS0 (not-routed): ~15%
2. Cloud providers (AWS, Google, Azure): ~8%
3. Telecom providers: ~45%
4. Regional ISPs: ~32%

### Geographic Distribution
- North America: ~35%
- Europe: ~28%
- Asia Pacific: ~25%
- Other regions: ~12%

---

## Maintenance

### Updating Databases

**Monthly Update Procedure:**

1. **Download latest IP-to-ASN database:**
   ```bash
   cd data/
   wget https://iptoasn.com/data/ip2asn-v4.tsv.gz
   gunzip ip2asn-v4.tsv.gz
   ```

2. **Regenerate IP-to-Country database:**
   ```bash
   awk -F'\t' '{print $1"\t"$2"\t"$4}' ip2asn-v4.tsv > ip2country-v4.tsv
   ```

3. **Verify integrity:**
   ```bash
   wc -l ip2asn-v4.tsv ip2country-v4.tsv
   # Should show same number of lines
   ```

4. **Test with classifier:**
   ```bash
   ./classify.exe -ipdb data/ip2asn-v4.tsv
   curl http://localhost:8080/api/v1/classify/ip/8.8.8.8
   ```

5. **Commit changes:**
   ```bash
   git add data/ip2asn-v4.tsv data/ip2country-v4.tsv
   git commit -m "Update IP databases ($(date +%Y-%m))"
   git push
   ```

---

## References

- **iptoasn.com:** https://iptoasn.com/
- **OneFirewall:** https://app.onefirewall.com/
- **ISO Country Codes:** https://en.wikipedia.org/wiki/ISO_3166-1_alpha-2
- **ASN Registry:** https://www.iana.org/assignments/as-numbers/as-numbers.xhtml

---

## License

The IP databases from iptoasn.com are provided under Creative Commons Attribution 4.0.

OneFirewall threat intelligence is proprietary and requires an API key.
