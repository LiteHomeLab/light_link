@echo off
cd /d %~dp0

echo ========================================
echo Starting Call Math Service Go
echo ========================================
echo.
echo Prerequisites:
echo   - Go must be installed
echo   - math-service-go provider must be running
echo.

REM Check if Go is installed
where go >nul 2>&1
if %ERRORLEVEL% neq 0 (
    echo ERROR: Go is not installed or not in PATH
    pause
    exit /b 1
)

echo Starting call-math-service-go...
go run main.go

if %ERRORLEVEL% neq 0 (
    echo.
    echo ERROR: Service exited with error code %ERRORLEVEL%
    pause
    exit /b %ERRORLEVEL%
)

echo.
echo Service completed successfully
pause
