#!/usr/bin/env python3
import argparse
import sys
from pathlib import Path

try:
    import tomllib
except ModuleNotFoundError:
    print("Error: Python 3.11+ tomllib is required", file=sys.stderr)
    raise SystemExit(1)


REQUIRED_TOP_LEVEL = ("version", "name", "channel", "categories", "rules")
REQUIRED_RULE_FIELDS = ("id", "category", "level", "summary", "message")
ALLOWED_LEVELS = {"error", "warn", "info"}
MR_GUARD_PATTERN_FIELDS = (
    "runtime_paths",
    "deploy_paths",
    "docs_required_paths",
    "doc_paths",
    "test_paths",
)


def load_rulebook(path: Path) -> dict:
    if not path.exists():
        raise ValueError(f"rulebook not found: {path}")
    with path.open("rb") as fh:
        data = tomllib.load(fh)
    if not isinstance(data, dict):
        raise ValueError("rulebook root must be a table")
    return data


def validate_rulebook(data: dict) -> None:
    for key in REQUIRED_TOP_LEVEL:
        if key not in data:
            raise ValueError(f"missing top-level field: {key}")

    if not isinstance(data["version"], int) or data["version"] <= 0:
        raise ValueError("version must be a positive integer")
    for key in ("name", "channel"):
        if not isinstance(data[key], str) or not data[key].strip():
            raise ValueError(f"{key} must be a non-empty string")

    categories = data["categories"]
    if not isinstance(categories, dict):
        raise ValueError("categories must be a table")
    required = categories.get("required")
    if not isinstance(required, list) or not required:
        raise ValueError("categories.required must be a non-empty list")
    seen_categories = set()
    for item in required:
        if not isinstance(item, str) or not item.strip():
            raise ValueError("categories.required entries must be non-empty strings")
        if item in seen_categories:
            raise ValueError(f"duplicate category: {item}")
        seen_categories.add(item)

    rules = data["rules"]
    if not isinstance(rules, list) or not rules:
        raise ValueError("rules must be a non-empty array")

    consumers = data.get("consumers")
    if not isinstance(consumers, dict):
        raise ValueError("consumers must be a table")
    mr_guard = consumers.get("mr_guard")
    if not isinstance(mr_guard, dict):
        raise ValueError("consumers.mr_guard must be a table")
    for field in MR_GUARD_PATTERN_FIELDS:
        value = mr_guard.get(field)
        if not isinstance(value, list) or not value:
            raise ValueError(f"consumers.mr_guard.{field} must be a non-empty list")
        for item in value:
            if not isinstance(item, str) or not item.strip():
                raise ValueError(f"consumers.mr_guard.{field} entries must be non-empty strings")

    seen_ids = set()
    for index, rule in enumerate(rules, start=1):
        if not isinstance(rule, dict):
            raise ValueError(f"rule #{index} must be a table")
        for field in REQUIRED_RULE_FIELDS:
            if field not in rule:
                raise ValueError(f"rule #{index} missing field: {field}")
        rule_id = rule["id"]
        if not isinstance(rule_id, str) or not rule_id.strip():
            raise ValueError(f"rule #{index} id must be a non-empty string")
        if rule_id in seen_ids:
            raise ValueError(f"duplicate rule id: {rule_id}")
        seen_ids.add(rule_id)
        category = rule["category"]
        if category not in seen_categories:
            raise ValueError(f"rule {rule_id} uses unknown category: {category}")
        level = rule["level"]
        if level not in ALLOWED_LEVELS:
            raise ValueError(f"rule {rule_id} has invalid level: {level}")
        for field in ("summary", "message"):
            value = rule[field]
            if not isinstance(value, str) or not value.strip():
                raise ValueError(f"rule {rule_id} field {field} must be a non-empty string")


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument(
        "--rulebook",
        default="rulebook/veilkey-rulebook.toml",
        help="path to the VeilKey rulebook TOML",
    )
    args = parser.parse_args()

    try:
        data = load_rulebook(Path(args.rulebook))
        validate_rulebook(data)
    except ValueError as exc:
        print(f"Error: {exc}", file=sys.stderr)
        return 1

    print("ok")
    return 0


if __name__ == "__main__":
    raise SystemExit(main())
