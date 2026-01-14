<#
  Helper script to set up Cloudflare API credentials for deploy_webui.ps1
  
  Usage:
    .\setup_cloudflare_env.ps1 -ApiToken "your_api_token" -ZoneId "your_zone_id"
    
  This will create encrypted credential files that can be used by deploy_webui.ps1
#>

param(
    [Parameter(Mandatory=$true)]
    [string]$ApiToken,

    [Parameter(Mandatory=$true)]
    [string]$ZoneId,
    
    [string]$CredentialsDir = "scripts"
)

function Write-Log { param($m) $ts = Get-Date -Format "yyyy-MM-dd HH:mm:ss"; Write-Host "[$ts] $m" }

Write-Log "Setting up Cloudflare API credentials"

# Ensure credentials directory exists
$credDir = Join-Path (Get-Location) $CredentialsDir
if (-not (Test-Path $credDir)) {
    New-Item -ItemType Directory -Path $credDir -Force | Out-Null
}

# Encrypt and save API token
$secureApiToken = ConvertTo-SecureString $ApiToken -AsPlainText -Force
$encryptedApiToken = ConvertFrom-SecureString $secureApiToken
$apiTokenFile = Join-Path $credDir ".cloudflare_api_token_enc.txt"
$encryptedApiToken | Set-Content -Path $apiTokenFile -Encoding UTF8

# Encrypt and save Zone ID
$secureZoneId = ConvertTo-SecureString $ZoneId -AsPlainText -Force
$encryptedZoneId = ConvertFrom-SecureString $secureZoneId
$zoneIdFile = Join-Path $credDir ".cloudflare_zone_id_enc.txt"
$encryptedZoneId | Set-Content -Path $zoneIdFile -Encoding UTF8

Write-Log "Cloudflare credentials saved to:"
Write-Log "  API Token: $apiTokenFile"
Write-Log "  Zone ID: $zoneIdFile"
Write-Log ""
Write-Log "Cloudflare credentials are now automatically loaded by deploy_webui.ps1"
Write-Log "You can use deploy_webui.ps1 with automatic Cloudflare cache clearing:"
Write-Log "  .\scripts\deploy_webui.ps1 -Server 'your_server' -User 'your_user' -Domain 'your_domain'"
Write-Log ""
Write-Log "The encrypted credentials will be automatically detected from:"
Write-Log "  API Token: $apiTokenFile"
Write-Log "  Zone ID: $zoneIdFile"