#!/usr/bin/env bash
set -euo pipefail

repo_root="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"
ci_file="${repo_root}/.gitlab-ci.yml"

docker_block="$(awk '
  /^docker:/ {capture=1}
  capture {print}
  capture && /^[^[:space:]].*:$/ && $0 !~ /^docker:/ {exit}
' "$ci_file")"

deploy_block="$(awk '
  /^deploy:/ {capture=1}
  capture {print}
  capture && /^[^[:space:]].*:$/ && $0 !~ /^deploy:/ {exit}
' "$ci_file")"

printf '%s\n' "$docker_block" | grep -q 'allow_failure: true'
printf '%s\n' "$deploy_block" | grep -q 'proxmox-host'

echo ok
