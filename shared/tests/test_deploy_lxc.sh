#!/bin/bash
set -euo pipefail

REPO_ROOT="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

assert_missing_cmd_fails() {
  local missing_path="$1"
  local want="$2"
  local out
  if out="$(PATH="$missing_path" bash "${REPO_ROOT}/scripts/deploy-lxc.sh" 2>&1)"; then
    echo "expected deploy-lxc.sh to fail without prerequisites" >&2
    exit 1
  fi
  grep -F "$want" <<<"$out" >/dev/null || {
    echo "missing expected error: $want" >&2
    echo "$out" >&2
    exit 1
  }
}

tmp_bin="$(mktemp -d)"
trap 'rm -rf "$tmp_bin"' EXIT

ln -sf /usr/bin/env "$tmp_bin/env"
ln -sf /bin/bash "$tmp_bin/bash"

assert_missing_cmd_fails "$tmp_bin" "required command not found: pct"

env_file="${tmp_bin}/veilkey-server.env"
pw_dir="${tmp_bin}/etc/veilkey"
pw_file="${pw_dir}/veilkey-server.password"
cat > "${env_file}" <<'EOF'
VEILKEY_ADDR=127.0.0.1:10181
VEILKEY_PASSWORD=legacy-secret
VEILKEY_DB_PATH=/opt/veilkey/data/veilkey.db
EOF

VEILKEY_SOURCE_ONLY=1 ENV_FILE="${env_file}" PW_FILE="${pw_file}" REPO_ROOT="${REPO_ROOT}" bash <<'EOF'
source "${REPO_ROOT}/scripts/deploy-lxc.sh"
lxc_exec() {
  shift
  local script
  script="$(mktemp)"
  printf '%s\n' "$*" > "${script}"
  bash "${script}"
  local rc=$?
  rm -f "${script}"
  return "${rc}"
}
migrate_legacy_password_env 104 "${ENV_FILE}" "${PW_FILE}"
EOF

grep -q '^VEILKEY_PASSWORD_FILE=' "${env_file}"
! grep -q '^VEILKEY_PASSWORD=' "${env_file}"
[[ -f "${pw_file}" ]]
[[ "$(tr -d '\r\n' < "${pw_file}")" == "legacy-secret" ]]

echo "ok"
