#!/usr/bin/env bash
# Sets up a Python virtual environment for runtime validation of generated code.
# Usage: ./setup.sh
set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
VENV_DIR="$SCRIPT_DIR/.venv"

if [ ! -d "$VENV_DIR" ]; then
    echo "Creating virtual environment..."
    if python3 -m venv "$VENV_DIR" 2>/dev/null; then
        :
    else
        echo "ensurepip not available, creating venv without pip and bootstrapping..."
        python3 -m venv --without-pip "$VENV_DIR"
        curl -sS https://bootstrap.pypa.io/get-pip.py -o "$SCRIPT_DIR/get-pip.py"
        "$VENV_DIR/bin/python3" "$SCRIPT_DIR/get-pip.py" --quiet
        rm -f "$SCRIPT_DIR/get-pip.py"
    fi
fi

echo "Installing dependencies..."
"$VENV_DIR/bin/pip" install --quiet \
    fastapi uvicorn pydantic httpx pytest \
    pyjwt "python-jose[cryptography]" "passlib[bcrypt]" "bcrypt<4.1"

echo "Ready. Run tests with:"
echo "  $VENV_DIR/bin/pytest $SCRIPT_DIR/test_generated_servers.py -v"
