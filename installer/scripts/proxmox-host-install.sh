#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

usage() {
  cat <<'EOF'
Usage: ./scripts/proxmox-host-install.sh [--activate] [--health] [root] [bundle_root]

Install the Proxmox host profile:
  proxmox-host = proxy
EOF
}

if [[ "${1:-}" =~ ^(-h|--help)$ ]]; then
  usage
  exit 0
fi

args=()
while [[ $# -gt 0 && "${1:-}" == --* ]]; do
  args+=("$1")
  shift
done

root="${1:-/}"
bundle_root="${2:-}"

if [[ -n "${bundle_root}" ]]; then
  exec "${ROOT_DIR}/install.sh" install-profile "${args[@]}" proxmox-host "${root}" "${bundle_root}"
fi
exec "${ROOT_DIR}/install.sh" install-profile "${args[@]}" proxmox-host "${root}"
