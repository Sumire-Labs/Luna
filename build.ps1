# PowerShell Build Script for Windows
param(
    [Parameter(Position=0)]
    [string]$Task = "build"
)

$BINARY_NAME = "luna-bot.exe"
$MAIN_PATH = "cmd/bot/main.go"

switch ($Task) {
    "build" {
        Write-Host "Building Luna Bot..." -ForegroundColor Green
        go build -o $BINARY_NAME $MAIN_PATH
    }
    "run" {
        Write-Host "Running Luna Bot..." -ForegroundColor Green
        go run $MAIN_PATH
    }
    "clean" {
        Write-Host "Cleaning..." -ForegroundColor Yellow
        go clean
        Remove-Item -Path $BINARY_NAME -ErrorAction SilentlyContinue
        Remove-Item -Path "dist" -Recurse -ErrorAction SilentlyContinue
    }
    "deps" {
        Write-Host "Installing dependencies..." -ForegroundColor Green
        go mod download
        go mod tidy
    }
    "test" {
        Write-Host "Running tests..." -ForegroundColor Green
        go test -v ./...
    }
    "fmt" {
        Write-Host "Formatting code..." -ForegroundColor Green
        go fmt ./...
    }
    default {
        Write-Host "Available tasks:" -ForegroundColor Cyan
        Write-Host "  build  - Build the application"
        Write-Host "  run    - Run the application"
        Write-Host "  clean  - Clean build artifacts"
        Write-Host "  deps   - Install dependencies"
        Write-Host "  test   - Run tests"
        Write-Host "  fmt    - Format code"
    }
}