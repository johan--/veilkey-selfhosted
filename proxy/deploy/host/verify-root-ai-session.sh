#!/usr/bin/env bash
set -euo pipefail

launcher_bin="${VEILKEY_ROOT_AI_LAUNCHER:-/usr/local/bin/veilkey-root-ai-session}"
session_config_bin="${VEILKEY_SESSION_CONFIG_BIN:-/usr/local/bin/veilkey-session-config}"

profile="${1:-}"
if [[ -n "$profile" ]]; then
  :
else
  profile="$("$session_config_bin" root-ai-default-profile)"
fi

unit_prefix="${VEILKEY_ROOT_AI_UNIT_PREFIX:-$("$session_config_bin" root-ai-unit-prefix)}"
scope_name="${VEILKEY_ROOT_AI_SCOPE:-${unit_prefix}-${profile}}"

out="$("$launcher_bin" "$profile")"
printf '%s\n' "$out"

expected_proxy="$("$session_config_bin" tool-proxy-url "$profile")"
printf '%s\n' "$out" | grep -q "^VEILKEY_PROXY_URL=${expected_proxy}$"
printf '%s\n' "$out" | grep -q '^VEILKEY_ROOT_AI=1$'
printf '%s\n' "$out" | grep -q "^VEILKEY_ROOT_AI_PROFILE=${profile}$"
printf '%s\n' "$out" | grep -q "${scope_name}"

echo "ok: root-ai session verify (${profile})"
