@echo off
REM OneFirewall Classifier - Windows Setup Script
REM This script sets up everything you need to run the classifier

echo.
echo ========================================
echo OneFirewall Classifier - Setup
echo ========================================
echo.

REM Check if Go is installed
where go >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Go is not installed!
    echo Please download and install Go from: https://go.dev/dl/
    echo.
    pause
    exit /b 1
)

echo [OK] Go is installed
go version
echo.

REM Create directories
echo Creating directories...
if not exist build mkdir build
if not exist data mkdir data
echo [OK] Directories created
echo.

REM Check if .env file exists
if not exist .env (
    echo [WARNING] .env file not found!
    echo Creating default .env file...
    (
        echo # OneFirewall Classifier Configuration
        echo ONEFIREWALL_API_KEY=your-api-key-here
    ) > .env
    echo [INFO] Please edit .env file and add your OneFirewall API key
    echo.
)

REM Check for IP database
if not exist data\ip2asn-v4.tsv (
    if not exist data\sample-ip2asn-v4.tsv (
        echo [WARNING] No IP database found!
        echo.
        echo You can either:
        echo 1. Download the full database (recommended for production)
        echo    Run: download-ipdb.bat
        echo.
        echo 2. Use sample data (good for testing)
        echo    The app will use built-in sample data
        echo.
    ) else (
        echo [OK] Sample IP database found
    )
) else (
    echo [OK] Full IP database found
)
echo.

REM Download Go dependencies
echo Downloading Go dependencies...
go mod download
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Failed to download dependencies
    pause
    exit /b 1
)
echo [OK] Dependencies downloaded
echo.

REM Build the application
echo Building OneFirewall Classifier...
go build -v -o build\onefirewall-classifier.exe cmd\server\main.go
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Build failed!
    pause
    exit /b 1
)
echo.
echo [SUCCESS] Build completed successfully!
echo.

echo ========================================
echo Setup Complete!
echo ========================================
echo.
echo To run the classifier:
echo   1. Edit .env file with your OneFirewall API key
echo   2. Run: run.bat
echo.
echo For more options, see README.md
echo.
pause
