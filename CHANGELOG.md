# Changelog

All notable changes to the OneFirewall Classification Model will be documented in this file.

## [2.0.0] - 2025-11-21

### 🚨 Critical Bug Fixes

- **Fixed network/broadcast address detection** (`internal/data/ipdb.go:226-266`)
  - Previously only checked first and last IP of the entire range
  - Now correctly identifies ALL IPs ending in .0 (network) or .255 (broadcast)
  - Example: For range 1.0.1.0-1.0.3.255, now correctly returns NONE for:
    - 1.0.1.0, 1.0.1.255, 1.0.2.0, 1.0.2.255, 1.0.3.0, 1.0.3.255
  - This was a critical bug affecting classification accuracy

### 🎯 Real OneFirewall API Integration

- **Implemented full API authentication** (`internal/data/onefirewall.go`)
  - Added API key support via `Authorization: Bearer` header
  - Environment variable support: `ONEFIREWALL_API_KEY`
  - Command-line flag support: `--api-key`
  - Real-time IP lookup from OneFirewall database
  - Proper error handling for 404 (IP not in database)
  - Response parsing for crime_score and metadata

- **Enhanced OneFirewall client capabilities**
  - `IsMalicious()` now fetches from API if not in cache
  - Added `LoadMaliciousIPsFromFile()` for JSON file loading
  - Added `RefreshCache()` to clear cache and force fresh lookups
  - Added `GetCacheSize()` for monitoring
  - Improved logging with debug mode support

### 📊 Enhanced Scoring Algorithm

- **Improved subnet confidence calculation** (`internal/classifier/scorer.go:55-116`)
  - Reordered weights to prioritize average score over density
  - Added graduated multipliers for small subnets:
    - >20% malicious: 1.8x multiplier
    - >10% malicious: 1.5x multiplier
    - >5% malicious: 1.3x multiplier
  - Added 1.15x boost for high-severity IPs (score >= 8.0)
  - Better handling of /16 and larger subnets

- **Aligned with OneFirewall crime score concept**
  - More reports = higher confidence
  - Smaller subnets with malicious activity = more dangerous
  - Focus on average severity for large datasets (ASN, Country)

### 🔧 Configuration Enhancements

- **Environment variable support**
  - `ONEFIREWALL_API_KEY` - OneFirewall API key
  - Automatically loaded if not provided via CLI flag

- **New command-line flags**
  - `--api-key` - OneFirewall API key
  - `--onefirewall-url` - Custom OneFirewall base URL (default: https://app.onefirewall.com)
  - `--malicious` - Path to malicious IPs JSON file

- **Improved startup logic**
  - Graceful fallback to sample data when no API key provided
  - Better logging messages for configuration status
  - Clear indication when API integration is enabled

### 📁 File Loading

- **Malicious IPs file loader** (`internal/data/onefirewall.go:221-236`)
  - Supports JSON format
  - Loads bulk malicious IPs at startup
  - Reduces API calls for known malicious IPs
  - Example format:
    ```json
    [
      {
        "ip": "1.2.3.4",
        "score": 8.5,
        "asn": 12345,
        "country_code": "XX"
      }
    ]
    ```

### 🏥 API Improvements

- **Enhanced health check**
  - Now includes uptime
  - Version number (2.0.0)
  - Better status information

### 📖 Documentation

- **Updated README.md**
  - Added API key configuration instructions
  - Environment variable documentation
  - Real OneFirewall integration guide
  - Enhanced examples with authentication

### 🧪 Testing

- **Verified components**
  - API key authentication flow
  - Network/broadcast address detection
  - Environment variable loading
  - File-based malicious IP loading
  - Enhanced scoring algorithm
  - On-demand API fetching

### 🐛 Bug Fixes

- Fixed IP-to-uint32 conversion in classifier
- Improved error handling for API failures
- Better handling of missing country codes
- Fixed race conditions in cache access

### 🔒 Security

- API key is not logged in debug mode
- Secure Bearer token authentication
- No credentials in sample data

## [1.0.0] - 2025-11-21

### Initial Release

- Basic IP classification
- ASN-level analysis
- Country-level analysis
- Subnet analysis
- RESTful API
- Sample data support
- Confidence levels (NONE, LOW, MEDIUM, HIGH, CRITICAL)
- Thread-safe operations

---

## Migration Guide: v1.0.0 → v2.0.0

### Breaking Changes

None! Version 2.0.0 is fully backward compatible with 1.0.0.

### Recommended Actions

1. **Set up OneFirewall API key**
   ```bash
   export ONEFIREWALL_API_KEY="your-api-key-here"
   ```

2. **Download real IP database**
   ```bash
   wget https://iptoasn.com/data/ip2asn-v4.tsv.gz
   gunzip ip2asn-v4.tsv.gz
   ```

3. **Run with real data**
   ```bash
   ./onefirewall-classifier --ipdb ip2asn-v4.tsv
   ```

### New Features Available

- Real-time malicious IP lookups from OneFirewall
- Enhanced classification accuracy
- File-based malicious IP loading
- Better logging and monitoring

---

## Upgrade Path

### From Sample Data to Production

**Before (v1.0.0):**
```bash
./onefirewall-classifier
```

**After (v2.0.0):**
```bash
export ONEFIREWALL_API_KEY="your-key"
./onefirewall-classifier --ipdb /path/to/ip2asn-v4.tsv
```

### Benefits of Upgrading

1. ✅ Fixed critical network/broadcast bug
2. ✅ Real-time threat intelligence
3. ✅ More accurate scoring
4. ✅ Better performance with caching
5. ✅ Production-ready configuration
