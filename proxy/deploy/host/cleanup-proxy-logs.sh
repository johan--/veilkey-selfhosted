#!/usr/bin/env bash
set -euo pipefail

proxy_vmid="${1:-100208}"
stamp="$(date +%Y%m%d-%H%M%S)"
archive_dir="/var/log/veilkey-proxy/archive/${stamp}"

echo "== archive proxy logs =="
echo "vmid=$proxy_vmid"
echo "archive_dir=$archive_dir"

vibe_lxc_ops "$proxy_vmid" "mkdir -p '$archive_dir' && shopt -s nullglob && for f in /var/log/veilkey-proxy/*.log /var/log/veilkey-proxy/*.jsonl; do base=\$(basename \"\$f\"); cp -a \"\$f\" '$archive_dir'/\"\$base\"; : >\"\$f\"; done && ls -la '$archive_dir'"

echo
echo "== post-clean current logs =="
vibe_lxc_ops "$proxy_vmid" "find /var/log/veilkey-proxy -maxdepth 1 -type f \\( -name '*.log' -o -name '*.jsonl' \\) -printf '%f %s\\n' | sort"
