param(
    [switch]$Logs
)

$ErrorActionPreference = "Stop"

$ScriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
Set-Location $ScriptDir

$EnvPath = Join-Path $ScriptDir ".env"
$EnvExamplePath = Join-Path $ScriptDir ".env.example"

function New-HexSecret {
    param([int]$Bytes = 32)

    $bytes = New-Object byte[] $Bytes
    $rng = [System.Security.Cryptography.RandomNumberGenerator]::Create()
    try {
        $rng.GetBytes($bytes)
    }
    finally {
        $rng.Dispose()
    }

    return -join ($bytes | ForEach-Object { $_.ToString("x2") })
}

function Get-EnvValue {
    param([string]$Key)

    if (-not (Test-Path $EnvPath)) {
        return $null
    }

    foreach ($line in [System.IO.File]::ReadAllLines($EnvPath)) {
        if ($line -match "^$([regex]::Escape($Key))=(.*)$") {
            return $matches[1]
        }
    }

    return $null
}

function Set-EnvValue {
    param(
        [string]$Key,
        [string]$Value
    )

    $lines = New-Object System.Collections.Generic.List[string]
    if (Test-Path $EnvPath) {
        $lines.AddRange([string[]][System.IO.File]::ReadAllLines($EnvPath))
    }

    $found = $false
    for ($i = 0; $i -lt $lines.Count; $i++) {
        if ($lines[$i] -match "^$([regex]::Escape($Key))=") {
            $lines[$i] = "$Key=$Value"
            $found = $true
            break
        }
    }

    if (-not $found) {
        $lines.Add("$Key=$Value")
    }

    [System.IO.File]::WriteAllLines($EnvPath, [string[]]$lines)
}

function Ensure-EnvSecret {
    param(
        [string]$Key,
        [string[]]$BadValues = @("")
    )

    $value = Get-EnvValue $Key
    if ($null -eq $value -or $BadValues -contains $value) {
        $value = New-HexSecret 32
        Set-EnvValue $Key $value
        Write-Host "Generated $Key"
    }
}

if (-not (Test-Path $EnvPath)) {
    if (-not (Test-Path $EnvExamplePath)) {
        throw ".env.example not found in deploy directory."
    }

    Copy-Item -LiteralPath $EnvExamplePath -Destination $EnvPath
    Write-Host "Created deploy\\.env from .env.example"
}

Ensure-EnvSecret -Key "POSTGRES_PASSWORD" -BadValues @("", "change_this_secure_password")
Ensure-EnvSecret -Key "JWT_SECRET" -BadValues @("")
Ensure-EnvSecret -Key "TOTP_ENCRYPTION_KEY" -BadValues @("")

$serverGuard = Get-EnvValue "SERVER_REGISTER_GUARD_TOKEN"
if ([string]::IsNullOrWhiteSpace($serverGuard) -or $serverGuard -eq "s2a-rg-7Kq2xZ9m") {
    $serverGuard = "rg-" + (New-HexSecret 24)
    Set-EnvValue "SERVER_REGISTER_GUARD_TOKEN" $serverGuard
    Write-Host "Generated SERVER_REGISTER_GUARD_TOKEN"
}

$viteGuard = Get-EnvValue "VITE_REG_GUARD_TOKEN"
if ($viteGuard -ne $serverGuard) {
    Set-EnvValue "VITE_REG_GUARD_TOKEN" $serverGuard
    Write-Host "Synced VITE_REG_GUARD_TOKEN with SERVER_REGISTER_GUARD_TOKEN"
}

foreach ($dir in @("data", "postgres_data", "redis_data")) {
    $path = Join-Path $ScriptDir $dir
    if (-not (Test-Path $path)) {
        New-Item -ItemType Directory -Path $path | Out-Null
    }
}

docker --version | Out-Null
docker compose version | Out-Null

$composeFiles = @("docker-compose.local.yml", "docker-compose.build.yml")
foreach ($file in $composeFiles) {
    if (-not (Test-Path (Join-Path $ScriptDir $file))) {
        throw "$file not found in deploy directory."
    }
}

Write-Host "Building local image and starting ISACAPI..."
docker compose -f docker-compose.local.yml -f docker-compose.build.yml up -d --build
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}

$port = Get-EnvValue "SERVER_PORT"
if ([string]::IsNullOrWhiteSpace($port)) {
    $port = "8080"
}

Write-Host ""
Write-Host "ISACAPI is starting: http://localhost:$port"
Write-Host "Status: docker compose -f docker-compose.local.yml -f docker-compose.build.yml ps"
Write-Host "Logs:   docker compose -f docker-compose.local.yml -f docker-compose.build.yml logs -f sub2api"

if ($Logs) {
    docker compose -f docker-compose.local.yml -f docker-compose.build.yml logs -f sub2api
}
