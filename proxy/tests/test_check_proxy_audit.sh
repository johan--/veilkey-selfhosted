#!/usr/bin/env bash
set -euo pipefail

cd "$(dirname "$0")/.."
. tests/lib/testlib.sh

tmp="$(mktemp -d)"
trap 'rm -rf "$tmp"' EXIT

cat >"$tmp/vibe_lxc_ops" <<'EOF'
#!/usr/bin/env bash
set -euo pipefail
case "$2" in
  *"tail -n '3' '/var/log/veilkey-proxy/codex.jsonl'"*)
    printf '%s\n' '{"action":"connect"}'
    ;;
  *"tail -n '3' '/var/log/veilkey-proxy/codex-rewrite.jsonl'"*)
    printf '%s\n' '{"veilkey":"VK:TEMP:abcd","pattern":"github-pat"}'
    ;;
  *"actions="*)
    printf '%s\n' "actions= {'connect': 1}"
    ;;
  *"patterns="*)
    printf '%s\n' "patterns= {'github-pat': 1}"
    printf '%s\n' "recent_refs= ['VK:TEMP:abcd']"
    ;;
  *)
    echo "unexpected: $2" >&2
    exit 1
    ;;
esac
EOF
chmod +x "$tmp/vibe_lxc_ops"
export PATH="$tmp:$PATH"

out="$(deploy/host/check-proxy-audit.sh codex 100208 3)"
assert_contains "$out" "profile=codex"
assert_contains "$out" "actions= {'connect': 1}"
assert_contains "$out" "VK:TEMP:abcd"

echo "ok: check-proxy-audit"
