#!/usr/bin/env bash
set -euo pipefail

echo "== services =="
systemctl --no-pager --full --lines=5 status \
  veilkey-egress-proxy@default.service \
  veilkey-egress-proxy@codex.service \
  veilkey-egress-proxy@claude.service \
  veilkey-egress-proxy@opencode.service \
  | sed -n '1,160p'

echo
echo "== listens =="
ss -ltnp | grep -E '18080|18081|18083|18084' || true

echo
echo "== http allow =="
curl -sS -D - -o /dev/null -x http://127.0.0.1:18080 http://github.com/ | sed -n '1,8p'

echo
echo "== http block =="
curl -sS -D - -o /dev/null -x http://127.0.0.1:18080 http://example.com/ | sed -n '1,8p'

echo
echo "== connect smoke =="
exec 3<>/dev/tcp/127.0.0.1/18081
printf 'CONNECT api.openai.com:443 HTTP/1.1\r\nHost: api.openai.com:443\r\n\r\n' >&3
dd bs=4096 count=1 <&3 2>/dev/null | sed -n '1,20p'
exec 3<&-
exec 3>&-

echo
echo "== recent logs =="
for f in \
  /var/log/veilkey-proxy/default.log \
  /var/log/veilkey-proxy/codex.log \
  /var/log/veilkey-proxy/default.jsonl \
  /var/log/veilkey-proxy/codex.jsonl \
  /var/log/veilkey-proxy/default-rewrite.jsonl \
  /var/log/veilkey-proxy/codex-rewrite.jsonl; do
  echo "--- $f ---"
  test -f "$f" && tail -n 20 "$f" || echo "missing"
done
