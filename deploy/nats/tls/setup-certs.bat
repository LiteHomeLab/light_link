@echo off
REM LightLink Certificate Setup Script
REM Double-click to generate all certificates

setlocal enabledelayedexpansion

REM Generate timestamp folder name
for /f "tokens=2-4 delims=/ " %%a in ('date /t') do (set mydate=%%c-%%a-%%b)
for /f "tokens=1-2 delims=.: " %%a in ('time /t') do (set mytime=%%a-%%b)
set "TIMESTAMP=%mydate%_%mytime%"
set "TIMESTAMP=%TIMESTAMP: =0%"

echo ========================================
echo LightLink Certificate Setup
echo ========================================
echo.
echo Output folder: certs-%TIMESTAMP%\
echo.
echo This will create:
echo   1. nats-server/ folder - for NATS server
echo   2. client/ folder - for all client services
echo.

REM Get script directory
set "SCRIPT_DIR=%~dp0"
cd /d "%SCRIPT_DIR%"

REM ========================================
REM Create Output Folder
REM ========================================

echo Creating output folder...
set "OUTPUT_DIR=%SCRIPT_DIR%certs-%TIMESTAMP%"
if exist "%OUTPUT_DIR%" rmdir /S /Q "%OUTPUT_DIR%"
mkdir "%OUTPUT_DIR%"
cd /d "%OUTPUT_DIR%"
echo Output folder created: certs-%TIMESTAMP%\
echo.

REM ========================================
REM Find OpenSSL
REM ========================================

echo [1/6] Locating OpenSSL...
set "OPENSSL_CMD="

where openssl >nul 2>&1
if not errorlevel 1 (
    set "OPENSSL_CMD=openssl"
    goto openssl_found
)

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

echo ERROR: OpenSSL not found
echo Please install Git for Windows: https://git-scm.com/download/win
pause
exit /b 1

:openssl_found
echo Found OpenSSL
echo.

REM ========================================
REM Generate CA Certificate
REM ========================================

echo [2/6] Generating CA certificate...
"%OPENSSL_CMD%" genrsa -out ca.key 2048
if errorlevel 1 goto error
"%OPENSSL_CMD%" req -new -x509 -days 10950 -key ca.key -out ca.crt -subj "/CN=LightLink CA"
if errorlevel 1 goto error
echo CA certificate created
echo.

REM ========================================
REM Generate NATS Server Certificate
REM ========================================

echo [3/6] Generating NATS server certificate...
"%OPENSSL_CMD%" genrsa -out server.key 2048
if errorlevel 1 goto error
"%OPENSSL_CMD%" req -new -key server.key -out server.csr -subj "/CN=nats-server"
if errorlevel 1 goto error
"%OPENSSL_CMD%" x509 -req -days 10950 -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt
if errorlevel 1 goto error
echo NATS server certificate created
echo.

REM ========================================
REM Generate Client Certificate
REM ========================================

echo [4/6] Generating client certificate (for all clients)...
"%OPENSSL_CMD%" genrsa -out client.key 2048
if errorlevel 1 goto error
"%OPENSSL_CMD%" req -new -key client.key -out client.csr -subj "/CN=lightlink-client"
if errorlevel 1 goto error
"%OPENSSL_CMD%" x509 -req -days 10950 -in client.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out client.crt
if errorlevel 1 goto error
echo Client certificate created
echo.

REM ========================================
REM Create NATS Server Folder
REM ========================================

echo [5/6] Creating nats-server folder...
mkdir nats-server

copy /Y ca.crt nats-server\ >nul
copy /Y server.crt nats-server\ >nul
copy /Y server.key nats-server\ >nul

echo nats-server/ folder created
echo.

REM ========================================
REM Create Client Folder
REM ========================================

echo Creating client folder...
mkdir client

copy /Y ca.crt client\ >nul
copy /Y client.crt client\ >nul
copy /Y client.key client\ >nul

echo client/ folder created
echo.

REM ========================================
REM Generate Documentation
REM ========================================

echo [6/6] Generating documentation...

REM Main README in output folder
echo # LightLink Certificate Package > README.txt
echo. >> README.txt
echo Generated: %date% %time% >> README.txt
echo. >> README.txt
echo ## Folder Structure >> README.txt
echo. >> README.txt
echo This package contains two folders: >> README.txt
echo. >> README.txt
echo ### nats-server/ - NATS Server Certificates >> README.txt
echo Copy this folder to your NATS server directory. >> README.txt
echo. >> README.txt
echo Files: >> README.txt
echo   - ca.crt: CA Root Certificate >> README.txt
echo   - server.crt: NATS Server Certificate >> README.txt
echo   - server.key: NATS Server Private Key (KEEP SECRET!) >> README.txt
echo. >> README.txt
echo ### client/ - Client Service Certificates >> README.txt
echo Copy this folder to ANY client service directory. >> README.txt
echo All client services use the same certificates. >> README.txt
echo. >> README.txt
echo Files: >> README.txt
echo   - ca.crt: CA Root Certificate >> README.txt
echo   - client.crt: Client Certificate (for ALL services) >> README.txt
echo   - client.key: Client Private Key (KEEP SECRET!) >> README.txt
echo. >> README.txt
echo ## File Description >> README.txt
echo. >> README.txt
echo ### CA Certificate (ca.crt + ca.key) >> README.txt
echo   - Root certificate that signs all other certificates >> README.txt
echo   - ca.key: CA private key, kept on certificate generation machine only >> README.txt
echo   - ca.crt: CA public certificate, needed by both server and clients >> README.txt
echo. >> README.txt
echo ### NATS Server Certificate (server.crt + server.key) >> README.txt
echo   - Used by NATS server to identify itself >> README.txt
echo   - CN (Common Name): nats-server >> README.txt
echo   - Used by: NATS server only >> README.txt
echo. >> README.txt
echo ### Client Certificate (client.crt + client.key) >> README.txt
echo   - Used by client services to authenticate to NATS >> README.txt
echo   - CN (Common Name): lightlink-client >> README.txt
echo   - Used by: All client services (Go, Python, C#, etc.) >> README.txt
echo. >> README.txt
echo ## Deployment Steps >> README.txt
echo. >> README.txt
echo ### 1. Deploy to NATS Server >> README.txt
echo   Copy the nats-server/ folder to your NATS server: >> README.txt
echo. >> README.txt
echo   Configure NATS server to use: >> README.txt
echo     - ca_file: nats-server/ca.crt >> README.txt
echo     - cert_file: nats-server/server.crt >> README.txt
echo     - key_file: nats-server/server.key >> README.txt
echo. >> README.txt
echo ### 2. Deploy to Client Services >> README.txt
echo   Copy the client/ folder to your service directory. >> README.txt
echo. >> README.txt
echo   Configure your service to use: >> README.txt
echo     - ca: client/ca.crt >> README.txt
echo     - cert: client/client.crt >> README.txt
echo     - key: client/client.key >> README.txt
echo     - server_name: nats-server (important! must match server cert CN) >> README.txt
echo. >> README.txt
echo ## Security Notes >> README.txt
echo   - .key files contain private keys, keep them secure! >> README.txt
echo   - Do not commit .key files to version control >> README.txt
echo   - Share certificates through secure channels only >> README.txt

REM README for nats-server folder
echo # NATS Server Certificate Package > nats-server\README.txt
echo. >> nats-server\README.txt
echo These certificates are for the NATS server. >> nats-server\README.txt
echo. >> nats-server\README.txt
echo ## Files >> nats-server\README.txt
echo   - ca.crt: CA Root Certificate >> nats-server\README.txt
echo   - server.crt: NATS Server Certificate >> nats-server\README.txt
echo   - server.key: NATS Server Private Key (KEEP SECRET!) >> nats-server\README.txt
echo. >> nats-server\README.txt
echo ## Usage >> nats-server\README.txt
echo Configure NATS server: >> nats-server\README.txt
echo   tls { >> nats-server\README.txt
echo     ca_file: "./nats-server/ca.crt" >> nats-server\README.txt
echo     cert_file: "./nats-server/server.crt" >> nats-server\README.txt
echo     key_file: "./nats-server/server.key" >> nats-server\README.txt
echo   } >> nats-server\README.txt

REM README for client folder
echo # Client Certificate Package > client\README.txt
echo. >> client\README.txt
echo All client services use these same certificates. >> client\README.txt
echo. >> client\README.txt
echo ## Files >> client\README.txt
echo   - ca.crt: CA Root Certificate >> client\README.txt
echo   - client.crt: Client Certificate (for ALL services) >> client\README.txt
echo   - client.key: Client Private Key (KEEP SECRET!) >> client\README.txt
echo. >> client\README.txt
echo ## Usage >> client\README.txt
echo Copy this folder to your service directory. >> client\README.txt
echo. >> client\README.txt
echo ### Go SDK >> client\README.txt
echo   tlsConfig := ^&client.TLSConfig{ >> client\README.txt
echo       CaFile:     "client/ca.crt", >> client\README.txt
echo       CertFile:   "client/client.crt", >> client\README.txt
echo       KeyFile:    "client/client.key", >> client\README.txt
echo       ServerName: "nats-server", >> client\README.txt
echo   } >> client\README.txt
echo. >> client\README.txt
echo ### Python SDK >> client\README.txt
echo   tls_config = TLSConfig( >> client\README.txt
echo       ca_file="client/ca.crt", >> client\README.txt
echo       cert_file="client/client.crt", >> client\README.txt
echo       key_file="client/client.key", >> client\README.txt
echo       server_name="nats-server" >> client\README.txt
echo   ) >> client\README.txt
echo. >> client\README.txt
echo ### C# SDK >> client\README.txt
echo   Options opts = ConnectionFactory.GetDefaultOptions(); >> client\README.txt
echo   opts.Url = "tls://your-nats-server:4222"; >> client\README.txt
echo   opts.SSL = true; >> client\README.txt
echo   opts.SetCertificate("client/ca.crt", "client/client.crt", "client/client.key"); >> client\README.txt

echo Documentation created
echo.

REM ========================================
REM Cleanup
REM ========================================

del *.csr 2>nul
del ca.srl 2>nul

REM ========================================
REM Summary
REM ========================================

echo ========================================
echo Setup Complete!
echo ========================================
echo.
echo Output folder: certs-%TIMESTAMP%\
echo.
echo Contents:
echo.
echo [README.txt] - Complete documentation (read this first!)
echo.
echo [nats-server/] - For NATS server
echo   - ca.crt
echo   - server.crt
echo   - server.key
echo   - README.txt
echo.
echo [client/] - For ALL client services
echo   - ca.crt
echo   - client.crt
echo   - client.key
echo   - README.txt
echo.
echo Next steps:
echo   1. Read certs-%TIMESTAMP%\README.txt
echo   2. Copy nats-server/ to NATS server
echo   3. Copy client/ to client services
echo.
pause
goto end

:error
echo.
echo ========================================
echo ERROR: Setup failed
echo ========================================
echo.
pause
exit /b 1

:end
