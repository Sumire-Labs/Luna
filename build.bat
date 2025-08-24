@echo off
setlocal

set BINARY_NAME=luna-bot.exe
set MAIN_PATH=cmd/bot/main.go

if "%1"=="" goto build
if "%1"=="build" goto build
if "%1"=="run" goto run
if "%1"=="clean" goto clean
if "%1"=="deps" goto deps
if "%1"=="test" goto test
if "%1"=="fmt" goto fmt
goto help

:build
echo Building Luna Bot...
go build -o %BINARY_NAME% %MAIN_PATH%
goto end

:run
echo Running Luna Bot...
go run %MAIN_PATH%
goto end

:clean
echo Cleaning...
go clean
del /Q %BINARY_NAME% 2>nul
rmdir /S /Q dist 2>nul
goto end

:deps
echo Installing dependencies...
go mod download
go mod tidy
goto end

:test
echo Running tests...
go test -v ./...
goto end

:fmt
echo Formatting code...
go fmt ./...
goto end

:help
echo Available commands:
echo   build.bat build  - Build the application
echo   build.bat run    - Run the application
echo   build.bat clean  - Clean build artifacts
echo   build.bat deps   - Install dependencies
echo   build.bat test   - Run tests
echo   build.bat fmt    - Format code

:end
endlocal