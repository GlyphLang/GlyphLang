#!/bin/bash
# Build Debian package for GlyphLang
# Run on Linux or in CI/CD

set -e

VERSION="${1:-1.0.0}"
ARCH="${2:-amd64}"

echo "Building GlyphLang .deb package v${VERSION} for ${ARCH}..."

# Setup directories
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
BUILD_DIR="$SCRIPT_DIR/build-deb"
PKG_DIR="$BUILD_DIR/glyph_${VERSION}_${ARCH}"

rm -rf "$BUILD_DIR"
mkdir -p "$PKG_DIR/DEBIAN"
mkdir -p "$PKG_DIR/usr/bin"
mkdir -p "$PKG_DIR/usr/share/doc/glyph"

# Copy control files
cp "$SCRIPT_DIR/deb/DEBIAN/control" "$PKG_DIR/DEBIAN/"
cp "$SCRIPT_DIR/deb/DEBIAN/postinst" "$PKG_DIR/DEBIAN/"
chmod 755 "$PKG_DIR/DEBIAN/postinst"

# Update version in control file
sed -i "s/^Version:.*/Version: ${VERSION}/" "$PKG_DIR/DEBIAN/control"
sed -i "s/^Architecture:.*/Architecture: ${ARCH}/" "$PKG_DIR/DEBIAN/control"

# Copy binary
if [ -f "$SCRIPT_DIR/../dist/glyph-linux-${ARCH}" ]; then
    cp "$SCRIPT_DIR/../dist/glyph-linux-${ARCH}" "$PKG_DIR/usr/bin/glyph"
elif [ -f "$SCRIPT_DIR/../dist/glyph-linux-amd64" ]; then
    cp "$SCRIPT_DIR/../dist/glyph-linux-amd64" "$PKG_DIR/usr/bin/glyph"
else
    echo "Error: Binary not found. Run 'make build-all' first."
    exit 1
fi
chmod 755 "$PKG_DIR/usr/bin/glyph"

# Add copyright/license
cat > "$PKG_DIR/usr/share/doc/glyph/copyright" << 'EOF'
Format: https://www.debian.org/doc/packaging-manuals/copyright-format/1.0/
Upstream-Name: glyph
Source: https://github.com/glyph-lang/glyph

Files: *
Copyright: 2025 GlyphLang Team
License: MIT
 Permission is hereby granted, free of charge, to any person obtaining a copy
 of this software and associated documentation files (the "Software"), to deal
 in the Software without restriction, including without limitation the rights
 to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 copies of the Software, and to permit persons to whom the Software is
 furnished to do so, subject to the following conditions:
 .
 The above copyright notice and this permission notice shall be included in all
 copies or substantial portions of the Software.
 .
 THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
 SOFTWARE.
EOF

# Build the package
dpkg-deb --build "$PKG_DIR"

# Move to dist
mv "$BUILD_DIR/glyph_${VERSION}_${ARCH}.deb" "$SCRIPT_DIR/../dist/"

echo "Created: dist/glyph_${VERSION}_${ARCH}.deb"

# Cleanup
rm -rf "$BUILD_DIR"
