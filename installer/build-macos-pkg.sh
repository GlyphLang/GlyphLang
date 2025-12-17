#!/bin/bash
# Build macOS .pkg installer for GlyphLang
# Must be run on macOS

set -e

VERSION="${1:-1.0.0}"
ARCH="${2:-$(uname -m)}"

# Map architecture names
case "$ARCH" in
    x86_64) ARCH_NAME="amd64" ;;
    arm64)  ARCH_NAME="arm64" ;;
    amd64)  ARCH_NAME="amd64" ;;
    *)      echo "Unknown architecture: $ARCH"; exit 1 ;;
esac

echo "Building GlyphLang .pkg installer v${VERSION} for ${ARCH_NAME}..."

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DIST_DIR="$SCRIPT_DIR/../dist"
BUILD_DIR="$SCRIPT_DIR/build-pkg"
PKG_ROOT="$BUILD_DIR/root"
PKG_SCRIPTS="$BUILD_DIR/scripts"

# Cleanup
rm -rf "$BUILD_DIR"
mkdir -p "$PKG_ROOT/usr/local/bin"
mkdir -p "$PKG_SCRIPTS"

# Copy binary
BINARY="$DIST_DIR/glyph-darwin-${ARCH_NAME}"
if [ ! -f "$BINARY" ]; then
    echo "Error: Binary not found at $BINARY"
    echo "Run 'make build-all' first."
    exit 1
fi

cp "$BINARY" "$PKG_ROOT/usr/local/bin/glyph"
chmod 755 "$PKG_ROOT/usr/local/bin/glyph"

# Create postinstall script
cat > "$PKG_SCRIPTS/postinstall" << 'EOF'
#!/bin/bash
echo ""
echo "GlyphLang installed successfully!"
echo ""
echo "Get started with:"
echo "  glyph --help"
echo "  glyph init my-project"
echo "  glyph dev my-project/main.glyph"
echo ""
exit 0
EOF
chmod 755 "$PKG_SCRIPTS/postinstall"

# Build component package
pkgbuild \
    --root "$PKG_ROOT" \
    --scripts "$PKG_SCRIPTS" \
    --identifier "dev.glyph-lang.glyph" \
    --version "$VERSION" \
    --install-location "/" \
    "$BUILD_DIR/glyph-component.pkg"

# Create distribution XML
cat > "$BUILD_DIR/distribution.xml" << EOF
<?xml version="1.0" encoding="utf-8"?>
<installer-gui-script minSpecVersion="1">
    <title>GlyphLang</title>
    <organization>dev.glyph-lang</organization>
    <domains enable_localSystem="true"/>
    <options customize="never" require-scripts="true" rootVolumeOnly="true"/>

    <welcome file="welcome.html" mime-type="text/html"/>
    <license file="license.txt" mime-type="text/plain"/>
    <conclusion file="conclusion.html" mime-type="text/html"/>

    <pkg-ref id="dev.glyph-lang.glyph"/>

    <choices-outline>
        <line choice="default">
            <line choice="dev.glyph-lang.glyph"/>
        </line>
    </choices-outline>

    <choice id="default"/>
    <choice id="dev.glyph-lang.glyph" visible="false">
        <pkg-ref id="dev.glyph-lang.glyph"/>
    </choice>

    <pkg-ref id="dev.glyph-lang.glyph" version="$VERSION" onConclusion="none">glyph-component.pkg</pkg-ref>
</installer-gui-script>
EOF

# Create resources
mkdir -p "$BUILD_DIR/resources"

cat > "$BUILD_DIR/resources/welcome.html" << 'EOF'
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, sans-serif; padding: 20px; }
        h1 { color: #7c3aed; }
        code { background: #f1f5f9; padding: 2px 6px; border-radius: 4px; }
    </style>
</head>
<body>
    <h1>GlyphLang</h1>
    <p><strong>AI-First Backend Language</strong></p>
    <p>Symbol-based syntax designed for LLM code generation. From prompt to production in nanoseconds.</p>
    <ul>
        <li>867ns compilation time</li>
        <li>2.93 ns/op VM execution</li>
        <li>Built-in security scanning</li>
        <li>WebSocket support</li>
    </ul>
    <p>This installer will install <code>glyph</code> to <code>/usr/local/bin</code>.</p>
</body>
</html>
EOF

cat > "$BUILD_DIR/resources/conclusion.html" << 'EOF'
<!DOCTYPE html>
<html>
<head>
    <style>
        body { font-family: -apple-system, BlinkMacSystemFont, sans-serif; padding: 20px; }
        h1 { color: #22c55e; }
        code { background: #f1f5f9; padding: 2px 6px; border-radius: 4px; }
        pre { background: #1e293b; color: #e2e8f0; padding: 15px; border-radius: 8px; }
    </style>
</head>
<body>
    <h1>Installation Complete!</h1>
    <p>GlyphLang has been installed successfully.</p>
    <p>Open Terminal and try:</p>
    <pre>glyph --help
glyph init my-project
glyph dev my-project/main.glyph</pre>
    <p>Visit <a href="https://github.com/glyph-lang/glyph">github.com/glyph-lang/glyph</a> for documentation.</p>
</body>
</html>
EOF

cat > "$BUILD_DIR/resources/license.txt" << 'EOF'
MIT License

Copyright (c) 2025 GlyphLang

Permission is hereby granted, free of charge, to any person obtaining a copy
of this software and associated documentation files (the "Software"), to deal
in the Software without restriction, including without limitation the rights
to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
copies of the Software, and to permit persons to whom the Software is
furnished to do so, subject to the following conditions:

The above copyright notice and this permission notice shall be included in all
copies or substantial portions of the Software.

THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE
SOFTWARE.
EOF

# Build product archive
productbuild \
    --distribution "$BUILD_DIR/distribution.xml" \
    --resources "$BUILD_DIR/resources" \
    --package-path "$BUILD_DIR" \
    "$DIST_DIR/glyph-${VERSION}-macos-${ARCH_NAME}.pkg"

echo "Created: dist/glyph-${VERSION}-macos-${ARCH_NAME}.pkg"

# Cleanup
rm -rf "$BUILD_DIR"
