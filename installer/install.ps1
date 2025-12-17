# GlyphLang PowerShell Installer
# Usage: iwr -useb https://glyph-lang.github.io/install.ps1 | iex
#    or: Invoke-WebRequest -Uri https://glyph-lang.github.io/install.ps1 -UseBasicParsing | Invoke-Expression

$ErrorActionPreference = 'Stop'

# Configuration
$Repo = "GlyphLang/GlyphLang"
$InstallDir = if ($env:Glyph_INSTALL_DIR) { $env:Glyph_INSTALL_DIR } else { "$env:LOCALAPPDATA\GlyphLang" }
$BinDir = if ($env:Glyph_BIN_DIR) { $env:Glyph_BIN_DIR } else { "$InstallDir\bin" }
$Version = if ($env:Glyph_VERSION) { $env:Glyph_VERSION } else { "latest" }

function Write-Info { param([string]$Message) Write-Host "[INFO] " -ForegroundColor Blue -NoNewline; Write-Host $Message }
function Write-Success { param([string]$Message) Write-Host "[SUCCESS] " -ForegroundColor Green -NoNewline; Write-Host $Message }
function Write-Warn { param([string]$Message) Write-Host "[WARNING] " -ForegroundColor Yellow -NoNewline; Write-Host $Message }
function Write-Error { param([string]$Message) Write-Host "[ERROR] " -ForegroundColor Red -NoNewline; Write-Host $Message; exit 1 }

function Get-Platform {
    $arch = if ([Environment]::Is64BitOperatingSystem) { "amd64" } else { "386" }

    if ($arch -eq "386") {
        Write-Error "32-bit Windows is not supported. Please use 64-bit Windows."
    }

    Write-Info "Detected platform: windows-$arch"
    return "windows-$arch"
}

function Get-DownloadUrl {
    param([string]$Platform)

    if ($Version -eq "latest") {
        $releaseUrl = "https://api.github.com/repos/$Repo/releases/latest"
        try {
            $release = Invoke-RestMethod -Uri $releaseUrl -UseBasicParsing
            $script:Version = $release.tag_name -replace '^v', ''
        } catch {
            $script:Version = "1.0.0"
            Write-Warn "Could not fetch latest version, using $Version"
        }
    }

    Write-Info "Installing GlyphLang version: $Version"

    # Raw binary (not zipped)
    $filename = "glyph-$Platform.exe"
    return "https://github.com/$Repo/releases/download/v$Version/$filename"
}

function Install-GlyphLang {
    param([string]$DownloadUrl)

    Write-Info "Creating installation directory: $InstallDir"
    New-Item -ItemType Directory -Force -Path $BinDir | Out-Null

    $binaryPath = "$BinDir\glyph.exe"

    Write-Info "Downloading GlyphLang..."
    [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12 -bor [Net.SecurityProtocolType]::Tls13
    $ProgressPreference = 'SilentlyContinue'

    try {
        # Use HttpClient with redirect handling for GitHub releases
        Add-Type -AssemblyName System.Net.Http
        $handler = New-Object System.Net.Http.HttpClientHandler
        $handler.AllowAutoRedirect = $true
        $client = New-Object System.Net.Http.HttpClient($handler)
        $client.DefaultRequestHeaders.Add("User-Agent", "GlyphLang-Installer")

        $response = $client.GetAsync($DownloadUrl).Result
        if ($response.IsSuccessStatusCode) {
            $bytes = $response.Content.ReadAsByteArrayAsync().Result
            [System.IO.File]::WriteAllBytes($binaryPath, $bytes)
        } else {
            throw "HTTP $($response.StatusCode)"
        }
        $client.Dispose()
    } catch {
        Write-Error "Download failed: $_"
    }

    if (-not (Test-Path $binaryPath)) {
        Write-Error "Download failed. Binary not found at $binaryPath"
    }

    $fileSize = (Get-Item $binaryPath).Length / 1MB
    Write-Info "Installed to $BinDir\glyph.exe ($([math]::Round($fileSize, 1)) MB)"
    Write-Success "GlyphLang installed successfully!"
}

function Add-ToPath {
    $currentPath = [Environment]::GetEnvironmentVariable("Path", "User")

    if ($currentPath -like "*$BinDir*") {
        Write-Info "GlyphLang is already in PATH"
        return
    }

    Write-Info "Adding GlyphLang to PATH..."
    $newPath = "$BinDir;$currentPath"
    [Environment]::SetEnvironmentVariable("Path", $newPath, "User")
    $env:Path = "$BinDir;$env:Path"
    Write-Success "Added GlyphLang to PATH"
}

function Test-Installation {
    $glyphPath = "$BinDir\glyph.exe"

    if (Test-Path $glyphPath) {
        Write-Success "Installation verified!"
        Write-Host ""
        Write-Host "To get started, open a new terminal and run:" -ForegroundColor Cyan
        Write-Host "  glyph --version"
        Write-Host "  glyph --help"
        Write-Host ""
        Write-Host "Or run now with the full path:" -ForegroundColor Cyan
        Write-Host "  $glyphPath --version"
        Write-Host ""
    } else {
        Write-Error "Installation verification failed"
    }
}

function Show-Banner {
    Write-Host ""
    Write-Host "   _____ _             _     _                        " -ForegroundColor Magenta
    Write-Host "  / ____| |           | |   | |                       " -ForegroundColor Magenta
    Write-Host " | |  __| |_   _ _ __ | |__ | |     __ _ _ __   __ _  " -ForegroundColor Magenta
    Write-Host " | | |_ | | | | | '_ \| '_ \| |    / _' | '_ \ / _' | " -ForegroundColor Magenta
    Write-Host " | |__| | | |_| | |_) | | | | |___| (_| | | | | (_| | " -ForegroundColor Magenta
    Write-Host "  \_____|_|\__, | .__/|_| |_|______\__,_|_| |_|\__, | " -ForegroundColor Magenta
    Write-Host "            __/ | |                             __/ | " -ForegroundColor Magenta
    Write-Host "           |___/|_|                            |___/  " -ForegroundColor Magenta
    Write-Host ""
    Write-Host "  AI-First Backend Language Installer" -ForegroundColor Cyan
    Write-Host ""
}

# Main
Show-Banner
$platform = Get-Platform
$downloadUrl = Get-DownloadUrl -Platform $platform
Install-GlyphLang -DownloadUrl $downloadUrl
Add-ToPath
Test-Installation
