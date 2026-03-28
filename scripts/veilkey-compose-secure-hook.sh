#!/usr/bin/env bash
set -euo pipefail

steps_json="${VEILKEY_BULK_APPLY_STEPS_JSON:-[]}"
lv_url="${VEILKEY_LOCALVAULT_URL:-https://127.0.0.1:10180}"
agent_secret_file="${VEILKEY_LOCALVAULT_AGENT_SECRET_FILE:-/opt/veilkey-localvault/data/agent_secret}"
runtime_env_content="${VEILKEY_RUNTIME_ENV_CONTENT:-}"

python3 - "$steps_json" "$lv_url" "$agent_secret_file" "$runtime_env_content" "${1:-}" "${2:-}" <<'PY'
import json
import os
import pathlib
import ssl
import sys
import urllib.request

steps = json.loads(sys.argv[1] or "[]")
lv_url = sys.argv[2].rstrip("/")
agent_secret_file = pathlib.Path(sys.argv[3])
runtime_env_content = sys.argv[4]
service_dir_arg = sys.argv[5].strip()
project_arg = sys.argv[6].strip()
ssl_ctx = ssl._create_unverified_context()
agent_secret = agent_secret_file.read_text().strip() if agent_secret_file.exists() else ""

if steps:
    target = pathlib.Path(steps[0]["target_path"])
    service_dir = target.parent
    project = project_arg or service_dir.name
else:
    if not service_dir_arg:
        raise SystemExit("no bulk-apply steps or service_dir argument supplied")
    service_dir = pathlib.Path(service_dir_arg)
    project = project_arg or service_dir.name
    target = service_dir / ".env.veil"

runtime_dir = pathlib.Path("/run/veilkey/docker-compose")
runtime_dir.mkdir(parents=True, exist_ok=True)
runtime_env = runtime_dir / f"{project}.env"

def resolve_value(value: str) -> str:
    if value.startswith(("VK:LOCAL:", "VK:EXTERNAL:", "VK:TEMP:")):
        req = urllib.request.Request(
            f"{lv_url}/api/resolve/{value}",
            headers={"Authorization": f"Bearer {agent_secret}"},
        )
        with urllib.request.urlopen(req, context=ssl_ctx) as resp:
            return json.load(resp)["value"]
    if value.startswith("VE:LOCAL:"):
        key = value[len("VE:LOCAL:"):]
        req = urllib.request.Request(
            f"{lv_url}/api/configs/{key}",
            headers={"Authorization": f"Bearer {agent_secret}"},
        )
        with urllib.request.urlopen(req, context=ssl_ctx) as resp:
            return json.load(resp)["value"]
    return value

if runtime_env_content:
    runtime_env.write_text(runtime_env_content if runtime_env_content.endswith("\n") else runtime_env_content + "\n")
else:
    lines = []
    for raw_line in target.read_text().splitlines():
        line = raw_line.strip()
        if not line or line.startswith("#"):
            lines.append(raw_line)
            continue
        key, sep, value = raw_line.partition("=")
        if not sep:
            raise SystemExit(f"invalid env line: {raw_line}")
        lines.append(f"{key}={resolve_value(value)}")
    runtime_env.write_text("\n".join(lines) + "\n")
os.chmod(runtime_env, 0o600)

compose_file = service_dir / "docker-compose.yml"
os.execvp(
    "bash",
    [
        "bash",
        "-lc",
        f"trap 'rm -f {runtime_env}' EXIT; docker compose -f {compose_file} up -d --force-recreate",
    ],
)
PY
