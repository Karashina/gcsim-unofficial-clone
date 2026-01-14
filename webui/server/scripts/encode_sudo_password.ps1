<#
Create a DPAPI-encrypted file containing the sudo password for use with deploy_webui.ps1.

Usage:
  .\encode_sudo_password.ps1 -OutFile C:\path\to\encpwd.txt

The script will prompt for the password (SecureString) and write an encrypted string to the output file
which can only be decrypted by the same Windows user account.
#>
param(
    [Parameter(Mandatory=$true)]
    [string]$OutFile
)

Write-Host "Enter sudo password (input hidden):"
$sec = Read-Host -AsSecureString
 $enc = $sec | ConvertFrom-SecureString
 $utf8NoBom = New-Object System.Text.UTF8Encoding($false)
 [System.IO.File]::WriteAllText($OutFile, $enc, $utf8NoBom)
Write-Host "Encrypted password written to $OutFile"
