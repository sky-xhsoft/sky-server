@echo off
chcp 65001 >nul
echo ========================================
echo    Sky-Server Linux 64-bit Build
echo ========================================
echo.

rem Set Go cross-compile environment variables
set GOOS=linux
set GOARCH=amd64
set CGO_ENABLED=0

echo [1/3] Cleaning old build...
if exist bin\sky-server (del /F /Q bin\sky-server)
if exist dist\sky-server-linux-amd64.tar.gz (del /F /Q dist\sky-server-linux-amd64.tar.gz)
mkdir dist 2>nul

echo.
echo [2/3] Downloading dependencies...
go mod download

echo.
echo [3/3] Building Linux 64-bit version...
go build -ldflags="-s -w" -o bin/sky-server ./cmd/server

if %ERRORLEVEL% neq 0 (
    echo.
    echo Build failed! Please check errors above.
    pause
    exit /b 1
)


if %ERRORLEVEL% neq 0 (
    echo.
    echo Packing failed! Make sure tar is installed (Git Bash or Windows 10+ has it).
    pause
    exit /b 1
)

echo.
echo ========================================
echo Build completed!
echo Output: dist/sky-server-linux-amd64.tar.gz
echo ========================================
echo.
pause
