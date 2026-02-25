#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
INSTALLER="${ROOT_DIR}/install.sh"

TMPDIR="$(mktemp -d)"
trap 'rm -rf "${TMPDIR}"' EXIT

HOME_DIR="${TMPDIR}/home"
FAKE_BIN="${TMPDIR}/fake-bin"
mkdir -p "${HOME_DIR}" "${FAKE_BIN}"

cat > "${HOME_DIR}/.zshrc" <<'EOF'
# test zshrc
EOF

cat > "${TMPDIR}/seed-payload" <<'EOF'
#!/bin/sh
if [ "${1:-}" = "--version" ]; then
    echo "seed version v9.9.9"
    exit 0
fi
exit 0
EOF
chmod +x "${TMPDIR}/seed-payload"

cat > "${FAKE_BIN}/uname" <<'EOF'
#!/bin/sh
if [ "${1:-}" = "-s" ]; then
    echo "Linux"
    exit 0
fi
if [ "${1:-}" = "-m" ]; then
    echo "x86_64"
    exit 0
fi
echo "unsupported uname call" >&2
exit 1
EOF
chmod +x "${FAKE_BIN}/uname"

cat > "${FAKE_BIN}/curl" <<'EOF'
#!/bin/sh
set -e
OUT=""
while [ "$#" -gt 0 ]; do
    case "$1" in
        -o)
            OUT="$2"
            shift 2
            ;;
        *)
            shift
            ;;
    esac
done

[ -n "$OUT" ] || {
    echo "missing -o output path" >&2
    exit 1
}

cp "${SEED_TEST_PAYLOAD}" "$OUT"
EOF
chmod +x "${FAKE_BIN}/curl"

HOME="${HOME_DIR}" \
SHELL="/bin/zsh" \
INSTALL_DIR="${HOME_DIR}/.local/bin" \
PATH="${FAKE_BIN}:${PATH}" \
SEED_TEST_PAYLOAD="${TMPDIR}/seed-payload" \
/bin/sh "${INSTALLER}" >/dev/null

INSTALLED_VERSION="$({
    PATH="${HOME_DIR}/.local/bin:${FAKE_BIN}:${PATH}" \
    seed --version
})"
if [ "${INSTALLED_VERSION}" != "seed version v9.9.9" ]; then
    printf 'Unexpected installed version output: %s\n' "${INSTALLED_VERSION}" >&2
    exit 1
fi

printf 'Installer integration test passed.\n'
