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

REM Check OpenSSL in PATH
set "OPENSSL_CMD="
where openssl >nul 2>&1
if not errorlevel 1 (
    set "OPENSSL_CMD=openssl"
    goto openssl_found
)

REM Check common OpenSSL installation paths
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

REM OpenSSL not found
echo ERROR: OpenSSL not found
echo.
echo Please install one of the following:
echo   1. Git for Windows: https://git-scm.com/download/win
echo   2. OpenSSL for Windows: https://slproweb.com/products/Win32OpenSSL.html
echo.
echo Then add OpenSSL to your PATH or restart this script.
pause
exit /b 1

:openssl_found
echo Using OpenSSL: %OPENSSL_CMD%

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
"%OPENSSL_CMD%" genrsa -out %SERVICE_NAME%.key 2048
if errorlevel 1 goto error
echo Private key generated successfully

REM Generate CSR
echo [2/4] Generating certificate signing request...
"%OPENSSL_CMD%" req -new -key %SERVICE_NAME%.key -out %SERVICE_NAME%.csr -subj "/CN=%SERVICE_NAME%"
if errorlevel 1 goto error
echo CSR generated successfully

REM Sign with CA
echo [3/4] Signing certificate with CA...
"%OPENSSL_CMD%" x509 -req -days 10950 -in %SERVICE_NAME%.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out %SERVICE_NAME%.crt
if errorlevel 1 goto error
echo Certificate signed successfully

REM Cleanup temporary files
echo.
echo Cleaning up temporary files...
del %SERVICE_NAME%.csr 2>nul
del ca.srl 2>nul

REM Add to client directory
echo [4/4] Adding certificate to client package...
if not exist "client" mkdir client

REM Ensure ca.crt exists in client directory
if not exist "client\ca.crt" (
    copy /Y ca.crt client\ >nul
)

REM Copy service certificate to client directory
copy /Y %SERVICE_NAME%.crt client\ >nul
copy /Y %SERVICE_NAME%.key client\ >nul
echo Certificate added to client\ directory

echo.
echo ========================================
echo Certificate generation completed!
echo ========================================
echo.
echo Generated files in current directory:
echo - %SERVICE_NAME%.key (Private Key - KEEP SECRET!)
echo - %SERVICE_NAME%.crt (Certificate)
echo.
echo Files added to client\ directory:
echo - ca.crt
echo - %SERVICE_NAME%.crt
echo - %SERVICE_NAME%.key
echo.
echo To deploy:
echo   1. Copy the entire client\ folder to your service directory
echo   2. Configure your service to use:
echo      ca: client/ca.crt
echo      cert: client/%SERVICE_NAME%.crt
echo      key: client/%SERVICE_NAME%.key
echo      server_name: nats-server
echo.

REM Append to manifest
if exist cert-manifest.txt (
    echo %SERVICE_NAME%: client/%SERVICE_NAME%.crt, %SERVICE_NAME%.key >> cert-manifest.txt
) else (
    echo # LightLink Certificate Manifest >> cert-manifest.txt
    echo # Format: ^<service-name^>: ^<cert-file^>, ^<key-file^> >> cert-manifest.txt
    echo. >> cert-manifest.txt
    echo %SERVICE_NAME%: client/%SERVICE_NAME%.crt, %SERVICE_NAME%.key >> cert-manifest.txt
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
