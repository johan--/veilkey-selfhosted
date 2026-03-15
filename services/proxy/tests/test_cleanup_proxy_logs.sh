#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."
. tests/lib/testlib.sh

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

cat >"$tmp/vibe_lxc_ops" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
echo "$*" >"$TEST_STATE_DIR/cmd.txt"
printf '%s\n' "archive-done"
printf '%s\n' "default.log 0"
printf '%s\n' "default.jsonl 0"
EOF
chmod +x "$tmp/vibe_lxc_ops"

export PATH="$tmp:$PATH"
export TEST_STATE_DIR="$tmp"

out="$(deploy/host/cleanup-proxy-logs.sh 100208)"
assert_contains "$out" "archive_dir=/var/log/veilkey-proxy/archive/"
assert_contains "$out" "default.log 0"
assert_file_contains "$tmp/cmd.txt" "100208"

echo "ok: cleanup-proxy-logs"
