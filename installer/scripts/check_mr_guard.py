#!/usr/bin/env python3
import os
import re
import subprocess
import sys
from pathlib import Path

try:
    import tomllib
except ModuleNotFoundError:
    print("mr-guard: Python 3.11+ tomllib is required", file=sys.stderr)
    raise SystemExit(1)


def sh(*args: str) -> str:
    return subprocess.check_output(args, text=True).strip()


def load_rulebook(path: Path) -> dict:
    with path.open("rb") as fh:
        return tomllib.load(fh)


def compile_patterns(items: list[str]) -> list[re.Pattern[str]]:
    return [re.compile(item) for item in items]


def any_match(paths: list[str], patterns: list[re.Pattern[str]]) -> bool:
    for path in paths:
        for pattern in patterns:
            if pattern.search(path):
                return True
    return False


def changed_files(base_sha: str, head: str = "HEAD") -> list[str]:
    out = sh("git", "diff", "--name-only", f"{base_sha}...{head}")
    return [line for line in out.splitlines() if line.strip()]


def resolve_base_sha(target_branch: str) -> str:
    env_base = os.environ.get("CI_MERGE_REQUEST_DIFF_BASE_SHA", "").strip()
    if env_base:
        return env_base
    try:
        subprocess.run(
            ["git", "rev-parse", "--verify", f"origin/{target_branch}"],
            check=True,
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
            text=True,
        )
    except subprocess.CalledProcessError:
        subprocess.run(
            ["git", "fetch", "origin", f"{target_branch}:{target_branch}"],
            check=False,
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
            text=True,
        )
        subprocess.run(
            ["git", "fetch", "origin", target_branch],
            check=False,
            stdout=subprocess.DEVNULL,
            stderr=subprocess.DEVNULL,
            text=True,
        )
    try:
        return sh("git", "merge-base", "HEAD", f"origin/{target_branch}")
    except subprocess.CalledProcessError:
        print("mr-guard: could not determine base sha", file=sys.stderr)
        raise SystemExit(1)


def main() -> int:
    repo_root = Path(__file__).resolve().parent.parent
    rulebook = load_rulebook(repo_root / "rulebook" / "veilkey-rulebook.toml")
    mr_guard = rulebook["consumers"]["mr_guard"]
    target_branch = os.environ.get("CI_MERGE_REQUEST_TARGET_BRANCH_NAME", "main")
    base_sha = resolve_base_sha(target_branch)
    changed = changed_files(base_sha)
    if not changed:
        print("mr-guard: no changed files")
        return 0
    for path in changed:
        print(path)

    runtime_patterns = compile_patterns(mr_guard["runtime_paths"])
    deploy_patterns = compile_patterns(mr_guard["deploy_paths"])
    docs_required_patterns = compile_patterns(mr_guard["docs_required_paths"])
    doc_patterns = compile_patterns(mr_guard["doc_paths"])
    test_patterns = compile_patterns(mr_guard["test_paths"])

    need_tests = False
    need_docs = False
    if any_match(changed, runtime_patterns):
        need_tests = True
    if any_match(changed, docs_required_patterns):
        need_docs = True
    if any_match(changed, deploy_patterns):
        need_tests = True
        need_docs = True

    if need_tests and not any_match(changed, test_patterns):
        print("mr-guard: runtime/deploy changes require test updates in the same MR", file=sys.stderr)
        return 1
    if need_docs and not any_match(changed, doc_patterns):
        print("mr-guard: user-facing or deploy changes require README/docs updates in the same MR", file=sys.stderr)
        return 1

    print("mr-guard: pass")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
