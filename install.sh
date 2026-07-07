#!/bin/sh
set -e

REPO="AntTheLimey/pgecloudctl"
BINARY="pgecloudctl"

detect_os() {
    case "$(uname -s)" in
        Linux*)  echo "linux" ;;
        Darwin*) echo "darwin" ;;
        *)       echo "unsupported" ;;
    esac
}

detect_arch() {
    case "$(uname -m)" in
        x86_64|amd64)  echo "amd64" ;;
        arm64|aarch64) echo "arm64" ;;
        *)             echo "unsupported" ;;
    esac
}

OS=$(detect_os)
ARCH=$(detect_arch)

if [ "$OS" = "unsupported" ] || [ "$ARCH" = "unsupported" ]; then
    echo "Error: unsupported platform $(uname -s)/$(uname -m)" >&2
    exit 1
fi

echo "Detected platform: ${OS}/${ARCH}"

VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
    | grep '"tag_name"' | sed -E 's/.*"([^"]+)".*/\1/')

if [ -z "$VERSION" ]; then
    echo "Error: could not determine latest version" >&2
    exit 1
fi

echo "Latest version: ${VERSION}"
VERSION_NUM="${VERSION#v}"

ARCHIVE="${BINARY}_${VERSION_NUM}_${OS}_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${ARCHIVE}"
CHECKSUM_URL="https://github.com/${REPO}/releases/download/${VERSION}/checksums.txt"

TMPDIR=$(mktemp -d)
trap 'rm -rf "$TMPDIR"' EXIT

echo "Downloading ${ARCHIVE}..."
curl -fsSL -o "${TMPDIR}/${ARCHIVE}" "$URL"

echo "Verifying checksum..."
curl -fsSL -o "${TMPDIR}/checksums.txt" "$CHECKSUM_URL"

EXPECTED=$(grep "${ARCHIVE}" "${TMPDIR}/checksums.txt" | awk '{print $1}')
if [ -z "$EXPECTED" ]; then
    echo "Error: checksum not found for ${ARCHIVE}" >&2
    exit 1
fi

if command -v sha256sum >/dev/null 2>&1; then
    ACTUAL=$(sha256sum "${TMPDIR}/${ARCHIVE}" | awk '{print $1}')
elif command -v shasum >/dev/null 2>&1; then
    ACTUAL=$(shasum -a 256 "${TMPDIR}/${ARCHIVE}" | awk '{print $1}')
else
    echo "Warning: no sha256 tool found, skipping checksum verification" >&2
    ACTUAL="$EXPECTED"
fi

if [ "$EXPECTED" != "$ACTUAL" ]; then
    echo "Error: checksum mismatch" >&2
    echo "  expected: ${EXPECTED}" >&2
    echo "  actual:   ${ACTUAL}" >&2
    exit 1
fi

echo "Extracting..."
tar -xzf "${TMPDIR}/${ARCHIVE}" -C "$TMPDIR"

INSTALL_DIR="/usr/local/bin"
if [ ! -w "$INSTALL_DIR" ]; then
    INSTALL_DIR="${HOME}/.local/bin"
    mkdir -p "$INSTALL_DIR"
fi

mv "${TMPDIR}/${BINARY}" "${INSTALL_DIR}/${BINARY}"
chmod +x "${INSTALL_DIR}/${BINARY}"

echo "Installed ${BINARY} ${VERSION} to ${INSTALL_DIR}/${BINARY}"

# Install Claude Code skill if Claude Code is detected. The skill files
# are embedded in the binary (`skill install`), so no extra download is
# needed and the skill always matches the installed version.
# ~/.claude/skills is the directory Claude Code discovers automatically;
# the ~/.claude/plugins location used before v0.5 is not.
CLAUDE_DIR="${HOME}/.claude"
if [ -d "$CLAUDE_DIR" ]; then
    echo "Claude Code detected. Installing pgecloudctl skill..."
    if ! "${INSTALL_DIR}/${BINARY}" skill install; then
        echo "Warning: skill install failed. Binary remains installed." >&2
        echo "Retry manually with: ${BINARY} skill install" >&2
    fi
else
    echo "Claude Code not detected — skill not installed."
    echo "Run '${BINARY} skill install' later to enable AI integration."
fi
