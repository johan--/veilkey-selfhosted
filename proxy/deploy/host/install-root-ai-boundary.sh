#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"

echo "install-root-ai-boundary.sh is deprecated; installing veilroot boundary instead" >&2
export VEILKEY_VEILROOT_BIN_DIR="${VEILKEY_VEILROOT_BIN_DIR:-${VEILKEY_ROOT_AI_BIN_DIR:-/usr/local/bin}}"
export VEILKEY_VEILROOT_SYSTEMD_DIR="${VEILKEY_VEILROOT_SYSTEMD_DIR:-${VEILKEY_ROOT_AI_SYSTEMD_DIR:-/etc/systemd/system}}"
export VEILKEY_VEILROOT_LOG_DIR="${VEILKEY_VEILROOT_LOG_DIR:-${VEILKEY_ROOT_AI_LOG_DIR:-/var/log/veilkey-proxy}}"
export VEILKEY_VEILROOT_USER="${VEILKEY_VEILROOT_USER:-veilroot}"
exec "$repo_root/deploy/host/install-veilroot-boundary.sh" "$@"
