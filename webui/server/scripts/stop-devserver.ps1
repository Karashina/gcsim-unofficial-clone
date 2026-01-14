<#
.SYNOPSIS
    Stop the running gcsim development server.

.DESCRIPTION
    This script finds and stops any running devserver processes on port 8381.

.PARAMETER Port
    Port number to check (default: 8381).

.PARAMETER Force
    Force stop without confirmation.

.EXAMPLE
    .\scripts\stop-devserver.ps1
    Find and stop devserver with confirmation

.EXAMPLE
    .\scripts\stop-devserver.ps1 -Force
    Stop immediately without confirmation

.EXAMPLE
    .\scripts\stop-devserver.ps1 -Port 9000
    Stop server running on port 9000

Comments are in English as requested.
#>

param(
    [string]$Port = "8381",
    [switch]$Force = $false
)

Write-Host "========================================" -ForegroundColor Cyan
Write-Host " Stop gcsim Development Server" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Find process using the port
$existingProcess = Get-NetTCPConnection -LocalPort $Port -ErrorAction SilentlyContinue | 
    Select-Object -First 1 -ExpandProperty OwningProcess

if (-not $existingProcess) {
    Write-Host "✓ No server running on port $Port" -ForegroundColor Green
    exit 0
}

$proc = Get-Process -Id $existingProcess -ErrorAction SilentlyContinue
if (-not $proc) {
    Write-Host "✓ No valid process found (port may have been just released)" -ForegroundColor Green
    exit 0
}

Write-Host "Found running server:" -ForegroundColor Yellow
Write-Host "  Process: $($proc.ProcessName)" -ForegroundColor White
Write-Host "  PID:     $existingProcess" -ForegroundColor White
Write-Host "  Port:    $Port" -ForegroundColor White
Write-Host ""

if (-not $Force) {
    $response = Read-Host "Stop this process? (y/N)"
    if ($response -ne 'y' -and $response -ne 'Y') {
        Write-Host "Cancelled." -ForegroundColor Yellow
        exit 0
    }
}

try {
    Stop-Process -Id $existingProcess -Force -ErrorAction Stop
    Write-Host ""
    Write-Host "✓ Server stopped successfully!" -ForegroundColor Green
    Write-Host ""
    Write-Host "You can restart it with:" -ForegroundColor Gray
    Write-Host "  .\scripts\run-local-devserver.ps1" -ForegroundColor White
} catch {
    Write-Host ""
    Write-Error "Failed to stop process: $_"
    exit 1
}
