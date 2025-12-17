#!/bin/bash
# Build RPM package for GlyphLang
# Run on Fedora/RHEL/CentOS or in CI/CD

set -e

VERSION="${1:-1.0.0}"

echo "Building GlyphLang .rpm package v${VERSION}..."

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
DIST_DIR="$SCRIPT_DIR/../dist"

# Setup RPM build environment
mkdir -p ~/rpmbuild/{BUILD,RPMS,SOURCES,SPECS,SRPMS}

# Copy binary to SOURCES
if [ -f "$DIST_DIR/glyph-linux-amd64" ]; then
    cp "$DIST_DIR/glyph-linux-amd64" ~/rpmbuild/SOURCES/
else
    echo "Error: Binary not found. Run 'make build-all' first."
    exit 1
fi

# Copy and update spec file
cp "$SCRIPT_DIR/rpm/glyph.spec" ~/rpmbuild/SPECS/
sed -i "s/^Version:.*/Version:        ${VERSION}/" ~/rpmbuild/SPECS/glyph.spec

# Build the RPM
rpmbuild -bb ~/rpmbuild/SPECS/glyph.spec

# Copy to dist
find ~/rpmbuild/RPMS -name "glyph-*.rpm" -exec cp {} "$DIST_DIR/" \;

echo "Created RPM package in dist/"
ls -la "$DIST_DIR"/*.rpm 2>/dev/null || echo "RPM files created in ~/rpmbuild/RPMS/"
