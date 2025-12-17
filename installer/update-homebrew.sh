#!/bin/bash
# Update Homebrew formula with SHA256 hashes after a release
# Usage: ./update-homebrew.sh <version>

set -e

VERSION="${1:-1.0.0}"
REPO="glyphlang/glyph"
FORMULA="homebrew/glyph.rb"

echo "Updating Homebrew formula for version $VERSION..."

# Function to get SHA256 for a release asset
get_sha256() {
    local asset="$1"
    local url="https://github.com/${REPO}/releases/download/v${VERSION}/${asset}"
    echo "Fetching SHA256 for $asset..." >&2
    curl -fsSL "$url" | sha256sum | cut -d' ' -f1
}

# Get all SHA256 hashes
echo "Downloading and hashing release assets..."
SHA_DARWIN_ARM64=$(get_sha256 "glyph-darwin-arm64.zip")
SHA_DARWIN_AMD64=$(get_sha256 "glyph-darwin-amd64.zip")
SHA_LINUX_ARM64=$(get_sha256 "glyph-linux-arm64.zip")
SHA_LINUX_AMD64=$(get_sha256 "glyph-linux-amd64.zip")

echo ""
echo "SHA256 Hashes:"
echo "  darwin-arm64: $SHA_DARWIN_ARM64"
echo "  darwin-amd64: $SHA_DARWIN_AMD64"
echo "  linux-arm64:  $SHA_LINUX_ARM64"
echo "  linux-amd64:  $SHA_LINUX_AMD64"
echo ""

# Update formula
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
sed -i "s/PLACEHOLDER_ARM64_SHA256/$SHA_DARWIN_ARM64/" "$SCRIPT_DIR/$FORMULA"
sed -i "s/PLACEHOLDER_AMD64_SHA256/$SHA_DARWIN_AMD64/" "$SCRIPT_DIR/$FORMULA"
sed -i "s/PLACEHOLDER_LINUX_ARM64_SHA256/$SHA_LINUX_ARM64/" "$SCRIPT_DIR/$FORMULA"
sed -i "s/PLACEHOLDER_LINUX_AMD64_SHA256/$SHA_LINUX_AMD64/" "$SCRIPT_DIR/$FORMULA"
sed -i "s/version \"[^\"]*\"/version \"$VERSION\"/" "$SCRIPT_DIR/$FORMULA"
sed -i "s|/v[0-9.]*|/v$VERSION|g" "$SCRIPT_DIR/$FORMULA"

echo "Updated $FORMULA with version $VERSION"
echo ""
echo "Next steps:"
echo "1. Review the changes: git diff $FORMULA"
echo "2. Commit: git add $FORMULA && git commit -m 'Update Homebrew formula to v$VERSION'"
echo "3. Push to tap repository"
