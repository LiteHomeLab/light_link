@echo off
REM LightLink Certificate Setup Script
REM Double-click to generate all certificates

setlocal enabledelayedexpansion

echo ========================================
echo LightLink Certificate Setup
echo ========================================
echo.
echo This will create:
echo   1. nats-server/ folder - for NATS server
echo   2. client/ folder - for all client services
echo.

REM Get script directory
set "SCRIPT_DIR=%~dp0"
cd /d "%SCRIPT_DIR%"

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
if exist "nats-server" rmdir /S /Q nats-server
mkdir nats-server

copy /Y ca.crt nats-server\ >nul
copy /Y server.crt nats-server\ >nul
copy /Y server.key nats-server\ >nul

echo nats-server/ folder created
echo.

REM ========================================
REM Create Client Folder
REM ========================================

echo [6/6] Creating client folder...
if exist "client" rmdir /S /Q client
mkdir client

copy /Y ca.crt client\ >nul
copy /Y client.crt client\ >nul
copy /Y client.key client\ >nul

echo client/ folder created
echo.

REM ========================================
REM Generate README Files
REM ========================================

echo Generating documentation...

REM README for nats-server folder
echo # NATS Server Certificate Package > nats-server\README.txt
echo. >> nats-server\README.txt
echo These certificates are for the NATS server. >> nats-server\README.txt
echo. >> nats-server\README.txt
echo Files: >> nats-server\README.txt
echo   - ca.crt (CA certificate) >> nats-server\README.txt
echo   - server.crt (Server certificate) >> nats-server\README.txt
echo   - server.key (Server private key - KEEP SECRET!) >> nats-server\README.txt

REM README for client folder
echo # Client Certificate Package > client\README.txt
echo. >> client\README.txt
echo Copy this entire folder to your client service directory. >> client\README.txt
echo All client services can use the same certificates. >> client\README.txt
echo. >> client\README.txt
echo Files: >> client\README.txt
echo   - ca.crt (CA certificate) >> client\README.txt
echo   - client.crt (Client certificate - for ALL services) >> client\README.txt
echo   - client.key (Client private key - KEEP SECRET!) >> client\README.txt
echo. >> client\README.txt
echo Configuration example: >> client\README.txt
echo   ca: client/ca.crt >> client\README.txt
echo   cert: client/client.crt >> client\README.txt
echo   key: client/client.key >> client\README.txt
echo   server_name: nats-server >> client\README.txt
echo. >> client\README.txt
echo NOTE: This single client certificate can be used by ALL services. >> client\README.txt

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
echo Two folders have been created:
echo.
echo [1] nats-server/ - For NATS server
echo     - ca.crt
echo     - server.crt
echo     - server.key
echo.
echo [2] client/ - For ALL client services
echo     - ca.crt
echo     - client.crt (same for all clients)
echo     - client.key (same for all clients)
echo.
echo Deploy nats-server/ to your NATS server.
echo Deploy client/ to ANY client service (all use same certificate).
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
