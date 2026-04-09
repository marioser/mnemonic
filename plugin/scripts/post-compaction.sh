#!/usr/bin/env bash
# post-compaction.sh — Re-inject Knowledge Protocol after context compaction

set -euo pipefail

# Re-run session start to re-inject protocol
exec "${CLAUDE_PLUGIN_ROOT}/scripts/session-start.sh"
