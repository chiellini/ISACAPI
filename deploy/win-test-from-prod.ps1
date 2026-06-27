param(
    [switch]$SkipSync,
    [switch]$Logs
)

$ErrorActionPreference = "Stop"

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $ScriptDir

$remote = "ec2-user@100.31.110.106"

if (-not $SkipSync) {
    & (Join-Path $ScriptDir "win-sync-db.ps1") -Remote $remote -Force
    if ($LASTEXITCODE -ne 0) {
        exit $LASTEXITCODE
    }
}

if ($Logs) {
    & (Join-Path $ScriptDir "win-build-deploy.ps1") -Logs
}
else {
    & (Join-Path $ScriptDir "win-build-deploy.ps1")
}

if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}
