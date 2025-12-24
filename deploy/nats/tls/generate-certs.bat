@echo off
REM TLS Certificate Generation Script for LightLink
REM This script generates CA, server, and client certificates for NATS TLS authentication

setlocal enabledelayedexpansion

echo ========================================
echo LightLink TLS Certificate Generator
echo ========================================
echo.

REM Get script directory
set "SCRIPT_DIR=%~dp0"
cd /d "%SCRIPT_DIR%"

REM Check if OpenSSL is available
where openssl >nul 2>&1
if errorlevel 1 (
    echo ERROR: OpenSSL not found in PATH
    echo Please install OpenSSL or add it to PATH
    echo Download from: https://slproweb.com/products/Win32OpenSSL.html
    pause
    exit /b 1
)

echo OpenSSL found, starting certificate generation...
echo.

REM Step 1: Generate CA Certificate
echo [1/6] Generating CA certificate...
openssl genrsa -out ca.key 2048
if errorlevel 1 goto error
openssl req -new -x509 -days 10950 -key ca.key -out ca.crt -subj "/CN=LightLink CA"
if errorlevel 1 goto error
echo CA certificate generated successfully
echo.

REM Step 2: Generate Server Certificate
echo [2/6] Generating server certificate...
openssl genrsa -out server.key 2048
if errorlevel 1 goto error
openssl req -new -key server.key -out server.csr -subj "/CN=nats-server"
if errorlevel 1 goto error
openssl x509 -req -days 10950 -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt
if errorlevel 1 goto error
echo Server certificate generated successfully
echo.

REM Step 3: Generate Client Certificate - demo-service
echo [3/6] Generating client certificate for demo-service...
openssl genrsa -out demo-service.key 2048
if errorlevel 1 goto error
openssl req -new -key demo-service.key -out demo-service.csr -subj "/CN=demo-service"
if errorlevel 1 goto error
openssl x509 -req -days 10950 -in demo-service.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out demo-service.crt
if errorlevel 1 goto error
echo demo-service certificate generated successfully
echo.

REM Step 4: Generate Client Certificate - test-service
echo [4/6] Generating client certificate for test-service...
openssl genrsa -out test-service.key 2048
if errorlevel 1 goto error
openssl req -new -key test-service.key -out test-service.csr -subj "/CN=test-service"
if errorlevel 1 goto error
openssl x509 -req -days 10950 -in test-service.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out test-service.crt
if errorlevel 1 goto error
echo test-service certificate generated successfully
echo.

REM Step 5: Generate Client Certificate - client-app
echo [5/6] Generating client certificate for client-app...
openssl genrsa -out client-app.key 2048
if errorlevel 1 goto error
openssl req -new -key client-app.key -out client-app.csr -subj "/CN=client-app"
if errorlevel 1 goto error
openssl x509 -req -days 10950 -in client-app.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out client-app.crt
if errorlevel 1 goto error
echo client-app certificate generated successfully
echo.

REM Step 6: Cleanup temporary files
echo [6/6] Cleaning up temporary files...
del *.csr 2>nul
del ca.srl 2>nul
echo Cleanup completed
echo.

REM Create users directory for NATS configuration
if not exist "users" mkdir users

REM Copy certificates to users directory for reference
copy /Y ca.crt users\ >nul 2>&1
copy /Y demo-service.crt users\ >nul 2>&1
copy /Y test-service.crt users\ >nul 2>&1
copy /Y client-app.crt users\ >nul 2>&1

echo ========================================
echo Certificate generation completed!
echo ========================================
echo.
echo Generated files:
echo - ca.key, ca.crt (CA certificate)
echo - server.key, server.crt (Server certificate)
echo - demo-service.key, demo-service.crt (Client certificate)
echo - test-service.key, test-service.crt (Client certificate)
echo - client-app.key, client-app.crt (Client certificate)
echo.
echo Certificate users (CN):
echo - demo-service
echo - test-service
echo - client-app
echo.
echo NOTE: Keep .key files secure and private!
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
