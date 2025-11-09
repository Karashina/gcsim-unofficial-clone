param(
    [string]$Path
)

if (-not $Path) {
    Write-Error "Path parameter is required. Usage: .\stage_and_commit.ps1 -Path <path>"
    exit 2
}

Write-Output "Staging all changes under: $Path"
# Stage everything under the path
git add -A -- $Path

# List staged files for that path
$staged = git diff --name-only --cached -- $Path
if ($staged -and $staged -ne '') {
    Write-Output "Staged files (count: $($staged.Count)) :"
    Write-Output $staged
    # Commit
    $msg = "chore(integrate): automated import/API pass for $Path"
    git commit -m $msg
    if ($LASTEXITCODE -eq 0) {
        Write-Output "Commit successful. New HEAD: $(git rev-parse --short HEAD)"
    }
    else {
        Write-Error "git commit failed with exit code $LASTEXITCODE"
        exit $LASTEXITCODE
    }
}
else {
    Write-Output "No staged changes under $Path to commit."
}
