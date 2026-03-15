#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
root="${1:-/}"

exec "${ROOT_DIR}/install.sh" plan-install proxmox-host-localvault "${root}"
