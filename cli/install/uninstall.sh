#!/usr/bin/env bash
set -euo pipefail

# VeilKey Guard Uninstaller

echo "=== VeilKey Guard Uninstall ==="

rm -f /usr/local/bin/veilkey-cli
echo "  removed: /usr/local/bin/veilkey-cli"

rm -f /usr/local/bin/vk_bash
echo "  removed: /usr/local/bin/vk_bash"

STATE_DIR="${VEILKEY_STATE_DIR:-${TMPDIR:-/tmp}/veilkey-cli}"
rm -rf "$STATE_DIR" 2>/dev/null || true
echo "  removed: $STATE_DIR"

INSTALL_DIR="${VEILKEY_INSTALL_DIR:-/opt/veilkey}"
if [ -d "$INSTALL_DIR" ]; then
    rm -rf "$INSTALL_DIR"
    echo "  removed: $INSTALL_DIR"
fi

echo ""
echo "Uninstall complete."
