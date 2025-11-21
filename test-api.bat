@echo off
REM OneFirewall Classifier - API Test Script
REM This script tests the API endpoints

echo.
echo ========================================
echo OneFirewall Classifier - API Tests
echo ========================================
echo.

REM Check if curl is available
where curl >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] curl is not available!
    echo.
    echo Please install curl or test manually:
    echo Open browser to: http://localhost:8080/health
    echo.
    pause
    exit /b 1
)

REM Check if server is running
curl -s http://localhost:8080/health >nul 2>nul
if %ERRORLEVEL% NEQ 0 (
    echo [ERROR] Server is not running!
    echo.
    echo Please start the server first:
    echo   run.bat
    echo.
    pause
    exit /b 1
)

echo Server is running!
echo.
echo ========================================
echo Test 1: Health Check
echo ========================================
curl -s http://localhost:8080/health
echo.
echo.

echo ========================================
echo Test 2: Classify IP (Malicious Example)
echo ========================================
echo IP: 112.118.109.197
curl -s http://localhost:8080/api/v1/classify/ip/112.118.109.197
echo.
echo.

echo ========================================
echo Test 3: Network Address (Should be NONE)
echo ========================================
echo IP: 112.118.109.0
curl -s http://localhost:8080/api/v1/classify/ip/112.118.109.0
echo.
echo.

echo ========================================
echo Test 4: Broadcast Address (Should be NONE)
echo ========================================
echo IP: 112.118.109.255
curl -s http://localhost:8080/api/v1/classify/ip/112.118.109.255
echo.
echo.

echo ========================================
echo Test 5: Private IP (Should be NONE)
echo ========================================
echo IP: 192.168.1.1
curl -s http://localhost:8080/api/v1/classify/ip/192.168.1.1
echo.
echo.

echo ========================================
echo Test 6: Classify ASN
echo ========================================
echo ASN: 9808
curl -s http://localhost:8080/api/v1/classify/asn/9808
echo.
echo.

echo ========================================
echo Test 7: Classify Country
echo ========================================
echo Country: CN
curl -s http://localhost:8080/api/v1/classify/country/CN
echo.
echo.

echo ========================================
echo Test 8: Batch Classification
echo ========================================
curl -s -X POST http://localhost:8080/api/v1/classify/batch ^
  -H "Content-Type: application/json" ^
  -d "[{\"ip\":\"112.118.109.197\"},{\"asn\":9808},{\"country\":\"CN\"}]"
echo.
echo.

echo ========================================
echo All Tests Complete!
echo ========================================
echo.
pause
