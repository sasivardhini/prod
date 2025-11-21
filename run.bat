@echo off
REM OneFirewall Classifier - Run Script
REM This script loads environment variables and runs the classifier

echo.
echo ========================================
echo OneFirewall Classifier v2.0.0
echo ========================================
echo.

REM Check if build exists
if not exist build\onefirewall-classifier.exe (
    echo [ERROR] Application not built!
    echo Please run: setup.bat
    echo.
    pause
    exit /b 1
)

REM Load environment variables from .env file
if exist .env (
    echo Loading configuration from .env file...
    for /f "usebackq tokens=1,* delims==" %%a in (".env") do (
        set "line=%%a"
        REM Skip comments and empty lines
        if not "!line:~0,1!"=="#" if not "!line!"=="" (
            set "%%a=%%b"
        )
    )
    echo [OK] Configuration loaded
    echo.
) else (
    echo [WARNING] .env file not found, using default settings
    echo.
)

REM Display configuration
echo Starting OneFirewall Classifier...
echo.
if defined ONEFIREWALL_API_KEY (
    echo [OK] OneFirewall API key is configured
) else (
    echo [WARNING] No API key found - will use sample data only
)
echo.
echo Server will be available at: http://localhost:8080
echo Press Ctrl+C to stop the server
echo.
echo ========================================
echo.

REM Determine which database to use
set DB_ARG=
if exist data\ip2asn-v4.tsv (
    set DB_ARG=-ipdb data\ip2asn-v4.tsv
    echo Using: Full IP database
) else if exist data\sample-ip2asn-v4.tsv (
    set DB_ARG=-ipdb data\sample-ip2asn-v4.tsv
    echo Using: Sample IP database
) else (
    echo Using: Built-in sample data
)
echo.

REM Run the application
build\onefirewall-classifier.exe %DB_ARG% -log-level info

echo.
echo Server stopped.
pause
