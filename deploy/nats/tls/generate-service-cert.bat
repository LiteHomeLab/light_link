@echo off
REM Service Certificate Generator for LightLink
REM Usage: generate-service-cert.bat <service-name>

setlocal enabledelayedexpansion

if "%~1"=="" (
    echo Usage: generate-service-cert.bat ^<service-name^>
    echo Example: generate-service-cert.bat my-service
    exit /b 1
)

set "SERVICE_NAME=%~1"
set "SCRIPT_DIR=%~dp0"
cd /d "%SCRIPT_DIR%"

echo ========================================
echo LightLink Service Certificate Generator
echo ========================================
echo.
echo Service Name: %SERVICE_NAME%
echo.

REM Check OpenSSL
where openssl >nul 2>&1
if errorlevel 1 (
    echo ERROR: OpenSSL not found in PATH
    echo Please install OpenSSL or add it to PATH
    pause
    exit /b 1
)

REM Check CA exists
if not exist "ca.crt" (
    echo ERROR: CA certificate not found
    echo Please run generate-certs.bat first to generate CA certificate
    pause
    exit /b 1
)

if not exist "ca.key" (
    echo ERROR: CA private key not found
    echo Please run generate-certs.bat first to generate CA certificate
    pause
    exit /b 1
)

REM Generate private key
echo [1/4] Generating private key for %SERVICE_NAME%...
openssl genrsa -out %SERVICE_NAME%.key 2048
if errorlevel 1 goto error
echo Private key generated successfully

REM Generate CSR
echo [2/4] Generating certificate signing request...
openssl req -new -key %SERVICE_NAME%.key -out %SERVICE_NAME%.csr -subj "/CN=%SERVICE_NAME%"
if errorlevel 1 goto error
echo CSR generated successfully

REM Sign with CA
echo [3/4] Signing certificate with CA...
openssl x509 -req -days 10950 -in %SERVICE_NAME%.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out %SERVICE_NAME%.crt
if errorlevel 1 goto error
echo Certificate signed successfully

REM Cleanup temporary files
echo.
echo Cleaning up temporary files...
del %SERVICE_NAME%.csr 2>nul
del ca.srl 2>nul

REM Create client distribution directory
echo [4/4] Creating client distribution package...
if not exist "clients" mkdir clients
if not exist "clients\%SERVICE_NAME%" mkdir clients\%SERVICE_NAME%

REM Copy files to client distribution directory
copy /Y ca.crt clients\%SERVICE_NAME%\ >nul
copy /Y %SERVICE_NAME%.crt clients\%SERVICE_NAME%\ >nul
copy /Y %SERVICE_NAME%.key clients\%SERVICE_NAME%\ >nul
echo Client package created: clients\%SERVICE_NAME%\

REM Generate README for client
echo # TLS Certificate Package for %SERVICE_NAME% > clients\%SERVICE_NAME%\README.md
echo. >> clients\%SERVICE_NAME%\README.md
echo Generated: %date% %time% >> clients\%SERVICE_NAME%\README.md
echo. >> clients\%SERVICE_NAME%\README.md
echo ## Files >> clients\%SERVICE_NAME%\README.md
echo. >> clients\%SERVICE_NAME%\README.md
echo - ca.crt - CA Root Certificate (trusted certificate) >> clients\%SERVICE_NAME%\README.md
echo - %SERVICE_NAME%.crt - Service Certificate >> clients\%SERVICE_NAME%\README.md
echo - %SERVICE_NAME%.key - Service Private Key (KEEP SECRET!) >> clients\%SERVICE_NAME%\README.md
echo. >> clients\%SERVICE_NAME%\README.md
echo ## Deployment >> clients\%SERVICE_NAME%\README.md
echo. >> clients\%SERVICE_NAME%\README.md
echo Copy all files in this folder to your service directory: >> clients\%SERVICE_NAME%\README.md
echo. >> clients\%SERVICE_NAME%\README.md
echo ```bash >> clients\%SERVICE_NAME%\README.md
echo # Copy to service directory >> clients\%SERVICE_NAME%\README.md
echo mkdir tls >> clients\%SERVICE_NAME%\README.md
echo copy *.* tls\ >> clients\%SERVICE_NAME%\README.md
echo ``` >> clients\%SERVICE_NAME%\README.md
echo. >> clients\%SERVICE_NAME%\README.md
echo ## Configuration (Default Paths) >> clients\%SERVICE_NAME%\README.md
echo. >> clients\%SERVICE_NAME%\README.md
echo After copying to tls/ directory, use these default paths: >> clients\%SERVICE_NAME%\README.md
echo. >> clients\%SERVICE_NAME%\README.md
echo ### Go SDK >> clients\%SERVICE_NAME%\README.md
echo ```go >> clients\%SERVICE_NAME%\README.md
echo tlsConfig := ^&client.TLSConfig{ >> clients\%SERVICE_NAME%\README.md
echo     CaFile:     "tls/ca.crt", >> clients\%SERVICE_NAME%\README.md
echo     CertFile:   "tls/%SERVICE_NAME%.crt", >> clients\%SERVICE_NAME%\README.md
echo     KeyFile:    "tls/%SERVICE_NAME%.key", >> clients\%SERVICE_NAME%\README.md
echo     ServerName: "nats-server", >> clients\%SERVICE_NAME%\README.md
echo } >> clients\%SERVICE_NAME%\README.md
echo ``` >> clients\%SERVICE_NAME%\README.md
echo. >> clients\%SERVICE_NAME%\README.md
echo ### Python SDK >> clients\%SERVICE_NAME%\README.md
echo ```python >> clients\%SERVICE_NAME%\README.md
echo tls_config = TLSConfig( >> clients\%SERVICE_NAME%\README.md
echo     ca_file="tls/ca.crt", >> clients\%SERVICE_NAME%\README.md
echo     cert_file="tls/%SERVICE_NAME%.crt", >> clients\%SERVICE_NAME%\README.md
echo     key_file="tls/%SERVICE_NAME%.key", >> clients\%SERVICE_NAME%\README.md
echo     server_name="nats-server" >> clients\%SERVICE_NAME%\README.md
echo ) >> clients\%SERVICE_NAME%\README.md
echo ``` >> clients\%SERVICE_NAME%\README.md
echo. >> clients\%SERVICE_NAME%\README.md
echo ### Environment Variables (Optional) >> clients\%SERVICE_NAME%\README.md
echo ```bash >> clients\%SERVICE_NAME%\README.md
echo set NATS_URL=tls://172.18.200.47:4222 >> clients\%SERVICE_NAME%\README.md
echo set TLS_CA=tls/ca.crt >> clients\%SERVICE_NAME%\README.md
echo set TLS_CERT=tls/%SERVICE_NAME%.crt >> clients\%SERVICE_NAME%\README.md
echo set TLS_KEY=tls/%SERVICE_NAME%.key >> clients\%SERVICE_NAME%\README.md
echo set TLS_SERVER_NAME=nats-server >> clients\%SERVICE_NAME%\README.md
echo ``` >> clients\%SERVICE_NAME%\README.md

echo.
echo ========================================
echo Certificate generation completed!
echo ========================================
echo.
echo Generated files in current directory:
echo - %SERVICE_NAME%.key (Private Key - KEEP SECRET!)
echo - %SERVICE_NAME%.crt (Certificate)
echo.
echo Client distribution package: clients\%SERVICE_NAME%\
echo   - ca.crt
echo   - %SERVICE_NAME%.crt
echo   - %SERVICE_NAME%.key
echo   - README.md
echo.
echo To deploy:
echo   1. Copy clients\%SERVICE_NAME%\* to your service directory\tls\
echo   2. Use default paths: tls/ca.crt, tls/%SERVICE_NAME%.crt, tls/%SERVICE_NAME%.key
echo   3. Set ServerName to "nats-server" for certificate verification
echo.

REM Append to manifest
if exist cert-manifest.txt (
    echo %SERVICE_NAME%: clients\%SERVICE_NAME% >> cert-manifest.txt
) else (
    echo # LightLink Certificate Manifest >> cert-manifest.txt
    echo # Format: ^<service-name^>: ^<client-distribution-path^> >> cert-manifest.txt
    echo. >> cert-manifest.txt
    echo %SERVICE_NAME%: clients\%SERVICE_NAME% >> cert-manifest.txt
)

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
