$dirs = @(
  'internal/characters/aino',
  'internal/characters/albedo',
  'internal/characters/alhaitham',
  'internal/characters/aloy',
  'internal/characters/amber',
  'internal/characters/arlecchino',
  'internal/characters/ayaka',
  'internal/characters/ayato',
  'internal/characters/baizhu',
  'internal/characters/barbara'
)
$changed = @()
foreach ($d in $dirs) {
  if (-not (Test-Path $d)) { continue }
  Get-ChildItem -Path $d -Filter *.go -Recurse | ForEach-Object {
    $p = $_.FullName
    $text = Get-Content -Raw -LiteralPath $p -Encoding UTF8
    if ($text -match '\binfo\.' -and $text -notmatch '"github.com/genshinsim/gcsim/pkg/core/info"') {
      if ($text -match 'import\s*\(') {
        $text = $text -replace '(import\s*\()','${1}\n\t"github.com/genshinsim/gcsim/pkg/core/info"'
      } else {
        # try to add single import line after package decl
        $text = $text -replace '(package\s+[\w_]+\r?\n)','$1\nimport "github.com/genshinsim/gcsim/pkg/core/info"\n'
      }
      Set-Content -LiteralPath $p -Value $text -Encoding UTF8
      $changed += $p
    }
  }
}
if ($changed.Count -gt 0) { Write-Host "Added info import to:`n"; $changed | ForEach-Object { Write-Host $_ } } else { Write-Host 'No files needed info import' }
