# 🚀 Quick Setup - OneFirewall Classifier

## Step 1: Download the IP Database

The real IP-to-ASN database must be downloaded from iptoasn.com:

### On Windows:

Open PowerShell and run:
```powershell
# Navigate to your project directory
cd C:\Users\sasiv\Downloads\prod-claude-onefirewall-classification-model-01DqmPZK73yrJnSEDUgd53SU\prod-claude-onefirewall-classification-model-01DqmPZK73yrJnSEDUgd53SU

# Create data directory if it doesn't exist
if (!(Test-Path "data")) { New-Item -ItemType Directory -Path "data" }

# Download the database (~25 MB compressed)
Invoke-WebRequest -Uri "https://iptoasn.com/data/ip2asn-v4.tsv.gz" -OutFile "data\ip2asn-v4.tsv.gz"

# Extract the database
tar -xzf data\ip2asn-v4.tsv.gz -C data

# Remove the compressed file
Remove-Item data\ip2asn-v4.tsv.gz

# Verify the database
Get-Item data\ip2asn-v4.tsv
```

**Or download manually:**
1. Visit: https://iptoasn.com/
2. Click "Download" → Download `ip2asn-v4.tsv.gz`
3. Extract the `.gz` file (use 7-Zip or built-in Windows extraction)
4. Move `ip2asn-v4.tsv` to the `data\` folder in your project

---

## Step 2: Verify Your .env File

Your `.env` file is already configured with your OneFirewall API key:

```env
ONEFIREWALL_API_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ0cyI6MTc1NjYwMTIwNywiZ3VpZCI6Ik9GQS1HVUlELTU3MjQtMjAwNy0wMjMyIiwiaWF0IjoxNzU2NjAxMjA3fQ.tPdQGERKfN7mEALoEfn_yk9RubsxZj7pFIXmmT8-F7I
```

✅ This file is already set up correctly - no changes needed!

---

## Step 3: Build the Application

Open Command Prompt:
```cmd
cd C:\Users\sasiv\Downloads\prod-claude-onefirewall-classification-model-01DqmPZK73yrJnSEDUgd53SU\prod-claude-onefirewall-classification-model-01DqmPZK73yrJnSEDUgd53SU

go build -o build\onefirewall-classifier.exe cmd\server\main.go
```

---

## Step 4: Run the Application

### Option A: With .env file (Recommended)

**Windows Command Prompt:**
```cmd
REM Load environment variable from .env
set ONEFIREWALL_API_KEY=eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ0cyI6MTc1NjYwMTIwNywiZ3VpZCI6Ik9GQS1HVUlELTU3MjQtMjAwNy0wMjMyIiwiaWF0IjoxNzU2NjAxMjA3fQ.tPdQGERKfN7mEALoEfn_yk9RubsxZj7pFIXmmT8-F7I

REM Run with the real database
build\onefirewall-classifier.exe -ipdb data\ip2asn-v4.tsv
```

**Windows PowerShell:**
```powershell
# Load environment variable
$env:ONEFIREWALL_API_KEY = "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ0cyI6MTc1NjYwMTIwNywiZ3VpZCI6Ik9GQS1HVUlELTU3MjQtMjAwNy0wMjMyIiwiaWF0IjoxNzU2NjAxMjA3fQ.tPdQGERKfN7mEALoEfn_yk9RubsxZj7pFIXmmT8-F7I"

# Run with the real database
.\build\onefirewall-classifier.exe -ipdb data\ip2asn-v4.tsv
```

### Option B: Using command-line flag

```cmd
build\onefirewall-classifier.exe -ipdb data\ip2asn-v4.tsv -api-key eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJ0cyI6MTc1NjYwMTIwNywiZ3VpZCI6Ik9GQS1HVUlELTU3MjQtMjAwNy0wMjMyIiwiaWF0IjoxNzU2NjAxMjA3fQ.tPdQGERKfN7mEALoEfn_yk9RubsxZj7pFIXmmT8-F7I
```

---

## Step 5: Test the API

Server will be running at: **http://localhost:8080**

### Test in Browser:
- http://localhost:8080/health
- http://localhost:8080/api/v1/classify/ip/112.118.109.197

### Test with curl:
```cmd
curl http://localhost:8080/health
curl http://localhost:8080/api/v1/classify/ip/112.118.109.197
curl http://localhost:8080/api/v1/classify/asn/9808
curl http://localhost:8080/api/v1/classify/country/CN
```

### Test with PowerShell:
```powershell
Invoke-RestMethod -Uri http://localhost:8080/health | ConvertTo-Json
Invoke-RestMethod -Uri http://localhost:8080/api/v1/classify/ip/112.118.109.197 | ConvertTo-Json
```

---

## Expected Output

When you start the server, you should see:

```
INFO[0000] Starting OneFirewall Classifier Service...
INFO[0000] OneFirewall API key configured - real-time API integration enabled
INFO[0000] Loading IP database from: data\ip2asn-v4.tsv
INFO[0000] Loaded 123456 IP ranges
INFO[0000] No malicious IPs file specified. Will fetch data from OneFirewall API on-demand.
INFO[0000] Server starting on :8080
INFO[0000] API endpoints:
INFO[0000]   GET  /health
INFO[0000]   GET  /api/v1/health
INFO[0000]   GET  /api/v1/classify/ip/{ip}
INFO[0000]   GET  /api/v1/classify/asn/{asn}
INFO[0000]   GET  /api/v1/classify/country/{country}
INFO[0000]   POST /api/v1/classify/batch
```

---

## File Structure

After setup, your directory should have:

```
prod/
├── .env                            ✅ Your API key (already set up)
├── data/
│   └── ip2asn-v4.tsv              ✅ Real database (~100 MB, 500k+ ranges)
├── build/
│   └── onefirewall-classifier.exe ✅ Compiled application
├── cmd/
├── internal/
├── go.mod
└── README.md
```

---

## Troubleshooting

### Database file not found
Make sure `data\ip2asn-v4.tsv` exists:
```cmd
dir data\ip2asn-v4.tsv
```

### API key not loading
Verify the environment variable is set:
```cmd
echo %ONEFIREWALL_API_KEY%
```

### Port already in use
Use a different port:
```cmd
build\onefirewall-classifier.exe -ipdb data\ip2asn-v4.tsv -port 9090
```

---

## Summary

✅ **.env file** - Already configured with your API key
✅ **Real database** - Download from https://iptoasn.com/data/ip2asn-v4.tsv.gz
✅ **Build** - `go build -o build\onefirewall-classifier.exe cmd\server\main.go`
✅ **Run** - `build\onefirewall-classifier.exe -ipdb data\ip2asn-v4.tsv`
✅ **Test** - http://localhost:8080/health

That's it! 🚀
