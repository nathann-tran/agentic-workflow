$ErrorActionPreference = 'Stop'

$Repo = if ($env:AWM_REPO) { $env:AWM_REPO } else { 'nathann-tran/agentic-workflow' }
$Version = $env:AWM_VERSION
if ($Version) { $Version = $Version.TrimStart('v') }
$BaseUrl = $env:AWM_BASE_URL
$Temp = $null

function Get-ExpectedHash([string]$SumsPath, [string]$Name) {
    foreach ($Line in Get-Content $SumsPath) {
        $Parts = $Line -split '\s+', 2
        if ($Parts.Count -eq 2 -and $Parts[1].Trim() -eq $Name) { return $Parts[0] }
    }
    throw "No checksum for $Name in $SumsPath"
}

try {
    $Arch = [System.Runtime.InteropServices.RuntimeInformation]::OSArchitecture.ToString().ToLowerInvariant()
    if ($Arch -eq 'x64') { $Arch = 'amd64' }
    if ($Arch -ne 'amd64') { throw "Unsupported Windows architecture: $Arch" }

    $Root = if ($PSScriptRoot) { Split-Path -Parent $PSScriptRoot } else { $null }
    $Binary = $null
    $Verify = $null
    if ($Root) {
        $Candidate = Join-Path $Root "dist/awm-windows-$Arch.exe"
        if (Test-Path $Candidate) { $Binary = $Candidate; $Verify = Join-Path $Root 'dist/SHA256SUMS.txt' }
    }

    if (-not $Binary) {
        $Go = Get-Command go -ErrorAction SilentlyContinue
        if ($Root -and (Test-Path (Join-Path $Root 'go.mod')) -and $Go) {
            $BuildDir = Join-Path $Root '.build'
            New-Item -ItemType Directory -Force -Path $BuildDir | Out-Null
            $env:CGO_ENABLED = '0'
            $env:GOOS = 'windows'
            $env:GOARCH = $Arch
            Push-Location $Root
            try { go build -trimpath -o (Join-Path $BuildDir 'awm.exe') ./cmd/awm }
            finally { Pop-Location }
            $Binary = Join-Path $BuildDir 'awm.exe'
        } else {
            $Temp = Join-Path ([System.IO.Path]::GetTempPath()) ([guid]::NewGuid().ToString())
            New-Item -ItemType Directory -Path $Temp | Out-Null
            if (-not $BaseUrl) {
                if ($Version) {
                    $BaseUrl = "https://github.com/$Repo/releases/download/v$Version"
                } else {
                    $BaseUrl = "https://github.com/$Repo/releases/latest/download"
                }
            }
            if (-not $Version) {
                $Manifest = Join-Path $Temp 'RELEASE-MANIFEST.json'
                Invoke-WebRequest -UseBasicParsing -Uri "$BaseUrl/RELEASE-MANIFEST.json" -OutFile $Manifest
                $Version = (Get-Content -Raw $Manifest | ConvertFrom-Json).version
            }
            $ArchiveName = "awm_${Version}_windows_${Arch}.zip"
            $SumsPath = Join-Path $Temp 'SHA256SUMS.txt'
            $ArchivePath = Join-Path $Temp $ArchiveName
            Invoke-WebRequest -UseBasicParsing -Uri "$BaseUrl/SHA256SUMS.txt" -OutFile $SumsPath
            Invoke-WebRequest -UseBasicParsing -Uri "$BaseUrl/$ArchiveName" -OutFile $ArchivePath
            $Expected = Get-ExpectedHash $SumsPath $ArchiveName
            $Actual = (Get-FileHash -Algorithm SHA256 -Path $ArchivePath).Hash.ToLowerInvariant()
            if ($Actual -ne $Expected) { throw "Checksum mismatch for $ArchiveName" }
            Expand-Archive -LiteralPath $ArchivePath -DestinationPath $Temp -Force
            $Binary = Join-Path $Temp 'awm.exe'
        }
    }

    if ($Verify -and (Test-Path $Verify)) {
        $Name = Split-Path -Leaf $Binary
        $Expected = Get-ExpectedHash $Verify $Name
        $Actual = (Get-FileHash -Algorithm SHA256 -Path $Binary).Hash.ToLowerInvariant()
        if ($Actual -ne $Expected) { throw "Checksum mismatch for $Binary" }
    }

    $DestDir = Join-Path $HOME '.local/bin'
    $Dest = Join-Path $DestDir 'awm.exe'
    New-Item -ItemType Directory -Force -Path $DestDir | Out-Null
    if (Test-Path $Dest) {
        $Answer = Read-Host "Replace existing $Dest? [y/N]"
        if ($Answer -notin @('y', 'Y', 'yes', 'YES')) { throw 'Installation cancelled.' }
        $Stamp = [DateTime]::UtcNow.ToString('yyyyMMddTHHmmssZ')
        Copy-Item $Dest "$Dest.bak.$Stamp"
    }
    Copy-Item $Binary $Dest -Force
    Write-Host "Installed AWM at $Dest"
    if (($env:PATH -split ';') -notcontains $DestDir) {
        Write-Host "Add $DestDir to your user PATH."
    }
    Write-Host 'Run: awm install'
} finally {
    if ($Temp) { Remove-Item -Recurse -Force $Temp -ErrorAction SilentlyContinue }
}
