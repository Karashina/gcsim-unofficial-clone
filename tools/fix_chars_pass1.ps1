# Fix pass 1: normalize imports, replace c.Index -> c.Index(), convert common combat.* -> info.*
param()

$dirs = Get-ChildItem -Path 'internal/characters' -Directory -Recurse | ForEach-Object {
    # Only include directories that contain at least one .go file (leaf or nested)
    $dir = $_.FullName
    $hasGo = (Get-ChildItem -Path $dir -Filter *.go -Recurse -ErrorAction SilentlyContinue).Count -gt 0
    if ($hasGo) { $dir }
}

$changedFiles = @()

function Convert-ImportBlock([string]$text) {
    if ($text -match "import\s*\(") {
        # Extract import block
        $pre, $impBlock, $post = $text -split '(?s)(import\s*\([^)]*\))', 3
        if (-not $impBlock) { return $text }
        # get lines inside parentheses
        $inner = $impBlock -replace '^import\s*\(', '' -replace '\)\s*$', ''
        $lines = $inner -split "\r?\n" | ForEach-Object { $_.Trim() } | Where-Object { $_ -ne '' }
        # unique and preserve order
        $seen = @{}
        $uniq = New-Object System.Collections.ArrayList
        foreach ($l in $lines) {
            if (-not $seen.ContainsKey($l)) { $seen[$l] = $true; [void]$uniq.Add($l) }
        }
        $newInner = $uniq -join "`n"
        $newImp = "import(`n`t$newInner`n)"
        return ($pre + $newImp + $post)
    }
    else {
        return $text
    }
}

foreach ($d in $dirs) {
    if (-not (Test-Path $d)) { Write-Host "Skipping missing dir: $d"; continue }
    Get-ChildItem -Path $d -Filter *.go -Recurse | ForEach-Object {
        $path = $_.FullName
        $orig = Get-Content -Raw -LiteralPath $path -Encoding UTF8
        $text = $orig

        # Replace combat.* common types -> info.*
        $text = [regex]::Replace($text, '\bcombat\.(AttackInfo|Snapshot|AttackEvent|AttackCB|AttackCBFunc|AttackPattern|Target)\b', 'info.$1')

        # Replace GadgetTypReactableweapon legacy name if present
        $text = $text -replace '\bGadgetTypReactableweapon\b', 'GadgetTypReactableweapon'

        # Replace .Index (like c.Index) -> .Index() when not already called
        # Avoid replacing function declarations: simple heuristic - ignore lines with "func (" or "func "
        $lines = $text -split "\r?\n"
        for ($i = 0; $i -lt $lines.Count; $i++) {
            $l = $lines[$i]
            if ($l -match '^\s*func\b') { continue }
            # replace patterns like ' c.Index ' or 'c.Index,' or 'c.Index)' etc when not followed by (
            $lines[$i] = [regex]::Replace($l, '(\w+)\.Index\b(?!\()', '$1.Index()')
        }
        $text = $lines -join "`n"

        # Normalize import block (dedupe imports)
        $text = Convert-ImportBlock $text

        # If changed, write back
        if ($text -ne $orig) {
            Set-Content -LiteralPath $path -Value $text -Encoding UTF8
            $changedFiles += $path
        }
    }
}

if ($changedFiles.Count -gt 0) {
    Write-Host "Files changed:`n"; $changedFiles | ForEach-Object { Write-Host $_ }
    Write-Host "Running gofmt on changed files..."
    foreach ($f in $changedFiles) { & gofmt -w $f }
}
else {
    Write-Host "No files changed by the pass"
}

Write-Host "Running: go build ./..."
& go build ./...
$exit = $LASTEXITCODE
if ($exit -ne 0) { Write-Host "go build exit code: $exit"; exit $exit }
Write-Host "Build succeeded"
