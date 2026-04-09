#!/usr/bin/env bash
# install-plugin.sh — Install mnemonic as a Claude Code plugin
#
# Usage:
#   ./scripts/install-plugin.sh          # Install from local repo
#   ./scripts/install-plugin.sh --remove # Remove plugin
#
# What it does:
#   1. Copies plugin/ directory to ~/.claude/plugins/cache/mnemonic/mnemonic/0.1.0/
#   2. Registers the plugin in ~/.claude/plugins/installed_plugins.json
#   3. Enables the plugin in ~/.claude/settings.json
#
# Requires: mnemonic binary in PATH (install with 'go install' or from release)

set -euo pipefail

VERSION="0.1.0"
PLUGIN_NAME="mnemonic"
PLUGIN_KEY="${PLUGIN_NAME}@${PLUGIN_NAME}"
PLUGIN_DIR="${HOME}/.claude/plugins/cache/${PLUGIN_NAME}/${PLUGIN_NAME}/${VERSION}"
INSTALLED_FILE="${HOME}/.claude/plugins/installed_plugins.json"
SETTINGS_FILE="${HOME}/.claude/settings.json"

# Find repo root (where this script lives)
SCRIPT_DIR="$(cd "$(dirname "$0")" && pwd)"
REPO_DIR="$(dirname "$SCRIPT_DIR")"
PLUGIN_SRC="${REPO_DIR}/plugin"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

info()  { echo -e "${GREEN}[+]${NC} $1"; }
warn()  { echo -e "${YELLOW}[!]${NC} $1"; }
error() { echo -e "${RED}[x]${NC} $1"; exit 1; }

# Remove mode
if [ "${1:-}" = "--remove" ]; then
    info "Removing mnemonic plugin..."
    rm -rf "$PLUGIN_DIR"
    if [ -f "$INSTALLED_FILE" ]; then
        python3 -c "
import json
path = '${INSTALLED_FILE}'
with open(path) as f:
    data = json.load(f)
if 'plugins' in data and '${PLUGIN_KEY}' in data['plugins']:
    del data['plugins']['${PLUGIN_KEY}']
with open(path, 'w') as f:
    json.dump(data, f, indent=2)
" 2>/dev/null || true
    fi
    info "Plugin removed. Restart Claude Code to apply."
    exit 0
fi

# Check prerequisites
if ! command -v mnemonic &> /dev/null; then
    error "mnemonic binary not found in PATH. Install it first:
    go install github.com/marioser/mnemonic/cmd/mnemonic@latest
    # or download from https://github.com/marioser/mnemonic/releases"
fi

if [ ! -d "$PLUGIN_SRC" ]; then
    error "Plugin source not found at ${PLUGIN_SRC}. Run from the mnemonic repo root."
fi

# 1. Copy plugin files
info "Installing plugin to ${PLUGIN_DIR}..."
mkdir -p "$PLUGIN_DIR"
cp -R "${PLUGIN_SRC}/"* "$PLUGIN_DIR/"
cp -R "${PLUGIN_SRC}/".* "$PLUGIN_DIR/" 2>/dev/null || true
chmod +x "${PLUGIN_DIR}/scripts/"*.sh

# 2. Register in installed_plugins.json (version 2 format)
info "Registering plugin..."
mkdir -p "$(dirname "$INSTALLED_FILE")"

python3 -c "
import json, datetime, os

path = '${INSTALLED_FILE}'

# Load or create
if os.path.exists(path):
    with open(path) as f:
        data = json.load(f)
else:
    data = {'version': 2, 'plugins': {}}

# Ensure structure
if 'version' not in data:
    data = {'version': 2, 'plugins': data if isinstance(data, dict) else {}}
if 'plugins' not in data:
    data['plugins'] = {}

# Add/update entry
data['plugins']['${PLUGIN_KEY}'] = [{
    'scope': 'user',
    'installPath': '${PLUGIN_DIR}',
    'version': '${VERSION}',
    'installedAt': datetime.datetime.now(datetime.timezone.utc).strftime('%Y-%m-%dT%H:%M:%S.000Z'),
    'lastUpdated': datetime.datetime.now(datetime.timezone.utc).strftime('%Y-%m-%dT%H:%M:%S.000Z'),
    'source': {
        'type': 'github',
        'url': 'https://github.com/marioser/mnemonic'
    }
}]

with open(path, 'w') as f:
    json.dump(data, f, indent=2)
"

# 3. Enable in settings.json
info "Enabling plugin in settings..."
if [ ! -f "$SETTINGS_FILE" ]; then
    echo '{}' > "$SETTINGS_FILE"
fi

python3 -c "
import json

path = '${SETTINGS_FILE}'
with open(path) as f:
    data = json.load(f)

if 'enabledPlugins' not in data:
    data['enabledPlugins'] = {}
data['enabledPlugins']['${PLUGIN_KEY}'] = True

with open(path, 'w') as f:
    json.dump(data, f, indent=2)
"

info "Plugin installed successfully!"
echo ""
echo "  Version:  ${VERSION}"
echo "  Location: ${PLUGIN_DIR}"
echo "  Binary:   $(which mnemonic)"
echo ""
warn "Restart Claude Code to activate hooks and Knowledge Protocol."
echo ""
echo "  Hooks installed:"
echo "    SessionStart   Injects Knowledge Protocol + KB status"
echo "    UserPrompt     Contextual search nudges"
echo "    SubagentStop   Captures agent knowledge (async)"
echo "    SessionStop    Cleanup (async)"
echo "    Compaction     Re-injects protocol after context compaction"
