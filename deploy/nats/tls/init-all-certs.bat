@echo off
REM LightLink Complete Certificate Initialization Script
REM This script generates CA, NATS server certificates, and client packages

setlocal enabledelayedexpansion

echo ========================================
echo LightLink Certificate Initialization
echo ========================================
echo.
echo This will generate:
echo   1. CA Certificate (ca.crt, ca.key)
echo   2. NATS Server Certificate (server.crt, server.key)
echo   3. Client Certificate Packages
echo.

REM Get script directory
set "SCRIPT_DIR=%~dp0"
cd /d "%SCRIPT_DIR%"

REM ========================================
REM Step 1: Find OpenSSL
REM ========================================

echo [Step 1/6] Locating OpenSSL...
set "OPENSSL_CMD="

REM Check PATH
where openssl >nul 2>&1
if not errorlevel 1 (
    set "OPENSSL_CMD=openssl"
    goto openssl_found
)

REM Check common paths
if exist "C:\Program Files\Git\mingw64\bin\openssl.exe" (
    set "OPENSSL_CMD=C:\Program Files\Git\mingw64\bin\openssl.exe"
    goto openssl_found
)
if exist "C:\Program Files\Git\usr\bin\openssl.exe" (
    set "OPENSSL_CMD=C:\Program Files\Git\usr\bin\openssl.exe"
    goto openssl_found
)
if exist "C:\Program Files\OpenSSL-Win64\bin\openssl.exe" (
    set "OPENSSL_CMD=C:\Program Files\OpenSSL-Win64\bin\openssl.exe"
    goto openssl_found
)
if exist "C:\Program Files (x86)\Git\mingw64\bin\openssl.exe" (
    set "OPENSSL_CMD=C:\Program Files (x86)\Git\mingw64\bin\openssl.exe"
    goto openssl_found
)
if exist "C:\Program Files (x86)\Git\usr\bin\openssl.exe" (
    set "OPENSSL_CMD=C:\Program Files (x86)\Git\usr\bin\openssl.exe"
    goto openssl_found
)

echo ERROR: OpenSSL not found
echo Please install Git for Windows: https://git-scm.com/download/win
pause
exit /b 1

:openssl_found
echo Found OpenSSL: %OPENSSL_CMD%
echo.

REM ========================================
REM Step 2: Generate CA Certificate
REM ========================================

echo [Step 2/6] Generating CA Certificate...
"%OPENSSL_CMD%" genrsa -out ca.key 2048
if errorlevel 1 goto error
"%OPENSSL_CMD%" req -new -x509 -days 10950 -key ca.key -out ca.crt -subj "/CN=LightLink CA"
if errorlevel 1 goto error
echo CA Certificate created: ca.crt, ca.key
echo.

REM ========================================
REM Step 3: Generate NATS Server Certificate
REM ========================================

echo [Step 3/6] Generating NATS Server Certificate...
"%OPENSSL_CMD%" genrsa -out server.key 2048
if errorlevel 1 goto error
"%OPENSSL_CMD%" req -new -key server.key -out server.csr -subj "/CN=nats-server"
if errorlevel 1 goto error
"%OPENSSL_CMD%" x509 -req -days 10950 -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt
if errorlevel 1 goto error
echo NATS Server Certificate created: server.crt, server.key
echo.

REM ========================================
REM Step 4: Generate Default Client Certificates
REM ========================================

echo [Step 4/6] Generating Default Client Certificates...

REM demo-service
echo   - Generating demo-service certificate...
"%OPENSSL_CMD%" genrsa -out demo-service.key 2048
if errorlevel 1 goto error
"%OPENSSL_CMD%" req -new -key demo-service.key -out demo-service.csr -subj "/CN=demo-service"
if errorlevel 1 goto error
"%OPENSSL_CMD%" x509 -req -days 10950 -in demo-service.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out demo-service.crt
if errorlevel 1 goto error

REM test-service
echo   - Generating test-service certificate...
"%OPENSSL_CMD%" genrsa -out test-service.key 2048
if errorlevel 1 goto error
"%OPENSSL_CMD%" req -new -key test-service.key -out test-service.csr -subj "/CN=test-service"
if errorlevel 1 goto error
"%OPENSSL_CMD%" x509 -req -days 10950 -in test-service.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out test-service.crt
if errorlevel 1 goto error

REM client-app
echo   - Generating client-app certificate...
"%OPENSSL_CMD%" genrsa -out client-app.key 2048
if errorlevel 1 goto error
"%OPENSSL_CMD%" req -new -key client-app.key -out client-app.csr -subj "/CN=client-app"
if errorlevel 1 goto error
"%OPENSSL_CMD%" x509 -req -days 10950 -in client-app.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out client-app.crt
if errorlevel 1 goto error

echo Default client certificates created
echo.

REM ========================================
REM Step 5: Create Client Distribution Package
REM ========================================

echo [Step 5/6] Creating Client Distribution Package...

if not exist "client" mkdir client

REM Copy CA certificate
echo   - Copying CA certificate...
copy /Y ca.crt client\ >nul

REM Copy default client certificates
echo   - Copying client certificates...
copy /Y demo-service.crt client\ >nul
copy /Y demo-service.key client\ >nul
copy /Y test-service.crt client\ >nul
copy /Y test-service.key client\ >nul
copy /Y client-app.crt client\ >nul
copy /Y client-app.key client\ >nul

echo Client package created in client/ directory
echo.

REM ========================================
REM Step 6: Generate Documentation
REM ========================================

echo [Step 6/6] Generating documentation...

REM Generate README for client package
echo # LightLink Client Certificate Package > client\README.md
echo. >> client\README.md
echo Generated: %date% %time% >> client\README.md
echo. >> client\README.md
echo ## Files >> client\README.md
echo. >> client\README.md
echo - ca.crt - CA Root Certificate (trusted by all clients) >> client\README.md
echo. >> client\README.md
echo Default client certificates: >> client\README.md
echo - demo-service.crt / demo-service.key >> client\README.md
echo - test-service.crt / test-service.key >> client\README.md
echo - client-app.crt / client-app.key >> client\README.md
echo. >> client\README.md
echo ## Deployment >> client\README.md
echo. >> client\README.md
echo 1. Copy this entire client/ folder to your service directory >> client\README.md
echo 2. Use the certificate files for your service >> client\README.md
echo. >> client\README.md
echo ## Configuration Examples >> client\README.md
echo. >> client\README.md
echo ### For demo-service (Go SDK) >> client\README.md
echo ```go >> client\README.md
echo tlsConfig := ^&client.TLSConfig{ >> client\README.md
echo     CaFile:     "client/ca.crt", >> client\README.md
echo     CertFile:   "client/demo-service.crt", >> client\README.md
echo     KeyFile:    "client/demo-service.key", >> client\README.md
echo     ServerName: "nats-server", >> client\README.md
echo } >> client\README.md
echo ``` >> client\README.md
echo. >> client\README.md
echo ### For test-service (Python SDK) >> client\README.md
echo ```python >> client\README.md
echo tls_config = TLSConfig( >> client\README.md
echo     ca_file="client/ca.crt", >> client\README.md
echo     cert_file="client/test-service.crt", >> client\README.md
echo     key_file="client/test-service.key", >> client\README.md
echo     server_name="nats-server" >> client\README.md
echo ) >> client\README.md
echo ``` >> client\README.md
echo. >> client\README.md
echo ## Security Notice >> client\README.md
echo - Keep .key files secure and private! >> client\README.md
echo - Do not commit .key files to version control >> client\README.md

echo Documentation created
echo.

REM ========================================
REM Cleanup
REM ========================================

echo Cleaning up temporary files...
del *.csr 2>nul
del ca.srl 2>nul
echo.

REM ========================================
REM Summary
REM ========================================

echo ========================================
echo Initialization Complete!
echo ========================================
echo.
echo Generated Certificates:
echo   - ca.key, ca.crt (CA Certificate)
echo   - server.key, server.crt (NATS Server)
echo   - demo-service.key, demo-service.crt
echo   - test-service.key, test-service.crt
echo   - client-app.key, client-app.crt
echo.
echo Client Package (ready to deploy):
echo   - client\ (directory)
echo     - ca.crt
echo     - demo-service.crt, demo-service.key
echo     - test-service.crt, test-service.key
echo     - client-app.crt, client-app.key
echo     - README.md
echo.
echo ========================================
echo Deployment Instructions
echo ========================================
echo.
echo 1. NATS Server Configuration:
echo    Use: server.crt, server.key, ca.crt
echo.
echo 2. Client Services:
echo    Copy the entire client/ folder to your service directory:
echo.
echo    xcopy /E /I client your-service-directory\
echo.
echo    Then configure your service to use the certificate files:
echo      ca: client/ca.crt
echo      cert: client/demo-service.crt (or your service)
echo      key: client/demo-service.key (or your service)
echo      server_name: nats-server
echo.
echo To generate additional client certificates:
echo    generate-service-cert.bat your-service-name
echo.

goto end

:error
echo.
echo ========================================
echo ERROR: Certificate generation failed
echo ========================================
echo.
pause
exit /b 1

:end
pause
