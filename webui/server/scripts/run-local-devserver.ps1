<#
.SYNOPSIS
    Run the development server and open the web UI in a browser.

.DESCRIPTION
    This script runs the Go devserver at ./cmd/devserver which provides
    a simple static UI under /ui/ and mock API endpoints.

.PARAMETER BuildFrontend
    Build the frontend (webui-src) before starting the server.

.PARAMETER Foreground
    Run the server in the foreground (blocks terminal, shows logs).

.PARAMETER OpenBrowser
    Automatically open the browser (default: true).

.PARAMETER Port
    Port to run the server on (default: 8381).

.PARAMETER Help
    Display this help message.

.EXAMPLE
    .\scripts\run-local-devserver.ps1
    Start the server in background and open browser

.EXAMPLE
    .\scripts\run-local-devserver.ps1 -Foreground -BuildFrontend
    Build frontend and run server in foreground with live logs

.EXAMPLE
    .\scripts\run-local-devserver.ps1 -Port 9000 -OpenBrowser:$false
    Run on port 9000 without opening browser

Comments are in English as requested.
#>

param(
    [switch]$BuildFrontend = $false,
    [switch]$Foreground = $false,
    [switch]$OpenBrowser = $true,
    [string]$Port = "8381",
    [switch]$Help = $false
)

if ($Help) {
    Get-Help $MyInvocation.MyCommand.Path -Detailed
    exit 0
}

$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
# repo root is parent of the scripts directory
$root = Split-Path -Parent $scriptDir

# expose port to child processes via env var
$env:DEVSERVER_PORT = $Port

function Check-Command($name) { return (Get-Command $name -ErrorAction SilentlyContinue) -ne $null }

if (-not (Check-Command "go")) {
    Write-Error "Go not found in PATH. Please install Go and ensure 'go' is available."
    exit 1
}

# Optionally build frontend
if ($BuildFrontend) {
    Write-Host "Building frontend (webui-src)..."
    if (-not (Test-Path (Join-Path $root "webui-src"))) {
        Write-Warning "webui-src directory not found; skipping frontend build."
    } else {
        if (-not (Check-Command "npm")) {
            Write-Error "npm not found in PATH. Install Node.js and npm to build the frontend or omit -BuildFrontend."
            exit 1
        }
        Push-Location (Join-Path $root "webui-src")
        # Install and build
        Write-Host "Running: npm install"
        npm install
        if ($LASTEXITCODE -ne 0) { Pop-Location; Write-Error "npm install failed"; exit 1 }
        Write-Host "Running: npm run build"
        npm run build
        if ($LASTEXITCODE -ne 0) { Pop-Location; Write-Error "npm run build failed"; exit 1 }
        Pop-Location
        Write-Host "Frontend build complete."
    }
}

$url = "http://localhost:$Port/ui/"

# Check if port is already in use (for both foreground and background modes)
$existingProcess = Get-NetTCPConnection -LocalPort $Port -ErrorAction SilentlyContinue | 
    Select-Object -First 1 -ExpandProperty OwningProcess

if ($existingProcess) {
    $proc = Get-Process -Id $existingProcess -ErrorAction SilentlyContinue
    if ($proc) {
        Write-Host ""
        Write-Host "========================================" -ForegroundColor Red
        Write-Warning "Port $Port is already in use!"
        Write-Host "Process: $($proc.ProcessName) (PID: $existingProcess)" -ForegroundColor Yellow
        Write-Host "========================================" -ForegroundColor Red
        Write-Host ""
        $response = Read-Host "Kill existing process and continue? (y/N)"
        if ($response -eq 'y' -or $response -eq 'Y') {
            Stop-Process -Id $existingProcess -Force
            Write-Host "✓ Process stopped. Waiting 2 seconds..." -ForegroundColor Green
            Start-Sleep -Seconds 2
        } else {
            Write-Host "Exiting without starting new server." -ForegroundColor Yellow
            exit 0
        }
    }
}

Write-Host "========================================" -ForegroundColor Cyan
Write-Host " gcsim Development Server" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Port: $Port"
Write-Host "URL:  $url"
Write-Host ""

if ($Foreground) {
    Write-Host "Starting devserver in foreground (Ctrl+C to stop)..." -ForegroundColor Yellow
    Write-Host ""
    Push-Location $root
    & go run ./cmd/devserver
    $exitCode = $LASTEXITCODE
    Pop-Location
    exit $exitCode
} else {
    Write-Host "Starting devserver in background..." -ForegroundColor Yellow
    
    # Start devserver in a new hidden PowerShell window with job tracking
    $startInfo = New-Object System.Diagnostics.ProcessStartInfo
    $startInfo.FileName = "go"
    $startInfo.Arguments = "run ./cmd/devserver"
    $startInfo.WorkingDirectory = $root
    $startInfo.UseShellExecute = $false
    $startInfo.RedirectStandardOutput = $true
    $startInfo.RedirectStandardError = $true
    $startInfo.CreateNoWindow = $true
    
    $proc = New-Object System.Diagnostics.Process
    $proc.StartInfo = $startInfo
    $proc.Start() | Out-Null
    
    Write-Host ""
    Write-Host "Waiting for server to start..." -ForegroundColor Yellow
    Start-Sleep -Seconds 2

    # Verify server is running
    try {
        $response = Invoke-WebRequest -Uri $url -TimeoutSec 3 -UseBasicParsing -ErrorAction Stop
        Write-Host "✓ Server is running!" -ForegroundColor Green
    } catch {
        Write-Warning "Server may not have started successfully. Check manually: $url"
    }

    Write-Host ""
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host "Devserver started (PID: $($proc.Id))" -ForegroundColor Green
    Write-Host ""
    Write-Host "To view logs:" -ForegroundColor White
    Write-Host "  .\scripts\run-local-devserver.ps1 -Foreground" -ForegroundColor Gray
    Write-Host ""
    Write-Host "To stop the server:" -ForegroundColor White
    Write-Host "  Stop-Process -Id $($proc.Id)" -ForegroundColor Gray
    Write-Host "  Or find and kill the process: Get-Process | Where-Object {`$_.ProcessName -like '*go*'}" -ForegroundColor Gray
    Write-Host "========================================" -ForegroundColor Cyan
    Write-Host ""

    if ($OpenBrowser) {
        Write-Host "Opening browser..." -ForegroundColor Yellow
        try {
            Start-Process $url
        } catch {
            Write-Warning "Could not open browser automatically. Please open: $url"
        }
    } else {
        Write-Host "Server is ready at: $url" -ForegroundColor White
    }
}

