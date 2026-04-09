#!/usr/bin/env bash
# install.sh — Install mnemonic binary for any OS/arch
# Usage: curl -fsSL https://raw.githubusercontent.com/marioser/mnemonic/main/scripts/install.sh | bash
set -euo pipefail

VERSION="${MNEMONIC_VERSION:-latest}"
INSTALL_DIR="${MNEMONIC_INSTALL_DIR:-/usr/local/bin}"
REPO="marioser/mnemonic"
GITHUB="https://github.com/${REPO}"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m'

info()  { echo -e "${GREEN}[+]${NC} $1"; }
warn()  { echo -e "${YELLOW}[!]${NC} $1"; }
error() { echo -e "${RED}[x]${NC} $1"; exit 1; }
step()  { echo -e "${BLUE}[>]${NC} $1"; }

echo ""
echo "  ┌─────────────────────────────────────┐"
echo "  │  mnemonic installer                  │"
echo "  │  Knowledge management for AI agents  │"
echo "  └─────────────────────────────────────┘"
echo ""

# --- Detect OS ---
OS="$(uname -s | tr '[:upper:]' '[:lower:]')"
ARCH="$(uname -m)"

case "$OS" in
  linux)   OS="linux" ;;
  darwin)  OS="darwin" ;;
  *)       error "Unsupported OS: $OS. Only linux and darwin are supported." ;;
esac

case "$ARCH" in
  x86_64|amd64)   ARCH="amd64" ;;
  arm64|aarch64)   ARCH="arm64" ;;
  *)               error "Unsupported architecture: $ARCH. Only amd64 and arm64 are supported." ;;
esac

info "Detected: ${OS}/${ARCH}"

# --- Resolve version ---
if [ "$VERSION" = "latest" ]; then
  step "Fetching latest version..."
  VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" | grep '"tag_name"' | head -1 | sed 's/.*"tag_name": *"//;s/".*//')
  if [ -z "$VERSION" ]; then
    error "Could not determine latest version. Check ${GITHUB}/releases"
  fi
fi

info "Version: ${VERSION}"

# --- Download ---
FILENAME="mnemonic_${VERSION#v}_${OS}_${ARCH}.tar.gz"
URL="${GITHUB}/releases/download/${VERSION}/${FILENAME}"

step "Downloading ${URL}..."
TMP_DIR=$(mktemp -d)
trap "rm -rf ${TMP_DIR}" EXIT

if ! curl -fsSL "$URL" -o "${TMP_DIR}/mnemonic.tar.gz"; then
  error "Download failed. Check if release ${VERSION} exists: ${GITHUB}/releases"
fi

# --- Extract ---
step "Extracting..."
tar xzf "${TMP_DIR}/mnemonic.tar.gz" -C "${TMP_DIR}"

if [ ! -f "${TMP_DIR}/mnemonic" ]; then
  error "Binary not found in archive. Contents: $(ls ${TMP_DIR})"
fi

# --- Install ---
step "Installing to ${INSTALL_DIR}/mnemonic..."

if [ -w "$INSTALL_DIR" ]; then
  mv "${TMP_DIR}/mnemonic" "${INSTALL_DIR}/mnemonic"
  chmod +x "${INSTALL_DIR}/mnemonic"
else
  warn "Need sudo to install to ${INSTALL_DIR}"
  sudo mv "${TMP_DIR}/mnemonic" "${INSTALL_DIR}/mnemonic"
  sudo chmod +x "${INSTALL_DIR}/mnemonic"
fi

# --- Verify ---
if command -v mnemonic &> /dev/null; then
  info "Installed: $(mnemonic version 2>&1 | head -1)"
else
  warn "Binary installed to ${INSTALL_DIR}/mnemonic but not in PATH"
  warn "Add to PATH: export PATH=\$PATH:${INSTALL_DIR}"
fi

# --- Config ---
if [ ! -f "$HOME/.mnemonic/config.yaml" ]; then
  step "Creating default config at ~/.mnemonic/config.yaml..."
  mkdir -p "$HOME/.mnemonic"
  cat > "$HOME/.mnemonic/config.yaml" << 'YAML'
# Mnemonic config — edit host/port to match your ChromaDB server
chromadb:
  host: "localhost"
  port: 8000
  token: ""
  ssl: false

server:
  port: 7438
  host: "127.0.0.1"

search:
  default_results: 5
  min_similarity: 0.7
YAML
  info "Config created. Edit ~/.mnemonic/config.yaml to set your ChromaDB host."
else
  info "Config already exists at ~/.mnemonic/config.yaml"
fi

# --- Done ---
echo ""
info "Installation complete!"
echo ""
echo "  Next steps:"
echo "    1. Ensure ChromaDB is running (docker or remote)"
echo "    2. Edit ~/.mnemonic/config.yaml with your ChromaDB host"
echo "    3. Run: mnemonic status"
echo "    4. Install Claude Code plugin:"
echo "       claude plugin marketplace add marioser/mnemonic"
echo "       claude plugin install mnemonic@mnemonic"
echo "    5. Restart Claude Code"
echo ""
