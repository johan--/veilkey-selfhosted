#!/usr/bin/env bash
set -euo pipefail

root="${1:-/}"

stage() {
  printf '[lxc-allinone/purge] %s\n' "$*"
}

purge_staged_root() {
  local target_root="${1%/}"
  rm -f "${target_root}/etc/systemd/system/veilkey-keycenter.service"
  rm -f "${target_root}/etc/systemd/system/veilkey-localvault.service"
  rm -f "${target_root}/etc/systemd/system/veilkey-egress-proxy@.service"
  rm -f "${target_root}/etc/veilkey/keycenter.env"
  rm -f "${target_root}/etc/veilkey/keycenter.env.example"
  rm -f "${target_root}/etc/veilkey/localvault.env"
  rm -f "${target_root}/etc/veilkey/localvault.env.example"
  rm -f "${target_root}/etc/veilkey/proxy.env"
  rm -f "${target_root}/etc/veilkey/proxy.env.example"
  rm -f "${target_root}/etc/veilkey/services.enabled"
  rm -f "${target_root}/etc/veilkey/installer-profile.env"
  rm -f "${target_root}/etc/veilkey/session-tools.toml"
  rm -rf "${target_root}/etc/veilkey/bootstrap"
  rm -f "${target_root}/etc/profile.d/veilkey.sh"
  rm -f "${target_root}/usr/local/bin/veilkey-keycenter"
  rm -f "${target_root}/usr/local/bin/veilkey-localvault"
  rm -f "${target_root}/usr/local/bin/veilkey-session-config"
  rm -f "${target_root}/usr/local/bin/veilkey-proxy-launch"
  rm -f "${target_root}/usr/local/bin/veilroot-shell"
  rm -rf "${target_root}/usr/local/lib/veilkey-proxy"
  rm -rf "${target_root}/usr/local/share/veilkey"
  rm -rf "${target_root}/opt/veilkey"
}

if [[ "${root}" != "/" ]]; then
  stage "purging staged root ${root}"
  purge_staged_root "${root}"
  exit 0
fi

stage "this removes all-in-one services, env, binaries, data, proxy assets, and bootstrap ssh material"
stage "stopping services"
systemctl disable --now \
  veilkey-keycenter.service \
  veilkey-localvault.service \
  veilkey-egress-proxy@default.service \
  veilkey-egress-proxy@codex.service \
  veilkey-egress-proxy@claude.service \
  veilkey-egress-proxy@opencode.service 2>/dev/null || true

stage "removing local files"
purge_staged_root "/"
rm -rf /etc/systemd/system/multi-user.target.wants/veilkey-keycenter.service
rm -rf /etc/systemd/system/multi-user.target.wants/veilkey-localvault.service
rm -rf /etc/systemd/system/multi-user.target.wants/veilkey-egress-proxy@default.service
rm -rf /etc/systemd/system/multi-user.target.wants/veilkey-egress-proxy@codex.service
rm -rf /etc/systemd/system/multi-user.target.wants/veilkey-egress-proxy@claude.service
rm -rf /etc/systemd/system/multi-user.target.wants/veilkey-egress-proxy@opencode.service
systemctl daemon-reload
stage "completed"
