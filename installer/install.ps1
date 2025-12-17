# GlyphLang PowerShell Installer
# Usage: iwr -useb https://glyph-lang.github.io/install.ps1 | iex
#    or: Invoke-WebRequest -Uri https://glyph-lang.github.io/install.ps1 -UseBasicParsing | Invoke-Expression

$ErrorActionPreference = 'Stop'

# Configuration
$Repo = "glyph-lang/glyph"
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

    $filename = "glyph-$Platform.zip"
    return "https://github.com/$Repo/releases/download/v$Version/$filename"
}

function Install-GlyphLang {
    param([string]$DownloadUrl)

    Write-Info "Creating installation directory: $InstallDir"
    New-Item -ItemType Directory -Force -Path $BinDir | Out-Null

    $tempDir = New-Item -ItemType Directory -Force -Path "$env:TEMP\glyph-install-$(Get-Random)"
    $archivePath = "$tempDir\glyph.zip"

    try {
        Write-Info "Downloading GlyphLang..."
        [Net.ServicePointManager]::SecurityProtocol = [Net.SecurityProtocolType]::Tls12
        Invoke-WebRequest -Uri $DownloadUrl -OutFile $archivePath -UseBasicParsing

        Write-Info "Extracting..."
        Expand-Archive -Path $archivePath -DestinationPath $tempDir -Force

        # Find the binary
        $binary = Get-ChildItem -Path $tempDir -Filter "glyph*.exe" -Recurse | Select-Object -First 1
        if (-not $binary) {
            $binary = Get-ChildItem -Path $tempDir -Filter "glyph*" -Recurse | Where-Object { -not $_.PSIsContainer -and $_.Extension -ne ".zip" } | Select-Object -First 1
        }

        if (-not $binary) {
            Write-Error "Could not find glyph binary in archive"
        }

        Write-Info "Installing to $BinDir\glyph.exe..."
        Copy-Item -Path $binary.FullName -Destination "$BinDir\glyph.exe" -Force

        Write-Success "GlyphLang installed successfully!"
    } finally {
        Remove-Item -Path $tempDir -Recurse -Force -ErrorAction SilentlyContinue
    }
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
    Write-Host "  ██████╗ ██╗██████╗  █████╗ ██╗      █████╗ ███╗   ██╗ ██████╗ " -ForegroundColor Magenta
    Write-Host " ██╔══██╗██║██╔══██╗██╔══██╗██║     ██╔══██╗████╗  ██║██╔════╝ " -ForegroundColor Magenta
    Write-Host " ███████║██║██║  ██║███████║██║     ███████║██╔██╗ ██║██║  ███╗" -ForegroundColor Magenta
    Write-Host " ██╔══██║██║██║  ██║██╔══██║██║     ██╔══██║██║╚██╗██║██║   ██║" -ForegroundColor Magenta
    Write-Host " ██║  ██║██║██████╔╝██║  ██║███████╗██║  ██║██║ ╚████║╚██████╔╝" -ForegroundColor Magenta
    Write-Host " ╚═╝  ╚═╝╚═╝╚═════╝ ╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝╚═╝  ╚═══╝ ╚═════╝ " -ForegroundColor Magenta
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
