@echo off
REM Download IP-to-ASN Database
REM This script downloads the full IP database from iptoasn.com

echo.
echo ========================================
echo Download IP-to-ASN Database
echo ========================================
echo.

REM Check if curl is available
where curl >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] curl is not available!
    echo.
    echo Please download the database manually:
    echo 1. Visit: https://iptoasn.com/
    echo 2. Download: ip2asn-v4.tsv.gz
    echo 3. Extract the .gz file
    echo 4. Move ip2asn-v4.tsv to the data\ folder
    echo.
    echo Or install curl and run this script again.
    echo.
    pause
    exit /b 1
)

echo Downloading IP-to-ASN database...
echo This may take a few minutes depending on your connection...
echo.

REM Create data directory if it doesn't exist
if not exist data mkdir data

REM Download the database
curl -L -o data\ip2asn-v4.tsv.gz https://iptoasn.com/data/ip2asn-v4.tsv.gz
if %ERRORLEVEL% NEQ 0 (
    echo.
    echo [ERROR] Download failed!
    echo.
    echo Please download manually from:
    echo https://iptoasn.com/data/ip2asn-v4.tsv.gz
    echo.
    echo Then extract and place in data\ folder
    echo.
    pause
    exit /b 1
)

echo [OK] Download complete!
echo.

REM Check if tar/gzip is available for extraction
where tar >nul 2>nul
if %ERRORLEVEL% EQU 0 (
    echo Extracting database...
    tar -xzf data\ip2asn-v4.tsv.gz -C data
    if %ERRORLEVEL% EQU 0 (
        echo [OK] Extraction complete!
        echo.
        del data\ip2asn-v4.tsv.gz
        echo [OK] Cleaned up temporary files
    ) else (
        echo [WARNING] Extraction failed
        echo Please extract data\ip2asn-v4.tsv.gz manually
    )
) else (
    echo.
    echo [INFO] Please extract data\ip2asn-v4.tsv.gz manually
    echo You can use Windows built-in extraction or 7-Zip
)

echo.
echo ========================================
echo Done!
echo ========================================
echo.

if exist data\ip2asn-v4.tsv (
    echo [SUCCESS] Database is ready!
    echo Location: data\ip2asn-v4.tsv
    echo.
    for %%A in (data\ip2asn-v4.tsv) do echo Size: %%~zA bytes
    echo.
    echo You can now run: run.bat
) else (
    echo [INFO] After extraction, ensure the file is at:
    echo data\ip2asn-v4.tsv
)
echo.
pause
