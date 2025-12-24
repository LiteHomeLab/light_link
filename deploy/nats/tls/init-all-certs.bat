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
REM Step 5: Create Client Distribution Packages
REM ========================================

echo [Step 5/6] Creating Client Distribution Packages...

if not exist "clients" mkdir clients

REM Create package for demo-service
echo   - Creating package for demo-service...
if not exist "clients\demo-service" mkdir clients\demo-service
copy /Y ca.crt clients\demo-service\ >nul
copy /Y demo-service.crt clients\demo-service\ >nul
copy /Y demo-service.key clients\demo-service\ >nul

REM Create package for test-service
echo   - Creating package for test-service...
if not exist "clients\test-service" mkdir clients\test-service
copy /Y ca.crt clients\test-service\ >nul
copy /Y test-service.crt clients\test-service\ >nul
copy /Y test-service.key clients\test-service\ >nul

REM Create package for client-app
echo   - Creating package for client-app...
if not exist "clients\client-app" mkdir clients\client-app
copy /Y ca.crt clients\client-app\ >nul
copy /Y client-app.crt clients\client-app\ >nul
copy /Y client-app.key clients\client-app\ >nul

echo Client packages created in clients/ directory
echo.

REM ========================================
REM Step 6: Generate Documentation
REM ========================================

echo [Step 6/6] Generating documentation...

REM Generate README for each client package
for %%s in (demo-service test-service client-app) do (
    echo # TLS Certificate Package for %%s > clients\%%s\README.md
    echo. >> clients\%%s\README.md
    echo ## Deployment >> clients\%%s\README.md
    echo. >> clients\%%s\README.md
    echo Copy all files in this folder to your service directory: >> clients\%%s\README.md
    echo. >> clients\%%s\README.md
    echo ```batch >> clients\%%s\README.md
    echo mkdir tls >> clients\%%s\README.md
    echo copy *.* tls\ >> clients\%%s\README.md
    echo ``` >> clients\%%s\README.md
    echo. >> clients\%%s\README.md
    echo ## Configuration (Default Paths) >> clients\%%s\README.md
    echo. >> clients\%%s\README.md
    echo After copying to tls/ directory, use these default paths: >> clients\%%s\README.md
    echo. >> clients\%%s\README.md
    echo ### Go SDK >> clients\%%s\README.md
    echo ```go >> clients\%%s\README.md
    echo tlsConfig := ^&client.TLSConfig{ >> clients\%%s\README.md
    echo     CaFile:     "tls/ca.crt", >> clients\%%s\README.md
    echo     CertFile:   "tls/%%s.crt", >> clients\%%s\README.md
    echo     KeyFile:    "tls/%%s.key", >> clients\%%s\README.md
    echo     ServerName: "nats-server", >> clients\%%s\README.md
    echo } >> clients\%%s\README.md
    echo ``` >> clients\%%s\README.md
    echo. >> clients\%%s\README.md
    echo ### Python SDK >> clients\%%s\README.md
    echo ```python >> clients\%%s\README.md
    echo tls_config = TLSConfig( >> clients\%%s\README.md
    echo     ca_file="tls/ca.crt", >> clients\%%s\README.md
    echo     cert_file="tls/%%s.crt", >> clients\%%s\README.md
    echo     key_file="tls/%%s.key", >> clients\%%s\README.md
    echo     server_name="nats-server" >> clients\%%s\README.md
    echo ) >> clients\%%s\README.md
    echo ``` >> clients\%%s\README.md
)

REM Create certificate manifest
echo # LightLink Certificate Manifest > cert-manifest.txt
echo # Auto-generated on %date% %time% >> cert-manifest.txt
echo. >> cert-manifest.txt
echo ## NATS Server (server-side only) >> cert-manifest.txt
echo nats-server: server.crt, server.key >> cert-manifest.txt
echo. >> cert-manifest.txt
echo ## Client Distribution Packages >> cert-manifest.txt
echo demo-service: clients/demo-service/ >> cert-manifest.txt
echo test-service: clients/test-service/ >> cert-manifest.txt
echo client-app: clients/client-app/ >> cert-manifest.txt

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
echo Client Distribution Packages (ready to deploy):
echo   - clients\demo-service\
echo   - clients\test-service\
echo   - clients\client-app\
echo.
echo Each package contains:
echo   - ca.crt
echo   - {service}.crt
echo   - {service}.key
echo   - README.md
echo.
echo ========================================
echo Deployment Instructions
echo ========================================
echo.
echo 1. NATS Server Configuration:
echo    Use: server.crt, server.key, ca.crt
echo.
echo 2. Client Services:
echo    Copy the entire client package folder to your service:
echo.
echo    For example, to deploy demo-service:
echo      xcopy /E /I clients\demo-service your-service\tls\
echo.
echo 3. Then use default tls/ paths in your service config:
echo    ca: tls/ca.crt
echo    cert: tls/demo-service.crt
echo    key: tls/demo-service.key
echo    server_name: nats-server
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
