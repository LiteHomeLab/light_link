@echo off
REM OpenSSL Diagnostic Tool for LightLink
echo ========================================
echo OpenSSL Diagnostic Tool
echo ========================================
echo.

echo [1] Checking PATH for openssl...
where openssl >nul 2>&1
if not errorlevel 1 (
    echo FOUND in PATH:
    where openssl
    echo.
    goto test_openssl
)
echo NOT found in PATH
echo.

echo [2] Checking common Git for Windows paths...
set "FOUND=0"
if exist "C:\Program Files\Git\mingw64\bin\openssl.exe" (
    echo FOUND: C:\Program Files\Git\mingw64\bin\openssl.exe
    set "OPENSSL_CMD=C:\Program Files\Git\mingw64\bin\openssl.exe"
    set "FOUND=1"
)
if exist "C:\Program Files\Git\usr\bin\openssl.exe" (
    echo FOUND: C:\Program Files\Git\usr\bin\openssl.exe
    set "OPENSSL_CMD=C:\Program Files\Git\usr\bin\openssl.exe"
    set "FOUND=1"
)
if exist "C:\Program Files (x86)\Git\mingw64\bin\openssl.exe" (
    echo FOUND: C:\Program Files (x86)\Git\mingw64\bin\openssl.exe
    set "OPENSSL_CMD=C:\Program Files (x86)\Git\mingw64\bin\openssl.exe"
    set "FOUND=1"
)
if exist "C:\Program Files (x86)\Git\usr\bin\openssl.exe" (
    echo FOUND: C:\Program Files (x86)\Git\usr\bin\openssl.exe
    set "OPENSSL_CMD=C:\Program Files (x86)\Git\usr\bin\openssl.exe"
    set "FOUND=1"
)

if "%FOUND%"=="0" (
    echo NOT found in Git paths
    echo.
) else (
    echo.
    goto test_openssl
)

echo [3] Checking OpenSSL-Win64 paths...
if exist "C:\Program Files\OpenSSL-Win64\bin\openssl.exe" (
    echo FOUND: C:\Program Files\OpenSSL-Win64\bin\openssl.exe
    set "OPENSSL_CMD=C:\Program Files\OpenSSL-Win64\bin\openssl.exe"
    set "FOUND=1"
    echo.
    goto test_openssl
)
if exist "C:\Program Files (x86)\OpenSSL-Win32\bin\openssl.exe" (
    echo FOUND: C:\Program Files (x86)\OpenSSL-Win32\bin\openssl.exe
    set "OPENSSL_CMD=C:\Program Files (x86)\OpenSSL-Win32\bin\openssl.exe"
    set "FOUND=1"
    echo.
    goto test_openssl
)
echo NOT found in OpenSSL paths
echo.

echo [4] Searching C: drive for openssl.exe (this may take a while)...
echo Searching...
for /f "delims=" %%f in ('dir /s /b "C:\openssl.exe" 2^>nul') do (
    echo FOUND: %%f
    set "OPENSSL_CMD=%%f"
    set "FOUND=1"
)
echo.

if "%FOUND%"=="0" (
    echo ========================================
    echo ERROR: OpenSSL not found anywhere!
    echo ========================================
    echo.
    echo Please install Git for Windows or OpenSSL-Win64
    echo.
    echo Git for Windows: https://git-scm.com/download/win
    echo OpenSSL-Win64: https://slproweb.com/products/Win32OpenSSL.html
    pause
    exit /b 1
)

:test_openssl
echo ========================================
echo Testing OpenSSL
echo ========================================
echo.
"%OPENSSL_CMD%" version
if errorlevel 1 (
    echo ERROR: OpenSSL found but failed to run
    pause
    exit /b 1
)
echo.
echo SUCCESS: OpenSSL is working!
echo.
echo Path: %OPENSSL_CMD%
echo.
echo You can now run: generate-service-cert.bat ^<service-name^>
echo.
pause
