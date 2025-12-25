@echo off
REM United Debug - Start All Services (Providers + Manager Base + Web)
REM This script starts all services including the web UI

setlocal EnableDelayedExpansion

REM Set project root
set "PROJECT_ROOT=%~dp0"
set "CLIENT_CERTS=%PROJECT_ROOT%client"

echo ========================================
echo LightLink United Debug - Full Start
echo ========================================
echo.
echo Client Certs: %CLIENT_CERTS%
echo.

REM Check if client certs exist
if not exist "%CLIENT_CERTS%\ca.crt" (
    echo ERROR: Client certificates not found in %CLIENT_CERTS%
    echo Please ensure the client folder exists with certificates.
    pause
    exit /b 1
)

REM Check if node_modules exists for web
if not exist "%PROJECT_ROOT%manager_base\web\node_modules" (
    echo WARNING: node_modules not found in web folder
    echo Please run: cd manager_base\web ^&^& npm install
    echo.
)

echo [1/7] Starting Go Provider (math-service)...
start "Go-math-service" /MIN /D "%PROJECT_ROOT%examples\provider\go\math-service" math-service.exe
timeout /t 2 /nobreak >nul

echo [2/7] Starting Python Provider (data-service)...
start "Python-data-service" /MIN /D "%PROJECT_ROOT%examples\provider\python\data_service" python main.py
timeout /t 2 /nobreak >nul

echo [3/7] Starting Python Provider (math-service)...
start "Python-math-service" /MIN /D "%PROJECT_ROOT%examples\provider\python\math_service" python main.py
timeout /t 2 /nobreak >nul

echo [4/7] Starting C# Provider (MathService)...
start "CSharp-MathService" /MIN /D "%PROJECT_ROOT%examples\provider\csharp\MathService\bin\Release\net6.0" MathService.exe
timeout /t 3 /nobreak >nul

echo [5/7] Starting C# Provider (TextService)...
start "CSharp-TextService" /MIN /D "%PROJECT_ROOT%examples\provider\csharp\TextService\bin\Release\net6.0" TextService.exe
timeout /t 3 /nobreak >nul

echo [6/7] Starting Manager Base Server...
start "Manager-Base-Server" /MIN /D "%PROJECT_ROOT%manager_base\server" go run main.go
timeout /t 5 /nobreak >nul

echo [7/7] Starting Web UI...
start "Manager-Base-Web" /D "%PROJECT_ROOT%manager_base\web" cmd /k npm run dev
timeout /t 3 /nobreak >nul

echo.
echo ========================================
echo All services started!
echo ========================================
echo.
echo Services running:
echo   - Go math-service       (window: Go-math-service)
echo   - Python data-service   (window: Python-data-service)
echo   - Python math-service   (window: Python-math-service)
echo   - C# MathService        (window: CSharp-MathService)
echo   - C# TextService        (window: CSharp-TextService)
echo   - Manager Base Server   (window: Manager-Base-Server)
echo   - Manager Base Web      (window: Manager-Base-Web)
echo.
echo Access URLs:
echo   - Web UI:  http://localhost:5173
echo   - API:     http://localhost:8080
echo.
echo Default login:
echo   - Username: admin
echo   - Password: admin123
echo.
echo Press any key to stop all services...
pause >nul

echo.
echo Stopping all services...
taskkill /FI "WINDOWTITLE eq Go-math-service*" /F >nul 2>&1
taskkill /FI "WINDOWTITLE eq Python-data-service*" /F >nul 2>&1
taskkill /FI "WINDOWTITLE eq Python-math-service*" /F >nul 2>&1
taskkill /FI "WINDOWTITLE eq CSharp-MathService*" /F >nul 2>&1
taskkill /FI "WINDOWTITLE eq CSharp-TextService*" /F >nul 2>&1
taskkill /FI "WINDOWTITLE eq Manager-Base-Server*" /F >nul 2>&1
taskkill /FI "WINDOWTITLE eq Manager-Base-Web*" /F >nul 2>&1

echo All services stopped.
