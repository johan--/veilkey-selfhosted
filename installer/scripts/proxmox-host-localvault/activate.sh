#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
root="${1:-/}"

printf '[host-localvault/activate] enabling and restarting services in %s\n' "${root}"
exec "${ROOT_DIR}/install.sh" activate "${root}"
