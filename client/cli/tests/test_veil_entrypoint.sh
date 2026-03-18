#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

mock_veilroot="$tmp/veilroot-shell"
cat >"$mock_veilroot" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
printf '%s\n' "$@" > "${VEIL_TEST_LOG}"
EOF
chmod +x "$mock_veilroot"

log="$tmp/veil.args"
VEILKEY_VEILROOT_SHELL_BIN="$mock_veilroot" \
VEIL_TEST_LOG="$log" \
  bash ./deploy/host/veil codex

grep -Fx 'open' "$log" >/dev/null
grep -Fx 'codex' "$log" >/dev/null

out="$(
  VEILKEY_VEILROOT_SHELL_BIN="$tmp/missing-veilroot-shell" \
    bash ./deploy/host/veil 2>&1 || true
)"
printf '%s\n' "$out" | grep -q 'required session entrypoint not found'

echo "ok: veil entrypoint"
