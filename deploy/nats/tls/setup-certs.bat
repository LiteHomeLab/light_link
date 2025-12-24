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
echo   2. client/ folder - for client services
echo.

REM Get script directory
set "SCRIPT_DIR=%~dp0"
cd /d "%SCRIPT_DIR%"

REM ========================================
REM Find OpenSSL
REM ========================================

echo [1/7] Locating OpenSSL...
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

echo [2/7] Generating CA certificate...
"%OPENSSL_CMD%" genrsa -out ca.key 2048
if errorlevel 1 goto error
"%OPENSSL_CMD%" req -new -x509 -days 10950 -key ca.key -out ca.crt -subj "/CN=LightLink CA"
if errorlevel 1 goto error
echo CA certificate created
echo.

REM ========================================
REM Generate NATS Server Certificate
REM ========================================

echo [3/7] Generating NATS server certificate...
"%OPENSSL_CMD%" genrsa -out server.key 2048
if errorlevel 1 goto error
"%OPENSSL_CMD%" req -new -key server.key -out server.csr -subj "/CN=nats-server"
if errorlevel 1 goto error
"%OPENSSL_CMD%" x509 -req -days 10950 -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt
if errorlevel 1 goto error
echo NATS server certificate created
echo.

REM ========================================
REM Generate Client Certificates
REM ========================================

echo [4/7] Generating client certificates...

"%OPENSSL_CMD%" genrsa -out demo-service.key 2048
"%OPENSSL_CMD%" req -new -key demo-service.key -out demo-service.csr -subj "/CN=demo-service"
"%OPENSSL_CMD%" x509 -req -days 10950 -in demo-service.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out demo-service.crt

"%OPENSSL_CMD%" genrsa -out test-service.key 2048
"%OPENSSL_CMD%" req -new -key test-service.key -out test-service.csr -subj "/CN=test-service"
"%OPENSSL_CMD%" x509 -req -days 10950 -in test-service.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out test-service.crt

"%OPENSSL_CMD%" genrsa -out client-app.key 2048
"%OPENSSL_CMD%" req -new -key client-app.key -out client-app.csr -subj "/CN=client-app"
"%OPENSSL_CMD%" x509 -req -days 10950 -in client-app.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out client-app.crt

echo Client certificates created
echo.

REM ========================================
REM Create NATS Server Folder
REM ========================================

echo [5/7] Creating nats-server folder...
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

echo [6/7] Creating client folder...
if exist "client" rmdir /S /Q client
mkdir client

copy /Y ca.crt client\ >nul
copy /Y demo-service.crt client\ >nul
copy /Y demo-service.key client\ >nul
copy /Y test-service.crt client\ >nul
copy /Y test-service.key client\ >nul
copy /Y client-app.crt client\ >nul
copy /Y client-app.key client\ >nul

echo client/ folder created
echo.

REM ========================================
REM Generate README Files
REM ========================================

echo [7/7] Generating documentation...

REM README for nats-server folder
echo # NATS Server Certificate Package > nats-server\README.txt
echo. >> nats-server\README.txt
echo These certificates are for the NATS server. >> nats-server\README.txt
echo. >> nats-server\README.txt
echo Files: >> nats-server\README.txt
echo   - ca.crt (CA certificate) >> nats-server\README.txt
echo   - server.crt (Server certificate) >> nats-server\README.txt
echo   - server.key (Server private key) >> nats-server\README.txt

REM README for client folder
echo # Client Certificate Package > client\README.txt
echo. >> client\README.txt
echo Copy this entire folder to your client service directory. >> client\README.txt
echo. >> client\README.txt
echo Files: >> client\README.txt
echo   - ca.crt (CA certificate) >> client\README.txt
echo   - demo-service.crt / demo-service.key >> client\README.txt
echo   - test-service.crt / test-service.key >> client\README.txt
echo   - client-app.crt / client-app.key >> client\README.txt
echo. >> client\README.txt
echo Configuration example: >> client\README.txt
echo   ca: client/ca.crt >> client\README.txt
echo   cert: client/demo-service.crt >> client\README.txt
echo   key: client/demo-service.key >> client\README.txt
echo   server_name: nats-server >> client\README.txt

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
echo [2] client/ - For client services
echo     - ca.crt
echo     - demo-service.crt / demo-service.key
echo     - test-service.crt / test-service.key
echo     - client-app.crt / client-app.key
echo.
echo Deploy nats-server/ to your NATS server.
echo Deploy client/ to your client services.
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
