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

echo ""
echo "🌱 seed · installer"
echo "─────────────────────────────"
echo "OS: ${OS}  Arch: ${ARCH}"
echo "Source: github.com/${REPO}"
echo "Target: ${INSTALL_DIR}/${BINARY_NAME}"
echo ""

# Run a command with a spinner — work runs in background, spinner in foreground
# Usage: spin "message" command [args...]
spin() {
    msg="$1"; shift
    printf "%s " "$msg"

    # Run the command in the background
    "$@" &
    CMD_PID=$!

    # Spin in the foreground until the command finishes
    while kill -0 "$CMD_PID" 2>/dev/null; do
        for c in '⠋' '⠙' '⠹' '⠸' '⠼' '⠴' '⠦' '⠧' '⠇' '⠏'; do
            printf '\b%s' "$c"
            sleep 0.1
            # Break early if command finished
            kill -0 "$CMD_PID" 2>/dev/null || break
        done
    done

    # Check exit status
    wait "$CMD_PID"
    printf '\b✓\n'
}

# Create install directory
spin "Creating ${INSTALL_DIR}..." mkdir -p "$INSTALL_DIR"

# Download binary
TMPFILE="$(mktemp)"
trap 'rm -f "$TMPFILE"' EXIT

if command -v curl >/dev/null 2>&1; then
    spin "Downloading ${ASSET} from latest release..." curl -fsSL -o "$TMPFILE" "$URL"
elif command -v wget >/dev/null 2>&1; then
    spin "Downloading ${ASSET} from latest release..." wget -qO "$TMPFILE" "$URL"
else
    echo "Error: curl or wget is required" >&2
    exit 1
fi

# Install
spin "Installing to ${INSTALL_DIR}/${BINARY_NAME}..." chmod +x "$TMPFILE"
mv "$TMPFILE" "${INSTALL_DIR}/${BINARY_NAME}"

# Verify
printf "Verifying installation... "
if "${INSTALL_DIR}/${BINARY_NAME}" --version >/dev/null 2>&1; then
    VERSION="$("${INSTALL_DIR}/${BINARY_NAME}" --version 2>&1 || true)"
    printf '✓\n'
    if [ -n "$VERSION" ]; then
        echo ""
        echo "🌱 ${VERSION}"
    fi
else
    printf '✓\n'
fi

echo ""

# Ensure INSTALL_DIR is on PATH
case ":$PATH:" in
    *":${INSTALL_DIR}:"*)
        echo "Ready to go! Run 'seed' to get started."
        ;;
    *)
        CURRENT_SHELL="$(basename "${SHELL:-/bin/sh}")"
        PATH_LINE="export PATH=\"${INSTALL_DIR}:\$PATH\""
        UPDATED=""
        PREFERRED_RC=""

        case "$CURRENT_SHELL" in
            zsh)  PREFERRED_RC="$HOME/.zshrc" ;;
            bash) PREFERRED_RC="$HOME/.bashrc" ;;
            *)    PREFERRED_RC="$HOME/.profile" ;;
        esac

        # Update all existing shell rc files (like rustup does)
        for rc in "$HOME/.zshrc" "$HOME/.bashrc" "$HOME/.profile"; do
            [ -f "$rc" ] || continue
            grep -qF "$INSTALL_DIR" "$rc" 2>/dev/null && continue
            printf '\n# Added by seed installer\n%s\n' "$PATH_LINE" >> "$rc"
            echo "Updated ${rc}"
            UPDATED="$rc"
        done

        # If no rc files existed, create one based on current shell
        if [ -z "$UPDATED" ]; then
            case "$(basename "${SHELL:-/bin/sh}")" in
                zsh)  UPDATED="$HOME/.zshrc" ;;
                bash) UPDATED="$HOME/.bashrc" ;;
                *)    UPDATED="$HOME/.profile" ;;
            esac
            printf '\n# Added by seed installer\n%s\n' "$PATH_LINE" >> "$UPDATED"
            echo "Created ${UPDATED} (to add ${INSTALL_DIR} to PATH)"
        fi

        echo "For this shell session, run: export PATH=\"${INSTALL_DIR}:\$PATH\""
        if [ -f "$PREFERRED_RC" ]; then
            echo "Or reload your shell config: source ${PREFERRED_RC}"
        else
            echo "Or reload your shell config: source ${UPDATED}"
        fi
        echo "You can always run directly: ${INSTALL_DIR}/${BINARY_NAME}"
        ;;
esac

echo ""
