@echo off
REM Test RPC Services - One at a time
REM This script tests each math-service implementation individually

setlocal enabledelayedexpansion

set NATS_URL=nats://172.18.200.47:4222
set CERT_DIR=light_link_platform/client

echo ========================================
echo RPC Service Testing Script
echo ========================================
echo.
echo This script will test each service individually.
echo Please start only ONE service at a time.
echo.

:menu
echo Available tests:
echo 1. Test Go math-service
echo 2. Test C# math-service
echo 3. Test Python math-service
echo 4. Test all services (sequential)
echo 5. Exit
echo.
set /p choice="Select test to run (1-5): "

if "%choice%"=="1" goto test_go
if "%choice%"=="2" goto test_csharp
if "%choice%"=="3" goto test_python
if "%choice%"=="4" goto test_all
if "%choice%"=="5" goto end
goto menu

:test_go
echo.
echo ========================================
echo Testing Go math-service
echo ========================================
echo.
echo Make sure Go math-service is running:
echo   cd light_link_platform/examples/provider/go/math-service
echo   go run main.go
echo.
pause
echo.
echo Sending RPC calls to: math-service
echo Subject: $SRV.math-service.add
echo.
go run test_rpc_client.go math-service add
echo.
goto menu

:test_csharp
echo.
echo ========================================
echo Testing C# math-service
echo ========================================
echo.
echo Make sure C# math-service is running:
echo   cd light_link_platform/examples/provider/csharp/MathService
echo   dotnet run
echo.
pause
echo.
echo Sending RPC calls to: math-service-csharp
echo Subject: $SRV.math-service-csharp.add
echo.
go run test_rpc_client.go math-service-csharp add
echo.
goto menu

:test_python
echo.
echo ========================================
echo Testing Python math-service
echo ========================================
echo.
echo Make sure Python math-service is running:
echo   cd light_link_platform/examples/provider/python/math_service
echo   python main.py
echo.
pause
echo.
echo Sending RPC calls to: math-service
echo Subject: $SRV.math-service.add
echo.
go run test_rpc_client.go math-service add
echo.
goto menu

:test_all
echo.
echo ========================================
echo Testing All Services Sequentially
echo ========================================
echo.
echo This will test each service one by one.
echo You will need to start each service when prompted.
echo.

echo.
echo [1/3] Go math-service
echo Please start: cd light_link_platform/examples/provider/go/math-service ^&^& go run main.go
pause
go run test_rpc_client.go math-service add
echo.

echo [2/3] C# math-service
echo Please start: cd light_link_platform/examples/provider/csharp/MathService ^&^& dotnet run
pause
go run test_rpc_client.go math-service-csharp add
echo.

echo [3/3] Python math-service
echo Please start: cd light_link_platform/examples/provider/python/math_service ^&^& python main.py
pause
go run test_rpc_client.go math-service add
echo.

goto menu

:end
echo.
echo Testing complete.
pause
