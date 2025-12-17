#!/bin/bash
# GlyphLang Installer Script
# Usage: curl -fsSL https://glyph-lang.github.io/install.sh | bash
#    or: wget -qO- https://glyph-lang.github.io/install.sh | bash

set -e

# Configuration
REPO="glyph-lang/glyph"
INSTALL_DIR="${GLYPH_INSTALL_DIR:-$HOME/.glyph}"
BIN_DIR="${Glyph_BIN_DIR:-$INSTALL_DIR/bin}"
VERSION="${Glyph_VERSION:-latest}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[0;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

info() { echo -e "${BLUE}[INFO]${NC} $1"; }
success() { echo -e "${GREEN}[SUCCESS]${NC} $1"; }
warn() { echo -e "${YELLOW}[WARNING]${NC} $1"; }
error() { echo -e "${RED}[ERROR]${NC} $1"; exit 1; }

# Detect OS and architecture
detect_platform() {
    OS="$(uname -s)"
    ARCH="$(uname -m)"

    case "$OS" in
        Linux*)  OS="linux" ;;
        Darwin*) OS="darwin" ;;
        *)       error "Unsupported operating system: $OS" ;;
    esac

    case "$ARCH" in
        x86_64|amd64)  ARCH="amd64" ;;
        arm64|aarch64) ARCH="arm64" ;;
        *)             error "Unsupported architecture: $ARCH" ;;
    esac

    # Linux arm64 is now supported
    if [ "$OS" = "linux" ] && [ "$ARCH" = "arm64" ]; then
        info "Linux ARM64 detected"
    fi

    PLATFORM="${OS}-${ARCH}"
    info "Detected platform: $PLATFORM"
}

# Get download URL
get_download_url() {
    if [ "$VERSION" = "latest" ]; then
        # Get latest release from GitHub API
        RELEASE_URL="https://api.github.com/repos/${REPO}/releases/latest"
        if command -v curl &> /dev/null; then
            VERSION=$(curl -fsSL "$RELEASE_URL" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/' | sed 's/^v//')
        elif command -v wget &> /dev/null; then
            VERSION=$(wget -qO- "$RELEASE_URL" | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/' | sed 's/^v//')
        else
            error "Neither curl nor wget found. Please install one of them."
        fi

        if [ -z "$VERSION" ]; then
            # Fallback to hardcoded version
            VERSION="1.0.0"
            warn "Could not fetch latest version, using $VERSION"
        fi
    fi

    info "Installing GlyphLang version: $VERSION"

    # Construct download URL
    FILENAME="glyph-${PLATFORM}"
    if [ "$OS" = "darwin" ]; then
        DOWNLOAD_URL="https://github.com/${REPO}/releases/download/v${VERSION}/${FILENAME}.zip"
    else
        DOWNLOAD_URL="https://github.com/${REPO}/releases/download/v${VERSION}/${FILENAME}.zip"
    fi
}

# Download and install
install() {
    info "Creating installation directory: $INSTALL_DIR"
    mkdir -p "$BIN_DIR"

    TEMP_DIR=$(mktemp -d)
    trap "rm -rf $TEMP_DIR" EXIT

    info "Downloading GlyphLang..."
    ARCHIVE_PATH="$TEMP_DIR/glyph.zip"

    if command -v curl &> /dev/null; then
        curl -fsSL "$DOWNLOAD_URL" -o "$ARCHIVE_PATH" || error "Download failed. Check if the release exists."
    elif command -v wget &> /dev/null; then
        wget -q "$DOWNLOAD_URL" -O "$ARCHIVE_PATH" || error "Download failed. Check if the release exists."
    fi

    info "Extracting..."
    if command -v unzip &> /dev/null; then
        unzip -q "$ARCHIVE_PATH" -d "$TEMP_DIR"
    else
        error "unzip not found. Please install unzip."
    fi

    # Find the binary
    BINARY=$(find "$TEMP_DIR" -name "glyph*" -type f ! -name "*.zip" | head -1)
    if [ -z "$BINARY" ]; then
        error "Could not find glyph binary in archive"
    fi

    info "Installing to $BIN_DIR/glyph..."
    mv "$BINARY" "$BIN_DIR/glyph"
    chmod +x "$BIN_DIR/glyph"

    success "GlyphLang installed successfully!"
}

# Setup PATH
setup_path() {
    SHELL_NAME=$(basename "$SHELL")
    PROFILE=""

    case "$SHELL_NAME" in
        bash)
            if [ -f "$HOME/.bashrc" ]; then
                PROFILE="$HOME/.bashrc"
            elif [ -f "$HOME/.bash_profile" ]; then
                PROFILE="$HOME/.bash_profile"
            fi
            ;;
        zsh)
            PROFILE="$HOME/.zshrc"
            ;;
        fish)
            PROFILE="$HOME/.config/fish/config.fish"
            ;;
    esac

    # Check if already in PATH
    if echo "$PATH" | grep -q "$BIN_DIR"; then
        info "GlyphLang is already in PATH"
        return
    fi

    if [ -n "$PROFILE" ] && [ -f "$PROFILE" ]; then
        # Check if export already exists
        if grep -q "Glyph" "$PROFILE" 2>/dev/null; then
            info "PATH entry already exists in $PROFILE"
        else
            echo "" >> "$PROFILE"
            echo "# GlyphLang" >> "$PROFILE"
            if [ "$SHELL_NAME" = "fish" ]; then
                echo "set -gx PATH \"$BIN_DIR\" \$PATH" >> "$PROFILE"
            else
                echo "export PATH=\"$BIN_DIR:\$PATH\"" >> "$PROFILE"
            fi
            success "Added GlyphLang to PATH in $PROFILE"
        fi
    else
        warn "Could not detect shell profile. Add this to your shell config:"
        echo "  export PATH=\"$BIN_DIR:\$PATH\""
    fi
}

# Verify installation
verify() {
    if [ -x "$BIN_DIR/glyph" ]; then
        success "Installation verified!"
        echo ""
        echo "To get started, restart your terminal or run:"
        echo "  export PATH=\"$BIN_DIR:\$PATH\""
        echo ""
        echo "Then try:"
        echo "  glyph --version"
        echo "  glyph --help"
        echo ""
    else
        error "Installation verification failed"
    fi
}

# Main
main() {
    echo ""
    echo "  █████╗ ██╗██████╗  █████╗ ██╗      █████╗ ███╗   ██╗ ██████╗ "
    echo " ██╔══██╗██║██╔══██╗██╔══██╗██║     ██╔══██╗████╗  ██║██╔════╝ "
    echo " ███████║██║██║  ██║███████║██║     ███████║██╔██╗ ██║██║  ███╗"
    echo " ██╔══██║██║██║  ██║██╔══██║██║     ██╔══██║██║╚██╗██║██║   ██║"
    echo " ██║  ██║██║██████╔╝██║  ██║███████╗██║  ██║██║ ╚████║╚██████╔╝"
    echo " ╚═╝  ╚═╝╚═╝╚═════╝ ╚═╝  ╚═╝╚══════╝╚═╝  ╚═╝╚═╝  ╚═══╝ ╚═════╝ "
    echo ""
    echo "  AI-First Backend Language Installer"
    echo ""

    detect_platform
    get_download_url
    install
    setup_path
    verify
}

main "$@"
