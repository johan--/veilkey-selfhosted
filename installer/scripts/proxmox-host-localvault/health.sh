#!/usr/bin/env bash
set -euo pipefail

ROOT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/../.." && pwd)"
root="${1:-/}"

printf '[host-localvault/health] verifying local scaffold and post-install checks in %s\n' "${root}"
exec "${ROOT_DIR}/install.sh" post-install-health "${root}"
