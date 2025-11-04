param(
  [string]$Root = "internal/characters"
)

# Generic pass: normalize imports, replace combat.* -> info.*, fix .Index usages for all .go files under $Root
Write-Host "Running automated fix pass on: $Root"

$files = Get-ChildItem -Path $Root -Filter *.go -Recurse -ErrorAction SilentlyContinue | ForEach-Object { $_.FullName }
if ($files.Count -eq 0) { Write-Host "No .go files under $Root"; exit 0 }

$changedFiles = @()

function Normalize-ImportBlock([string]$text) {
    if ($text -match "import\s*\(") {
        $pre, $impBlock, $post = $text -split '(?s)(import\s*\([^)]*\))', 3
        if (-not $impBlock) { return $text }
        $inner = $impBlock -replace '^import\s*\(', '' -replace '\)\s*$', ''
        $lines = $inner -split "\r?\n" | ForEach-Object { $_.Trim() } | Where-Object { $_ -ne '' }
        $seen = @{}
        $uniq = New-Object System.Collections.ArrayList
        foreach ($l in $lines) {
            if (-not $seen.ContainsKey($l)) { $seen[$l] = $true; [void]$uniq.Add($l) }
        }
        $newInner = $uniq -join "`n"
        $newImp = "import(`n`t$newInner`n)"
        return ($pre + $newImp + $post)
    } else {
        return $text
    }
}

foreach ($path in $files) {
    $orig = Get-Content -Raw -LiteralPath $path -Encoding UTF8
    $text = $orig

    # Replace combat.* common types -> info.*
    $text = [regex]::Replace($text, '\bcombat\.(AttackInfo|Snapshot|AttackEvent|AttackCB|AttackCBFunc|AttackPattern|Target)\b', 'info.$1')

    # Replace .Index -> .Index() when not already called
    $lines = $text -split "\r?\n"
    for ($i = 0; $i -lt $lines.Count; $i++) {
        $l = $lines[$i]
        if ($l -match '^\s*func\b') { continue }
        $lines[$i] = [regex]::Replace($l, '(\w+)\.Index\b(?!\()', '$1.Index()')
    }
    $text = $lines -join "`n"

    $text = Normalize-ImportBlock $text

    if ($text -ne $orig) {
        Set-Content -LiteralPath $path -Value $text -Encoding UTF8
        $changedFiles += $path
    }
}

if ($changedFiles.Count -gt 0) {
    Write-Host "Files changed:`n"; $changedFiles | ForEach-Object { Write-Host $_ }
    Write-Host "Running gofmt on changed files..."
    foreach ($f in $changedFiles) { & gofmt -w $f }
} else {
    Write-Host "No files changed by the pass"
}

Write-Host "Running: go build ./..."
& go build ./...
$exit=$LASTEXITCODE
if ($exit -ne 0) { Write-Host "go build exit code: $exit"; exit $exit }
Write-Host "Build succeeded"
