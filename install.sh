#!/bin/sh
set -e

REPO="justinphilpott/seed"
INSTALL_DIR="${INSTALL_DIR:-$HOME/.local/bin}"
BINARY_NAME="seed"

# Detect OS
OS="$(uname -s)"
case "$OS" in
    Linux)  OS="linux" ;;
    Darwin) OS="darwin" ;;
    *)      echo "Error: unsupported OS: $OS" >&2; exit 1 ;;
esac

# Detect architecture
ARCH="$(uname -m)"
case "$ARCH" in
    x86_64)  ARCH="amd64" ;;
    aarch64) ARCH="arm64" ;;
    arm64)   ARCH="arm64" ;;
    *)       echo "Error: unsupported architecture: $ARCH" >&2; exit 1 ;;
esac

ASSET="seed-${OS}-${ARCH}"
URL="https://github.com/${REPO}/releases/latest/download/${ASSET}"

echo "Installing seed (${OS}/${ARCH})..."

# Create install directory
mkdir -p "$INSTALL_DIR"

# Download binary
TMPFILE="$(mktemp)"
trap 'rm -f "$TMPFILE"' EXIT

if command -v curl >/dev/null 2>&1; then
    curl -fsSL -o "$TMPFILE" "$URL"
elif command -v wget >/dev/null 2>&1; then
    wget -qO "$TMPFILE" "$URL"
else
    echo "Error: curl or wget is required" >&2
    exit 1
fi

# Install
mv "$TMPFILE" "${INSTALL_DIR}/${BINARY_NAME}"
chmod +x "${INSTALL_DIR}/${BINARY_NAME}"

# Verify
if "${INSTALL_DIR}/${BINARY_NAME}" --version >/dev/null 2>&1; then
    echo "seed installed to ${INSTALL_DIR}/${BINARY_NAME}"
    VERSION="$("${INSTALL_DIR}/${BINARY_NAME}" --version 2>&1 || true)"
    if [ -n "$VERSION" ]; then
        echo "$VERSION"
    fi
else
    echo "seed installed to ${INSTALL_DIR}/${BINARY_NAME}"
fi

# PATH hint
case ":$PATH:" in
    *":${INSTALL_DIR}:"*) ;;
    *)
        echo ""
        echo "Add ${INSTALL_DIR} to your PATH:"
        echo "  export PATH=\"${INSTALL_DIR}:\$PATH\""
        ;;
esac
