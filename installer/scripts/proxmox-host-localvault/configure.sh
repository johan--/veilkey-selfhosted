#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
root="${1:-/}"

printf '[host-localvault/configure] rendering env/service scaffold into %s\n' "${root}"
exec "${ROOT_DIR}/install.sh" configure proxmox-host-localvault "${root}"
