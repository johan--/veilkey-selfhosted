#!/usr/bin/env bash
set -euo pipefail

usage() {
  cat <<'EOF'
Usage: ./scripts/windows_remote_check.sh <host>

Checks:
  - RDP 3389
  - WinRM 5985
  - WinRM TLS 5986
EOF
}

HOST="${1:-}"

if [[ -z "${HOST}" ]]; then
  usage >&2
  exit 1
fi

command -v nc >/dev/null 2>&1 || {
  echo "Error: nc not found" >&2
  exit 1
}

for port in 3389 5985 5986; do
  if nc -z -w 3 "${HOST}" "${port}" >/dev/null 2>&1; then
    echo "open ${HOST}:${port}"
  else
    echo "closed ${HOST}:${port}"
  fi
done
