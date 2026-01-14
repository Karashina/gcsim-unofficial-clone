<#
  Simple deploy script for webui using scp/ssh over a custom external port.

  Usage example:
    .\deploy_webui.ps1 -Server "192.168.1.233" -User "deploy" -Port 7056 -RemotePath "/var/www/gcsim-uoc"

  Requirements:
    - OpenSSH client available (ssh, scp) in PATH on Windows host.
    - Destination server reachable at given port (router forwards 7056 -> server SSH port).
#>

param(
  # Defaults embedded so frequent values don't need to be re-specified.
  # You can still override these by passing parameters on the command line.
  [string]$Server = "192.168.1.233",

  [string]$User = "uocuser",

  # Note: this script will not pass an explicit SSH port to ssh/scp;
  # ssh/scp will use the default port (22) or the user's SSH config (~/.ssh/config).
    [string]$RemotePath = "/var/www/html",

    [string]$LocalWebuiPath = "webui",

    [string]$KeyFile = "C:\Users\linol\HL-Creds\id_ed25519",

    [string]$BuildCmd = "npm run --prefix webui run build",

    [switch]$KeepRemoteTmp,

  [switch]$ReloadNginx = $true,
  [switch]$ConfigureNginx,
  [switch]$UseNginx,
  [string]$Domain = "gcsim-uoc.linole.net",
  [switch]$EnablePasswordlessSudo,
  [string]$SudoPasswordFile = "C:\Users\linol\HL-Creds\enc_sudo_pw.txt",
  [switch]$ClearCloudflareCache = $true,
  [string]$CloudflareApiToken = $null,
  [string]$CloudflareZoneId = $null,
  [string]$CloudflareApiTokenFile = "C:\Users\linol\HL-Creds\cloudflare_api_token_enc.txt",
  [string]$CloudflareZoneIdFile = "C:\Users\linol\HL-Creds\cloudflare_zone_id_enc.txt"
)

function Write-Log { param($m) $ts = Get-Date -Format "yyyy-MM-dd HH:mm:ss"; Write-Host "[$ts] $m" }

Write-Log "Starting deploy_webui.ps1"

# Validate local path
if (-not (Test-Path -LiteralPath $LocalWebuiPath)) {
    Write-Error "Local path '$LocalWebuiPath' not found."
    exit 2
}

# If a sudo password file is provided, validate it early so we fail fast before uploading files
if ($SudoPasswordFile) {
  if (-not (Test-Path -LiteralPath $SudoPasswordFile)) {
    Write-Error "Sudo password file not found: $SudoPasswordFile"
    exit 4
  }
}

# Optional build step
# Optional build step
# If a local source build directory exists (webui-src) prefer building locally there.
$localSrcRel = 'webui-src'
$localSrcPath = Join-Path (Get-Location) $localSrcRel
if (Test-Path -LiteralPath $localSrcPath) {
  Write-Log "Found local source directory '$localSrcRel' — building WebUI from source"
  $npmCmd = Get-Command npm -ErrorAction SilentlyContinue
  if (-not $npmCmd) {
    Write-Error "npm not found in PATH; cannot perform build. Please install Node.js."
    exit 5
  }
  
  # Ensure dependencies are installed
  Write-Log "Installing dependencies: npm install"
  Push-Location $localSrcPath
  try {
    npm install
    if ($LASTEXITCODE -ne 0) { 
      Write-Error "npm install failed (exit $LASTEXITCODE)"
      exit $LASTEXITCODE
    }
    
    # Run build
    Write-Log "Building WebUI: npm run build"
    npm run build
    if ($LASTEXITCODE -ne 0) { 
      Write-Error "npm run build failed (exit $LASTEXITCODE)"
      exit $LASTEXITCODE
    }
    
    Write-Log "WebUI build completed successfully"
  } finally {
    Pop-Location
  }
} else {
  Write-Warning "webui-src directory not found; deploying existing webui/ content without rebuild"
}

# Verify webui/app.js exists after build
$builtAppJs = Join-Path $LocalWebuiPath "app.js"
if (-not (Test-Path -LiteralPath $builtAppJs)) {
  Write-Error "Build verification failed: webui/app.js not found after build"
  exit 6
}
Write-Log "Build verification: webui/app.js exists ($(Get-Item $builtAppJs | Select-Object -ExpandProperty Length) bytes)"

# Check scp/ssh
if (-not (Get-Command scp -ErrorAction SilentlyContinue)) { Write-Error "scp not found in PATH."; exit 3 }
if (-not (Get-Command ssh -ErrorAction SilentlyContinue)) { Write-Error "ssh not found in PATH."; exit 3 }

 $sshTarget = "$User@$Server"
 $timestamp = (Get-Date -Format "yyyyMMddHHmmss")
 $remoteTmpName = ".deploy_webui_$timestamp"
 $remoteTmp = "~/$remoteTmpName"

 Write-Log "Creating remote temp directory $remoteTmp"
 $sshArgs = @()
 if ($KeyFile) { $sshArgs += @("-i", $KeyFile) }
 $sshArgs += $sshTarget

 & ssh @sshArgs "mkdir -p $remoteTmp" | Out-Null

# If requested, install a sudoers file on the remote host to allow passwordless sudo for deploy commands.
if ($EnablePasswordlessSudo) {
  Write-Log "Preparing sudoers file for user $User"
  $sudoersContent = @"
# gcsim deploy sudoers - created by deploy_webui.ps1
$User ALL=(ALL) NOPASSWD: /bin/mv, /bin/rm, /bin/chown, /bin/mkdir, /bin/systemctl, /usr/bin/systemctl
"@

  # write with LF newlines to avoid CRLF issues
  $localSudoTmp = [System.IO.Path]::GetTempFileName()
  $sudoText = $sudoersContent -replace "`r`n","`n"
  $utf8NoBom = New-Object System.Text.UTF8Encoding($false)
  [System.IO.File]::WriteAllText($localSudoTmp, $sudoText, $utf8NoBom)

  # upload to remote temp
  Write-Log "Uploading sudoers file to ${sshTarget}:$remoteTmp/gcsim-deploy-sudoers"
  $scpArgs = @()
  if ($KeyFile) { $scpArgs += @("-i", $KeyFile) }
  $scpArgs += @($localSudoTmp, "${sshTarget}:$remoteTmp/gcsim-deploy-sudoers")
  $scpCmd = Get-Command scp -ErrorAction Stop
  $scpProc = Start-Process -FilePath $scpCmd.Source -ArgumentList $scpArgs -NoNewWindow -Wait -PassThru
  if ($scpProc.ExitCode -ne 0) { Write-Error "scp (sudoers) failed with exit code $($scpProc.ExitCode)"; exit $scpProc.ExitCode }

  # Decrypt sudo password file if provided (DPAPI-encrypted via ConvertFrom-SecureString)
  $sudoPlain = $null
  if ($SudoPasswordFile) {
    if (-not (Test-Path $SudoPasswordFile)) { Write-Error "Sudo password file not found: $SudoPasswordFile"; exit 4 }
    $enc = Get-Content -Path $SudoPasswordFile -Raw
    # Normalize: remove BOM and whitespace/newlines
    $enc = $enc.Trim()
    if ($enc.Length -gt 0 -and [int][char]$enc[0] -eq 0xFEFF) { $enc = $enc.Substring(1) }
    $enc = $enc -replace "\r|\n", ""
    try {
      $secure = ConvertTo-SecureString -String $enc
      $bstr = [System.Runtime.InteropServices.Marshal]::SecureStringToBSTR($secure)
      $sudoPlain = [System.Runtime.InteropServices.Marshal]::PtrToStringAuto($bstr)
      [System.Runtime.InteropServices.Marshal]::ZeroFreeBSTR($bstr)
    } catch {
      Write-Error "Failed to decrypt sudo password file. Ensure it was created with scripts/encode_sudo_password.ps1 by the same Windows user. Error: $_"; exit 5
    }
  }

  # Check if sudo exists on remote host
  Write-Log "Checking for sudo on remote host"
  $checkOut = & ssh @sshArgs "command -v sudo >/dev/null 2>&1 && echo OK || echo MISSING"
  if (($checkOut -join "`n").Trim() -ne "OK") {
    Write-Log "Remote 'sudo' not found; skipping sudoers installation. You may need to install sudo on the server or run the necessary commands manually."
  } else {
    # Move into place using sudo (this will prompt for sudo password if necessary)
    Write-Log "Installing sudoers file on remote host (may prompt for sudo password)"
    $installCmd = "if [ -f /etc/sudoers.d/gcsim-deploy ]; then sudo mv /etc/sudoers.d/gcsim-deploy /etc/sudoers.d/gcsim-deploy.bak || true; fi; sudo mv $remoteTmp/gcsim-deploy-sudoers /etc/sudoers.d/gcsim-deploy; sudo chmod 0440 /etc/sudoers.d/gcsim-deploy; sudo visudo -c -f /etc/sudoers.d/gcsim-deploy"
    if ($sudoPlain) {
      # feed password to each sudo via -S (wrap whole command in bash -c to keep piping)
      $installCmd = $installCmd -replace 'sudo ', "echo '$sudoPlain' | sudo -S "
      & ssh @sshArgs "bash -lc '$installCmd'"
    } else {
      & ssh @sshArgs $installCmd
    }
    Write-Log "Remote sudoers installation step finished"
  }
}

 Write-Log "Uploading '$LocalWebuiPath' to ${sshTarget}:$remoteTmp"
 $scpArgs = @()
 if ($KeyFile) { $scpArgs += @("-i", $KeyFile) }
 $scpArgs += @("-r", $LocalWebuiPath, "${sshTarget}:$remoteTmp/")

 $scpCmd = Get-Command scp -ErrorAction Stop
 $scpProc = Start-Process -FilePath $scpCmd.Source -ArgumentList $scpArgs -NoNewWindow -Wait -PassThru
if ($scpProc.ExitCode -ne 0) { Write-Error "scp failed with exit code $($scpProc.ExitCode)"; exit $scpProc.ExitCode }

# Upload authentication code directories if they exist (internal/auth, internal/db)
$authDirs = @("internal\auth", "internal\db")
foreach ($authDir in $authDirs) {
  if (Test-Path -LiteralPath $authDir) {
    Write-Log "Uploading '$authDir' to ${sshTarget}:$remoteTmp/"
    $scpArgs = @()
    if ($KeyFile) { $scpArgs += @("-i", $KeyFile) }
    $scpArgs += @("-r", $authDir, "${sshTarget}:$remoteTmp/")
    $scpProc = Start-Process -FilePath $scpCmd.Source -ArgumentList $scpArgs -NoNewWindow -Wait -PassThru
    if ($scpProc.ExitCode -ne 0) { Write-Error "scp ($authDir) failed with exit code $($scpProc.ExitCode)"; exit $scpProc.ExitCode }
  } else {
    Write-Log "Authentication directory '$authDir' not found - skipping"
  }
}

# Create remote deploy script using a here-string with placeholders to avoid PowerShell interpolation problems
$localDirName = Split-Path -Leaf $LocalWebuiPath
$remoteScript = @'
#!/bin/bash
set -euo pipefail
TMP_DIR="__REMOTE_TMP__"
WEBUI_DIR="$TMP_DIR/__LOCAL_DIRNAME__"
TARGET="__REMOTE_PATH__"

echo "remote: preparing target $TARGET"

# CRITICAL: Preserve database directory before any rotation
if [ -d "$TARGET/data" ]; then
  echo "remote: preserving database directory"
  sudo mkdir -p /tmp/gcsim_data_preserve || true
  sudo cp -a "$TARGET/data" /tmp/gcsim_data_preserve/ || true
fi

if [ -d "$TARGET" ]; then
  echo "remote: rotating existing target"
  sudo rm -rf "${TARGET}.old" || true
  sudo mv "$TARGET" "${TARGET}.old" || true
fi

sudo mkdir -p "$(dirname "$TARGET")" || true
sudo rm -rf "$TARGET" || true
sudo mv "$WEBUI_DIR" "$TARGET"
sudo chown -R www-data:www-data "$TARGET" || true

# CRITICAL: Restore database directory after rotation
if [ -d /tmp/gcsim_data_preserve/data ]; then
  echo "remote: restoring database directory"
  sudo rm -rf "$TARGET/data" || true
  sudo mv /tmp/gcsim_data_preserve/data "$TARGET/data" || true
  sudo rm -rf /tmp/gcsim_data_preserve || true
  echo "remote: database directory restored successfully"
fi
if [ "__KEEP_TMP__" != "1" ]; then
  rm -rf "$TMP_DIR"
fi
__RELOAD_CADDY_SNIPPET__
echo "remote: deploy complete"

# If authentication directories were uploaded, install them to /opt/gcsim/internal/
if [ -d "$TMP_DIR/internal/auth" ] || [ -d "$TMP_DIR/internal/db" ]; then
  echo "remote: installing authentication code to /opt/gcsim/internal/"
  sudo mkdir -p /opt/gcsim/internal || true
  if [ -d "$TMP_DIR/internal/auth" ]; then
    sudo rm -rf /opt/gcsim/internal/auth || true
    sudo mv "$TMP_DIR/internal/auth" /opt/gcsim/internal/ || true
    echo "remote: installed internal/auth"
  fi
  if [ -d "$TMP_DIR/internal/db" ]; then
    sudo rm -rf /opt/gcsim/internal/db || true
    sudo mv "$TMP_DIR/internal/db" /opt/gcsim/internal/ || true
    echo "remote: installed internal/db"
  fi
  sudo chown -R www-data:www-data /opt/gcsim/internal || true
fi

# If a backend binary was uploaded, install it to /usr/local/bin and create a systemd unit
if [ -f "$TMP_DIR/gcsim-webui" ]; then
  echo "remote: installing backend binary to /usr/local/bin/gcsim-webui"
  
  # Stop service before replacing binary
  sudo systemctl stop gcsim-webui 2>/dev/null || true
  
  # Move binary with error checking
  if sudo mv "$TMP_DIR/gcsim-webui" /usr/local/bin/gcsim-webui; then
    echo "remote: binary moved successfully"
  else
    echo "remote: ERROR: failed to move binary" >&2
    exit 1
  fi
  
  # Set executable permission with error checking
  if sudo chmod +x /usr/local/bin/gcsim-webui; then
    echo "remote: binary permissions set"
  else
    echo "remote: ERROR: failed to set binary permissions" >&2
    exit 1
  fi
  
  # Verify binary exists and is executable
  if [ ! -x /usr/local/bin/gcsim-webui ]; then
    echo "remote: ERROR: binary not found or not executable after installation" >&2
    exit 1
  fi
  
  echo "remote: binary verified at /usr/local/bin/gcsim-webui"
  ls -lh /usr/local/bin/gcsim-webui
  
  sudo tee /etc/systemd/system/gcsim-webui.service >/dev/null <<'UNIT'
[Unit]
Description=gcsim-webui HTTP API with JWT Authentication
After=network.target
[Service]
User=www-data
WorkingDirectory=/var/www/html
EnvironmentFile=-/etc/gcsim-webui/.env
ExecStart=/usr/local/bin/gcsim-webui -addr=:8382
Restart=on-failure
RestartSec=5
[Install]
WantedBy=multi-user.target
UNIT
  sudo systemctl daemon-reload || true
  
  # Create environment directory with proper permissions
  sudo mkdir -p /etc/gcsim-webui || true
  sudo chown www-data:www-data /etc/gcsim-webui || true
  sudo chmod 750 /etc/gcsim-webui || true
  
  # If server.env was uploaded, install it to /etc/gcsim-webui/.env
  if [ -f "$TMP_DIR/server.env" ]; then
    echo "remote: installing server.env to /etc/gcsim-webui/.env"
    sudo mv "$TMP_DIR/server.env" /etc/gcsim-webui/.env || true
    sudo chown www-data:www-data /etc/gcsim-webui/.env || true
    sudo chmod 600 /etc/gcsim-webui/.env || true
    echo "remote: server.env installed successfully"
  else
    echo "remote: WARNING: server.env not found - service may fail to start"
  fi
  
  # Create data directory for database
  sudo mkdir -p /var/www/html/data || true
  sudo mkdir -p /var/www/html/data/backups || true
  
  # CRITICAL: Backup existing database before any operations
  if [ -f /var/www/html/data/gcsim.db ]; then
    BACKUP_NAME="gcsim_backup_\$(date +%Y%m%d_%H%M%S).db"
    echo "remote: BACKING UP database to /var/www/html/data/backups/\$BACKUP_NAME"
    sudo cp -p /var/www/html/data/gcsim.db "/var/www/html/data/backups/\$BACKUP_NAME" || true
    
    # Keep only last 10 backups
    sudo bash -c 'cd /var/www/html/data/backups && ls -t gcsim_backup_*.db 2>/dev/null | tail -n +11 | xargs -r rm --' || true
    echo "remote: Database backup completed"
  fi
  
  # Set ownership for data directory
  sudo chown -R www-data:www-data /var/www/html/data || true
  sudo chmod 750 /var/www/html/data || true
  sudo chmod 750 /var/www/html/data/backups 2>/dev/null || true
  
  # Fix permissions for existing database files
  if [ -f /var/www/html/data/gcsim.db ]; then
    sudo chown www-data:www-data /var/www/html/data/gcsim.db || true
    sudo chmod 660 /var/www/html/data/gcsim.db || true
    echo "remote: fixed permissions for gcsim.db (rw for www-data)"
  else
    echo "remote: WARNING: gcsim.db not found - will be created on first run"
  fi
  
  # Fix permissions for SQLite journal and WAL files if they exist
  sudo chown www-data:www-data /var/www/html/data/gcsim.db-* 2>/dev/null || true
  sudo chmod 660 /var/www/html/data/gcsim.db-* 2>/dev/null || true
  
  echo "remote: Database protection and permissions configured"
  
  # Restart service to apply new binary and env file
  echo "remote: reloading systemd daemon"
  sudo systemctl daemon-reload || true
  
  echo "remote: enabling gcsim-webui service"
  sudo systemctl enable gcsim-webui || true
  
  echo "remote: restarting gcsim-webui service"
  if sudo systemctl restart gcsim-webui; then
    echo "remote: service restarted successfully"
  else
    echo "remote: ERROR: service restart failed" >&2
    sudo journalctl -u gcsim-webui --no-pager -n 20
    exit 1
  fi
  
  # Wait a moment for service to start
  sleep 2
  
  # Verify service is running
  if sudo systemctl is-active --quiet gcsim-webui; then
    echo "remote: service is active and running"
    sudo systemctl status gcsim-webui --no-pager -l | head -10
  else
    echo "remote: ERROR: service failed to start" >&2
    sudo systemctl status gcsim-webui --no-pager -l
    sudo journalctl -u gcsim-webui --no-pager -n 30
    exit 1
  fi
  
  echo "remote: backend installed and systemd service restarted successfully"
fi
'@

# Replace placeholders with actual values
$keepVal = if ($KeepRemoteTmp) { '1' } else { '0' }
 # Use the uploader's home directory explicitly so running the script under sudo (which changes $HOME to /root)
 # does not make the script look in /root for the uploaded files. Use /home/<User> which is where we uploaded.
 $userHomeAbs = "/home/$User"
 $remoteScript = $remoteScript.Replace('__REMOTE_TMP__', "$userHomeAbs/$remoteTmpName")
$remoteScript = $remoteScript.Replace('__LOCAL_DIRNAME__', $localDirName)
$remoteScript = $remoteScript.Replace('__REMOTE_PATH__', $RemotePath)
$remoteScript = $remoteScript.Replace('__KEEP_TMP__', $keepVal)

# Optionally reload nginx on the remote host and clear cache
if ($ReloadNginx) {
  $reloadSnippet = @"
# Clear nginx cache directories if they exist
sudo find /var/cache/nginx -type f -delete 2>/dev/null || true
sudo find /tmp/nginx -type f -delete 2>/dev/null || true
# Clear any proxy cache directories
sudo find /var/lib/nginx/proxy -type f -delete 2>/dev/null || true
# Reload nginx to apply changes
sudo systemctl reload nginx || true
echo "remote: nginx cache cleared and reloaded"
"@
} else {
  $reloadSnippet = "# nginx reload skipped"
}
$remoteScript = $remoteScript -replace '__RELOAD_CADDY_SNIPPET__', $reloadSnippet

# Write temporary script file
$tmpScript = [System.IO.Path]::GetTempFileName()
$tmpScriptPath = "$tmpScript.sh"
  # Ensure LF line endings to avoid /bin/bash CRLF issues on remote
  $remoteScript = $remoteScript -replace "`r`n","`n"
  $utf8NoBom = New-Object System.Text.UTF8Encoding($false)
  [System.IO.File]::WriteAllText($tmpScriptPath, $remoteScript, $utf8NoBom)

Write-Log "Uploading remote deploy script to ${sshTarget}:$remoteTmp/deploy_remote.sh"
$scpArgs = @()
if ($KeyFile) { $scpArgs += @("-i", $KeyFile) }
$scpArgs += @($tmpScriptPath, "${sshTarget}:$remoteTmp/deploy_remote.sh")

$scpProc2 = Start-Process -FilePath $scpCmd.Source -ArgumentList $scpArgs -NoNewWindow -Wait -PassThru
if ($scpProc2.ExitCode -ne 0) { Write-Error "scp (script) failed with exit code $($scpProc2.ExitCode)"; exit $scpProc2.ExitCode }

# If a local Linux backend binary exists (gcsim-webui), upload it so the remote script can install it
$localBackendBinary = "gcsim-webui"
if (Test-Path -LiteralPath $localBackendBinary) {
  Write-Log "Uploading backend binary '$localBackendBinary' to ${sshTarget}:$remoteTmp/"
  $scpArgs = @()
  if ($KeyFile) { $scpArgs += @("-i", $KeyFile) }
  $scpArgs += @($localBackendBinary, "${sshTarget}:$remoteTmp/")
  $scpProcBin = Start-Process -FilePath $scpCmd.Source -ArgumentList $scpArgs -NoNewWindow -Wait -PassThru
  if ($scpProcBin.ExitCode -ne 0) { Write-Error "scp (binary) failed with exit code $($scpProcBin.ExitCode)"; exit $scpProcBin.ExitCode }
}

# If server.env exists locally, upload it so the remote script can install it as /etc/gcsim-webui/.env
$localServerEnv = "server.env"
if (Test-Path -LiteralPath $localServerEnv) {
  Write-Log "Uploading server environment file '$localServerEnv' to ${sshTarget}:$remoteTmp/"
  $scpArgs = @()
  if ($KeyFile) { $scpArgs += @("-i", $KeyFile) }
  $scpArgs += @($localServerEnv, "${sshTarget}:$remoteTmp/")
  $scpProcEnv = Start-Process -FilePath $scpCmd.Source -ArgumentList $scpArgs -NoNewWindow -Wait -PassThru
  if ($scpProcEnv.ExitCode -ne 0) { Write-Error "scp (server.env) failed with exit code $($scpProcEnv.ExitCode)"; exit $scpProcEnv.ExitCode }
} else {
  Write-Warning "server.env not found in current directory - service may fail without environment variables"
}

Write-Log "Setting executable permission and running remote script"
$sshExecArgs = @()
if ($KeyFile) { $sshExecArgs += @("-i", $KeyFile) }
$sshExecArgs += $sshTarget

# If we have an encrypted sudo password file, decrypt now and use it to feed sudo -S when running the remote deploy script
if ($SudoPasswordFile) {
  $enc = Get-Content -Path $SudoPasswordFile -Raw
  $enc = $enc.Trim()
  if ($enc.Length -gt 0 -and [int][char]$enc[0] -eq 0xFEFF) { $enc = $enc.Substring(1) }
  $enc = $enc -replace "\r|\n", ""
  try {
    $secure = ConvertTo-SecureString -String $enc
    $bstr = [System.Runtime.InteropServices.Marshal]::SecureStringToBSTR($secure)
    $sudoPlainForRun = [System.Runtime.InteropServices.Marshal]::PtrToStringAuto($bstr)
    [System.Runtime.InteropServices.Marshal]::ZeroFreeBSTR($bstr)
  } catch {
    Write-Error "Failed to decrypt sudo password file for run. Ensure file was created with scripts/encode_sudo_password.ps1 by the same Windows user. Error: $_"; exit 5
  }

  # We will run the remote script via bash -lc and prefix sudo with echo 'pw' | sudo -S
  # escape single quotes for remote single-quoted literal
  $escapedPw = $sudoPlainForRun -replace "'","'\''"
  # Use printf on the remote side to feed the password to sudo -S, then run the script as root
  $remoteRunCmd = "chmod +x $remoteTmp/deploy_remote.sh; printf '%s\n' '$escapedPw' | sudo -S bash $remoteTmp/deploy_remote.sh"
  $sshExecArgs += @($remoteRunCmd)
  $sshCmd = Get-Command ssh -ErrorAction Stop
  $proc = Start-Process -FilePath $sshCmd.Source -ArgumentList $sshExecArgs -NoNewWindow -Wait -PassThru
  if ($proc.ExitCode -ne 0) { Write-Error "Remote deploy failed with exit code $($proc.ExitCode)"; exit $proc.ExitCode }
} else {
  $sshExecArgs += @("chmod +x $remoteTmp/deploy_remote.sh && bash $remoteTmp/deploy_remote.sh")
  $sshCmd = Get-Command ssh -ErrorAction Stop
  $proc = Start-Process -FilePath $sshCmd.Source -ArgumentList $sshExecArgs -NoNewWindow -Wait -PassThru
  if ($proc.ExitCode -ne 0) { Write-Error "Remote deploy failed with exit code $($proc.ExitCode)"; exit $proc.ExitCode }
}

  # Caddy flow removed — nginx-only configuration lives below
## NGiNX configuration only
if ($ConfigureNginx) {
  if (-not $Domain) { Write-Error "-ConfigureNginx requires -Domain <domain>"; exit 6 }
  if (-not $UseNginx) { Write-Error "-ConfigureNginx requires -UseNginx switch"; exit 6 }

  Write-Log "Configuring nginx for domain $Domain to serve $RemotePath and listen on external port 7056"
  $sanDomain = $Domain -replace '[^A-Za-z0-9._-]','_'

  $installScript = @'
#!/bin/bash
set -euo pipefail
# Install nginx if missing
if ! command -v nginx >/dev/null 2>&1; then
  apt-get update
  apt-get install -y nginx openssl
fi

# Ensure directories
mkdir -p /etc/nginx/sites-available /etc/nginx/sites-enabled /etc/ssl/private /etc/ssl/certs

# If an Origin certificate/key are not provided, generate a self-signed origin cert
ORIG_KEY=/etc/ssl/private/gcsim-origin.key
ORIG_CRT=/etc/ssl/certs/gcsim-origin.pem
if [ ! -f "$ORIG_KEY" ] || [ ! -f "$ORIG_CRT" ]; then
  echo "Generating self-signed origin certificate for nginx (will be used by Cloudflare Tunnel as origin)"
  openssl req -x509 -nodes -days 3650 -newkey rsa:2048 \
    -keyout "$ORIG_KEY" -out "$ORIG_CRT" -subj "/CN=__DOMAIN__"
  chmod 600 "$ORIG_KEY"
  chmod 644 "$ORIG_CRT"
fi

cat > /etc/nginx/sites-available/gcsim-deploy-__SAN__ <<'NGINX'
server {
  listen 7056 default_server;
  server_name __DOMAIN__;
  return 301 https://$host$request_uri;
}

server {
  listen 443 ssl http2;
  server_name __DOMAIN__;

  # Use origin certificate so Cloudflare Tunnel (or Cloudflare proxy) can connect via HTTPS
  ssl_certificate __ORIG_CRT__;
  ssl_certificate_key __ORIG_KEY__;
  ssl_protocols TLSv1.2 TLSv1.3;
  ssl_prefer_server_ciphers on;
  ssl_ciphers HIGH:!aNULL:!MD5;

  root __REMOTE_PATH__;
  index index.html index.htm;

  # Ensure responses include UTF-8 charset where appropriate so browsers interpret
  # JavaScript and text resources correctly (helps avoid mojibake on some clients).
  charset utf-8;
  charset_types text/html text/plain text/css application/javascript application/json text/xml application/xml;

  location / {
    try_files $uri $uri/ =404;
    # Disable caching for HTML files to ensure fresh content
    location ~* \.html$ {
      add_header Cache-Control "no-cache, no-store, must-revalidate";
      add_header Pragma "no-cache";
      add_header Expires "0";
    }
    # Disable caching for JavaScript and CSS files during development
    location ~* \.(js|css)$ {
      add_header Cache-Control "no-cache, no-store, must-revalidate";
      add_header Pragma "no-cache";
      add_header Expires "0";
    }
  }
  # Proxy API requests to the local backend service
  location /api/ {
    proxy_pass http://127.0.0.1:8382;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    proxy_http_version 1.1;
    proxy_set_header Connection "";
  }
}
NGINX

ln -sf /etc/nginx/sites-available/gcsim-deploy-__SAN__ /etc/nginx/sites-enabled/gcsim-deploy-__SAN__
nginx -t || true
systemctl reload nginx || true
'@

  $userHomeAbs = "/home/$User"
  $installScript = $installScript.Replace('__REMOTE_TMP__', "$userHomeAbs/$remoteTmpName")
  $installScript = $installScript.Replace('__REMOTE_PATH__', $RemotePath)
  $installScript = $installScript.Replace('__DOMAIN__', $Domain)
  $installScript = $installScript.Replace('__SAN__', $sanDomain)
  # Replace origin cert placeholders with actual origin cert paths produced/used on the server
  $installScript = $installScript.Replace('__ORIG_CRT__', '/etc/ssl/certs/gcsim-origin.pem')
  $installScript = $installScript.Replace('__ORIG_KEY__', '/etc/ssl/private/gcsim-origin.key')
  $installScript = $installScript -replace "`r`n","`n"

  $localInstallTmp = [System.IO.Path]::GetTempFileName()
  $localInstallPath = "$localInstallTmp.sh"
  $utf8NoBom = New-Object System.Text.UTF8Encoding($false)
  [System.IO.File]::WriteAllText($localInstallPath, $installScript, $utf8NoBom)

  Write-Log "Ensuring remote temp directory $remoteTmp exists"
  & ssh @sshArgs "mkdir -p $remoteTmp" | Out-Null

  Write-Log "Uploading nginx installer to ${sshTarget}:$remoteTmp/install_nginx.sh"
  $scpArgs = @()
  if ($KeyFile) { $scpArgs += @("-i", $KeyFile) }
  $scpArgs += @($localInstallPath, "${sshTarget}:$remoteTmp/install_nginx.sh")
  $scpCmd = Get-Command scp -ErrorAction Stop
  $scpProc = Start-Process -FilePath $scpCmd.Source -ArgumentList $scpArgs -NoNewWindow -Wait -PassThru
  if ($scpProc.ExitCode -ne 0) { Write-Error "scp (nginx installer) failed with exit code $($scpProc.ExitCode)"; exit $scpProc.ExitCode }

  $remoteInstallPathAbs = "/home/$User/$remoteTmpName/install_nginx.sh"
  if ($SudoPasswordFile) {
    $enc = Get-Content -Path $SudoPasswordFile -Raw
    $enc = $enc.Trim()
    if ($enc.Length -gt 0 -and [int][char]$enc[0] -eq 0xFEFF) { $enc = $enc.Substring(1) }
    $enc = $enc -replace "\r|\n", ""
    try {
      $secure = ConvertTo-SecureString -String $enc
      $bstr = [System.Runtime.InteropServices.Marshal]::SecureStringToBSTR($secure)
      $sudoPlainForNginx = [System.Runtime.InteropServices.Marshal]::PtrToStringAuto($bstr)
      [System.Runtime.InteropServices.Marshal]::ZeroFreeBSTR($bstr)
    } catch {
      Write-Error "Failed to decrypt sudo password file for nginx step. Error: $_"; exit 7
    }
    $escapedPwN = $sudoPlainForNginx -replace "'","'\''"
    $sshNArgs = @()
    if ($KeyFile) { $sshNArgs += @("-i", $KeyFile) }
    $sshNArgs += $sshTarget
    $sshNArgs += @("printf '%s\n' '$escapedPwN' | sudo -S bash $remoteInstallPathAbs")
    $sshCmd = Get-Command ssh -ErrorAction Stop
    $procN = Start-Process -FilePath $sshCmd.Source -ArgumentList $sshNArgs -NoNewWindow -Wait -PassThru
    if ($procN.ExitCode -ne 0) { Write-Error "Remote nginx configuration failed with exit code $($procN.ExitCode)"; exit $procN.ExitCode }
  } else {
    $sshNArgs = @()
    if ($KeyFile) { $sshNArgs += @("-i", $KeyFile) }
    $sshNArgs += @("-t", $sshTarget)
    $sshNArgs += @("sudo bash $remoteInstallPathAbs")
    $sshCmd = Get-Command ssh -ErrorAction Stop
    $procN = Start-Process -FilePath $sshCmd.Source -ArgumentList $sshNArgs -NoNewWindow -Wait -PassThru
    if ($procN.ExitCode -ne 0) { Write-Error "Remote nginx configuration failed with exit code $($procN.ExitCode)"; exit $procN.ExitCode }
  }
  Write-Log "Remote nginx configuration finished for $Domain"
}

# Clear Cloudflare cache if requested
if ($ClearCloudflareCache -and $Domain) {
  Write-Log "Attempting to clear Cloudflare cache for domain $Domain"
  
  # Try to get API credentials from files if not provided directly
  if (-not $CloudflareApiToken) {
    # Check for encrypted file (try relative path first, then absolute)
    $tokenFilePath = $CloudflareApiTokenFile
    if (-not (Test-Path $tokenFilePath) -and -not [System.IO.Path]::IsPathRooted($tokenFilePath)) {
      $tokenFilePath = Join-Path (Get-Location) $CloudflareApiTokenFile
    }
    
    if (Test-Path $tokenFilePath) {
      try {
        $enc = Get-Content -Path $tokenFilePath -Raw
        $enc = $enc.Trim()
        if ($enc.Length -gt 0 -and [int][char]$enc[0] -eq 0xFEFF) { $enc = $enc.Substring(1) }
        $enc = $enc -replace "\r|\n", ""
        $secure = ConvertTo-SecureString -String $enc
        $bstr = [System.Runtime.InteropServices.Marshal]::SecureStringToBSTR($secure)
        $CloudflareApiToken = [System.Runtime.InteropServices.Marshal]::PtrToStringAuto($bstr)
        [System.Runtime.InteropServices.Marshal]::ZeroFreeBSTR($bstr)
        Write-Log "Loaded Cloudflare API token from $tokenFilePath"
      } catch {
        Write-Warning "Failed to decrypt Cloudflare API token file: $_"
      }
    }
    # Fall back to environment variable
    else {
      $CloudflareApiToken = $env:CLOUDFLARE_API_TOKEN
    }
  }
  
  if (-not $CloudflareZoneId) {
    # Check for encrypted file (try relative path first, then absolute)
    $zoneIdFilePath = $CloudflareZoneIdFile
    if (-not (Test-Path $zoneIdFilePath) -and -not [System.IO.Path]::IsPathRooted($zoneIdFilePath)) {
      $zoneIdFilePath = Join-Path (Get-Location) $CloudflareZoneIdFile
    }
    
    if (Test-Path $zoneIdFilePath) {
      try {
        $enc = Get-Content -Path $zoneIdFilePath -Raw
        $enc = $enc.Trim()
        if ($enc.Length -gt 0 -and [int][char]$enc[0] -eq 0xFEFF) { $enc = $enc.Substring(1) }
        $enc = $enc -replace "\r|\n", ""
        $secure = ConvertTo-SecureString -String $enc
        $bstr = [System.Runtime.InteropServices.Marshal]::SecureStringToBSTR($secure)
        $CloudflareZoneId = [System.Runtime.InteropServices.Marshal]::PtrToStringAuto($bstr)
        [System.Runtime.InteropServices.Marshal]::ZeroFreeBSTR($bstr)
        Write-Log "Loaded Cloudflare Zone ID from $zoneIdFilePath"
      } catch {
        Write-Warning "Failed to decrypt Cloudflare Zone ID file: $_"
      }
    }
    # Fall back to environment variable
    else {
      $CloudflareZoneId = $env:CLOUDFLARE_ZONE_ID
    }
  }
  
  if ($CloudflareApiToken -and $CloudflareZoneId) {
    try {
      Write-Log "Clearing Cloudflare cache for zone $CloudflareZoneId"
      
      # Purge all cache for the zone
      $headers = @{
        "Authorization" = "Bearer $CloudflareApiToken"
        "Content-Type" = "application/json"
      }
      
      $body = @{
        "purge_everything" = $true
      } | ConvertTo-Json
      
      $response = Invoke-RestMethod -Uri "https://api.cloudflare.com/client/v4/zones/$CloudflareZoneId/purge_cache" -Method POST -Headers $headers -Body $body
      
      if ($response.success) {
        Write-Log "Cloudflare cache cleared successfully"
      } else {
        Write-Warning "Cloudflare cache clear failed: $($response.errors | ConvertTo-Json)"
      }
    } catch {
      Write-Warning "Failed to clear Cloudflare cache: $_"
    }
  } else {
    Write-Log "Cloudflare cache clear skipped - missing API token or Zone ID"
    Write-Log "You can set up credentials using: .\scripts\setup_cloudflare_env.ps1 -ApiToken 'your_token' -ZoneId 'your_zone_id'"
    Write-Log "Or set environment variables: CLOUDFLARE_API_TOKEN and CLOUDFLARE_ZONE_ID"
    Write-Log "Or use parameters: -CloudflareApiToken and -CloudflareZoneId"
  }
}

Write-Log "Deploy completed successfully. Remote path: $RemotePath"
