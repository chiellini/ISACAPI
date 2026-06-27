param(
    [Parameter(Mandatory = $true)]
    [string]$Remote,

    [string]$RemotePostgresContainer = "sub2api-postgres",
    [string]$RemoteDbUser = "sub2api",
    [string]$RemoteDb = "sub2api",
    [string]$RemoteDbHost = "127.0.0.1",
    [int]$RemoteDbPort = 5432,
    [string]$RemotePgPassword = "",

    [string]$LocalDb = "",
    [string]$LocalDbUser = "",

    [switch]$Force
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

function Quote-SqlIdentifier {
    param([string]$Value)
    return '"' + $Value.Replace('"', '""') + '"'
}

function Quote-SqlLiteral {
    param([string]$Value)
    return "'" + $Value.Replace("'", "''") + "'"
}

function Quote-Sh {
    param([string]$Value)

    $single = [string][char]39
    $double = [string][char]34
    $escapedSingle = $single + $double + $single + $double + $single
    return $single + $Value.Replace($single, $escapedSingle) + $single
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

if ([string]::IsNullOrWhiteSpace($LocalDb)) {
    $LocalDb = Get-EnvValue "POSTGRES_DB"
    if ([string]::IsNullOrWhiteSpace($LocalDb)) {
        $LocalDb = "sub2api"
    }
}

if ([string]::IsNullOrWhiteSpace($LocalDbUser)) {
    $LocalDbUser = Get-EnvValue "POSTGRES_USER"
    if ([string]::IsNullOrWhiteSpace($LocalDbUser)) {
        $LocalDbUser = "sub2api"
    }
}

foreach ($dir in @("data", "postgres_data", "redis_data")) {
    $path = Join-Path $ScriptDir $dir
    if (-not (Test-Path $path)) {
        New-Item -ItemType Directory -Path $path | Out-Null
    }
}

docker --version | Out-Null
docker compose version | Out-Null
ssh -V 2>$null

if (-not $Force) {
    Write-Host ""
    Write-Host "This will replace the LOCAL Docker PostgreSQL database '$LocalDb'."
    Write-Host "Remote source: $Remote / database '$RemoteDb'"
    $confirmation = Read-Host "Type SYNC to continue"
    if ($confirmation -ne "SYNC") {
        throw "Cancelled."
    }
}

Write-Host "Starting local PostgreSQL and Redis..."
docker compose -f docker-compose.local.yml up -d postgres redis
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}

Write-Host "Waiting for local PostgreSQL..."
$ready = $false
for ($i = 0; $i -lt 60; $i++) {
    docker compose -f docker-compose.local.yml exec -T postgres pg_isready -U $LocalDbUser -d $LocalDb | Out-Null
    if ($LASTEXITCODE -eq 0) {
        $ready = $true
        break
    }

    Start-Sleep -Seconds 2
}

if (-not $ready) {
    throw "Local PostgreSQL did not become ready."
}

docker compose -f docker-compose.local.yml -f docker-compose.build.yml stop sub2api 2>$null | Out-Null

$quotedDbIdent = Quote-SqlIdentifier $LocalDb
$quotedUserIdent = Quote-SqlIdentifier $LocalDbUser
$quotedDbLiteral = Quote-SqlLiteral $LocalDb

Write-Host "Recreating local database '$LocalDb'..."
docker compose -f docker-compose.local.yml exec -T postgres psql -v ON_ERROR_STOP=1 -U $LocalDbUser -d postgres -c "SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE datname = $quotedDbLiteral AND pid <> pg_backend_pid();"
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}

docker compose -f docker-compose.local.yml exec -T postgres psql -v ON_ERROR_STOP=1 -U $LocalDbUser -d postgres -c "DROP DATABASE IF EXISTS $quotedDbIdent;"
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}

docker compose -f docker-compose.local.yml exec -T postgres psql -v ON_ERROR_STOP=1 -U $LocalDbUser -d postgres -c "CREATE DATABASE $quotedDbIdent OWNER $quotedUserIdent;"
if ($LASTEXITCODE -ne 0) {
    exit $LASTEXITCODE
}

if ([string]::IsNullOrWhiteSpace($RemotePostgresContainer)) {
    $passwordPrefix = ""
    if (-not [string]::IsNullOrEmpty($RemotePgPassword)) {
        $passwordPrefix = "PGPASSWORD=" + (Quote-Sh $RemotePgPassword) + " "
    }

    $remoteDumpCommand = $passwordPrefix +
        "pg_dump -h " + (Quote-Sh $RemoteDbHost) +
        " -p " + (Quote-Sh ([string]$RemoteDbPort)) +
        " -U " + (Quote-Sh $RemoteDbUser) +
        " -d " + (Quote-Sh $RemoteDb) +
        " --clean --if-exists --no-owner --no-acl"
}
else {
    $remoteDumpCommand =
        "docker exec -i " + (Quote-Sh $RemotePostgresContainer) +
        " pg_dump -U " + (Quote-Sh $RemoteDbUser) +
        " -d " + (Quote-Sh $RemoteDb) +
        " --clean --if-exists --no-owner --no-acl"
}

Write-Host "Streaming remote database into local Docker PostgreSQL..."
$restoreArgs = @(
    "compose",
    "-f", "docker-compose.local.yml",
    "exec", "-T", "postgres",
    "psql",
    "-v", "ON_ERROR_STOP=1",
    "-U", $LocalDbUser,
    "-d", $LocalDb
)

& ssh $Remote $remoteDumpCommand | & docker @restoreArgs
$restoreSucceeded = $?
if (-not $restoreSucceeded) {
    throw "Database restore failed."
}

Write-Host ""
Write-Host "Database sync complete."
Write-Host "Next: .\\win-build-deploy.ps1"
