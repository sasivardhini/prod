# ⚠️ Database Download Required

## The IP Database File

The `data/ip2asn-v4.tsv` file included in this repository is a **starter database** with major IP ranges for testing.

For **production use**, you need to download the **complete IP-to-ASN database** from iptoasn.com:

### Download Instructions:

**PowerShell (Windows):**
```powershell
cd data
Invoke-WebRequest -Uri "https://iptoasn.com/data/ip2asn-v4.tsv.gz" -OutFile "ip2asn-v4.tsv.gz"
tar -xzf ip2asn-v4.tsv.gz
Remove-Item ip2asn-v4.tsv.gz
```

**Linux/Mac:**
```bash
cd data
wget https://iptoasn.com/data/ip2asn-v4.tsv.gz
gunzip ip2asn-v4.tsv.gz
```

**Manual Download:**
1. Visit: https://iptoasn.com/
2. Download: `ip2asn-v4.tsv.gz` (~25 MB)
3. Extract and replace `data/ip2asn-v4.tsv`

### Database Stats:

**Included (Starter):**
- ~50 IP ranges
- Major providers: Google, Amazon, Cloudflare, Facebook, Microsoft
- Good for testing and development

**Full Database (Download):**
- ~500,000+ IP ranges
- Complete IPv4 coverage
- ~100 MB uncompressed
- Production-ready

### Running with the Database:

```bash
# With .env file
go build -o build/onefirewall-classifier.exe cmd/server/main.go
build/onefirewall-classifier.exe -ipdb data/ip2asn-v4.tsv
```

The application works with both the starter database and the full database!
