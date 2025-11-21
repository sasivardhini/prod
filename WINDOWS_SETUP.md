# 🪟 OneFirewall Classifier - Windows Setup Guide

Complete setup guide for running OneFirewall Classifier on Windows.

## Prerequisites

### 1. Install Go

1. Download Go from: https://go.dev/dl/
2. Install the Windows MSI installer
3. Verify installation:
   ```cmd
   go version
   ```

### 2. Install Git (Optional)

For cloning the repository: https://git-scm.com/download/win

---

## Quick Start (3 Steps)

### Step 1: Setup

Run the setup script to build the application:

```cmd
setup.bat
```

This will:
- ✅ Check if Go is installed
- ✅ Create necessary directories
- ✅ Download Go dependencies
- ✅ Build the application
- ✅ Create default .env file

### Step 2: Configure

Edit the `.env` file and add your OneFirewall API key:

```env
ONEFIREWALL_API_KEY=your-api-key-here
```

**Your API Key:**
```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ0cyI6MTc1NjYwMTIwNywiZ3VpZCI6Ik9GQS1HVUlELTU3MjQtMjAwNy0wMjMyIiwiaWF0IjoxNzU2NjAxMjA3fQ.tPdQGERKfN7mEALoEfn_yk9RubsxZj7pFIXmmT8-F7I
```

### Step 3: Run

```cmd
run.bat
```

The server will start at: **http://localhost:8080**

---

## Manual Setup

If you prefer to do it manually:

### 1. Build

```cmd
mkdir build
go build -o build\onefirewall-classifier.exe cmd\server\main.go
```

### 2. Set Environment Variable

```cmd
set ONEFIREWALL_API_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ0cyI6MTc1NjYwMTIwNywiZ3VpZCI6Ik9GQS1HVUlELTU3MjQtMjAwNy0wMjMyIiwiaWF0IjoxNzU2NjAxMjA3fQ.tPdQGERKfN7mEALoEfn_yk9RubsxZj7pFIXmmT8-F7I
```

### 3. Run

```cmd
build\onefirewall-classifier.exe
```

---

## Download Full IP Database (Optional)

For production use, download the full IP-to-ASN database:

### Option A: Using Script (Recommended)

```cmd
download-ipdb.bat
```

### Option B: Manual Download

1. Visit: https://iptoasn.com/
2. Download: `ip2asn-v4.tsv.gz` (~20-30 MB compressed)
3. Extract the `.gz` file (use 7-Zip or Windows built-in)
4. Move `ip2asn-v4.tsv` to the `data\` folder

Then run with:
```cmd
build\onefirewall-classifier.exe -ipdb data\ip2asn-v4.tsv
```

Or just use `run.bat` - it will automatically detect and use the database.

---

## Testing the API

### Option A: Using Test Script

```cmd
test-api.bat
```

This runs all test cases automatically.

### Option B: Manual Testing

Open a **new** command prompt while the server is running:

#### Health Check
```cmd
curl http://localhost:8080/health
```

#### Classify an IP
```cmd
curl http://localhost:8080/api/v1/classify/ip/112.118.109.197
```

#### Test in Browser
Open: http://localhost:8080/health

---

## Project Structure

```
prod/
├── build/                          # Compiled binaries
│   └── onefirewall-classifier.exe
├── cmd/
│   └── server/
│       └── main.go                 # Application entry point
├── internal/                       # Core application code
│   ├── api/                        # HTTP handlers
│   ├── classifier/                 # Classification logic
│   ├── data/                       # Data management
│   └── models/                     # Data models
├── data/                           # Data files
│   ├── sample-ip2asn-v4.tsv       # Sample IP database
│   └── ip2asn-v4.tsv              # Full IP database (download)
├── .env                            # Environment configuration
├── setup.bat                       # Setup script
├── run.bat                         # Run script
├── download-ipdb.bat              # Download database script
├── test-api.bat                   # API testing script
├── go.mod                         # Go dependencies
└── README.md                      # Full documentation
```

---

## Configuration Options

### Environment Variables (.env file)

```env
# Required: OneFirewall API key
ONEFIREWALL_API_KEY=your-key-here

# Optional: OneFirewall base URL
ONEFIREWALL_BASE_URL=https://app.onefirewall.com

# Optional: Server configuration
PORT=8080
LOG_LEVEL=info

# Optional: Data paths
IPDB_PATH=data/ip2asn-v4.tsv
MALICIOUS_IPS_PATH=data/malicious-ips.json
```

### Command-Line Options

```cmd
build\onefirewall-classifier.exe -h
```

Available flags:
- `-port` - Server port (default: 8080)
- `-ipdb` - Path to IP database TSV file
- `-malicious` - Path to malicious IPs JSON file
- `-api-key` - OneFirewall API key
- `-onefirewall-url` - OneFirewall base URL
- `-log-level` - Log level (debug, info, warn, error)

Example:
```cmd
build\onefirewall-classifier.exe ^
  -port 9090 ^
  -ipdb data\ip2asn-v4.tsv ^
  -log-level debug
```

---

## PowerShell Alternative

If you prefer PowerShell:

### Build
```powershell
go build -o build\onefirewall-classifier.exe cmd\server\main.go
```

### Set Environment Variable
```powershell
$env:ONEFIREWALL_API_KEY = "your-key-here"
```

### Run
```powershell
.\build\onefirewall-classifier.exe
```

### Test
```powershell
# Health check
Invoke-RestMethod -Uri http://localhost:8080/health | ConvertTo-Json

# Classify IP
Invoke-RestMethod -Uri http://localhost:8080/api/v1/classify/ip/112.118.109.197 | ConvertTo-Json
```

---

## Troubleshooting

### "go: command not found"

**Solution:** Install Go from https://go.dev/dl/

### "curl: command not found"

**Solution:** Either:
- Install curl from https://curl.se/windows/
- Use PowerShell's `Invoke-RestMethod`
- Test in browser: http://localhost:8080/health

### Port Already in Use

**Solution:** Use a different port:
```cmd
build\onefirewall-classifier.exe -port 9090
```

Or find what's using port 8080:
```cmd
netstat -ano | findstr :8080
```

### Build Fails

**Solution:** Clean and rebuild:
```cmd
rmdir /s /q build
go clean
go mod tidy
setup.bat
```

### .env File Not Loading

**Solution:** Ensure:
1. File is named exactly `.env` (no `.txt` extension)
2. Located in the project root directory
3. Contains valid `KEY=VALUE` format

### API Key Not Working

**Solution:** Verify:
1. API key is correct in `.env` file
2. No extra spaces or quotes around the key
3. File is saved after editing

---

## Next Steps

1. ✅ **Run setup.bat** to build the application
2. ✅ **Edit .env** with your API key
3. ✅ **Run run.bat** to start the server
4. ✅ **Run test-api.bat** to test the API
5. 📚 **Read README.md** for full API documentation

---

## Support

- 📖 Full documentation: `README.md`
- 📝 Changelog: `CHANGELOG.md`
- 🐛 Issues: Report on GitHub
- 💬 Questions: Check the documentation

---

## Production Deployment

For production use:

1. **Download full IP database:**
   ```cmd
   download-ipdb.bat
   ```

2. **Set environment variables permanently:**
   - Open System Properties → Environment Variables
   - Add `ONEFIREWALL_API_KEY` with your key

3. **Run as Windows Service:**
   - Use NSSM (Non-Sucking Service Manager)
   - Or Windows Task Scheduler

4. **Configure firewall:**
   ```cmd
   netsh advfirewall firewall add rule name="OneFirewall Classifier" dir=in action=allow protocol=TCP localport=8080
   ```

---

Enjoy using OneFirewall Classifier! 🎉
