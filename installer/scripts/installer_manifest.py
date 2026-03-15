#!/usr/bin/env python3
import argparse
import json
import sys
from pathlib import Path

try:
    import tomllib
except ModuleNotFoundError:
    print("Error: Python 3.11+ tomllib is required", file=sys.stderr)
    sys.exit(1)


def load_manifest(path: Path) -> dict:
    if not path.exists():
        print(f"Error: manifest not found: {path}", file=sys.stderr)
        sys.exit(1)
    with path.open("rb") as f:
        return tomllib.load(f)


def validate_manifest(data: dict) -> None:
    if "release" not in data:
        raise ValueError("missing [release]")
    if "components" not in data:
        raise ValueError("missing [components]")
    for field in ("name", "version", "channel"):
        if field not in data["release"]:
            raise ValueError(f"release missing field: {field}")
    for name, component in data["components"].items():
        for field in ("source", "project", "ref", "type", "install_order"):
            if field not in component:
                raise ValueError(f"component {name} missing field: {field}")
        for field in ("stage_assets", "post_install_verify"):
            if field in component and not isinstance(component[field], list):
                raise ValueError(f"component {name} field {field} must be a list")
    profiles = data.get("profiles", {})
    for profile_name, profile in profiles.items():
        components = profile.get("components")
        if not isinstance(components, list) or not components:
            raise ValueError(f"profile {profile_name} missing components list")
        for component_name in components:
            if component_name not in data["components"]:
                raise ValueError(
                    f"profile {profile_name} references unknown component: {component_name}"
                )


def sorted_components(data: dict, selected: list[str] | None = None):
    components = data["components"]
    names = selected or list(components.keys())
    return sorted(
        ((name, components[name]) for name in names),
        key=lambda item: int(item[1]["install_order"]),
    )


def encode_list_field(value) -> str:
    if not value:
        return ""
    if not isinstance(value, list):
        raise ValueError("stage metadata list field must be a list")
    return ",".join(str(item) for item in value)


def artifact_filename_for(name: str, component: dict) -> str:
    explicit = component.get("artifact_filename", "")
    if explicit:
        return str(explicit)
    artifact_url = str(component.get("artifact_url", ""))
    if "?" in artifact_url:
        artifact_url = artifact_url.split("?", 1)[0]
    candidate = artifact_url.rsplit("/", 1)[-1] if artifact_url else ""
    if candidate:
        return candidate
    return f"{name}.artifact"


def main() -> int:
    parser = argparse.ArgumentParser()
    parser.add_argument("--manifest", required=True)
    sub = parser.add_subparsers(dest="cmd", required=True)
    sub.add_parser("validate")
    sub.add_parser("list-components")
    sub.add_parser("list-profiles")
    sub.add_parser("lint-legacy-layout")
    p_plan = sub.add_parser("plan")
    p_plan.add_argument("profile")
    p_download = sub.add_parser("plan-download")
    p_download.add_argument("profile")
    p_stage = sub.add_parser("plan-stage")
    p_stage.add_argument("profile")
    sub.add_parser("print-json")
    args = parser.parse_args()

    data = load_manifest(Path(args.manifest))
    try:
        validate_manifest(data)
    except ValueError as exc:
        print(f"Error: {exc}", file=sys.stderr)
        return 1

    if args.cmd == "validate":
        print("ok")
        return 0

    if args.cmd == "list-components":
        for name, component in sorted_components(data):
            print(
                f"{name}\t{component['project']}\t{component['ref']}\t"
                f"{component['type']}\t{component['install_order']}"
            )
        return 0

    if args.cmd == "list-profiles":
        profiles = data.get("profiles", {})
        for name, profile in profiles.items():
            desc = profile.get("description", "")
            print(f"{name}\t{desc}")
        return 0

    if args.cmd == "lint-legacy-layout":
        root = Path(args.manifest).resolve().parent
        legacy_dirs = [
            path.name
            for path in (
                root / "veilkey-keycenter",
                root / "veilkey-localvault",
            )
            if path.exists()
        ]
        if legacy_dirs:
            print("legacy component directories still present:")
            for name in legacy_dirs:
                print(f"- {name}")
            return 0
        print("no legacy component directories found")
        return 0

    if args.cmd == "plan":
        profiles = data.get("profiles", {})
        if args.profile not in profiles:
            print(f"Error: unknown profile: {args.profile}", file=sys.stderr)
            return 1
        selected = profiles[args.profile].get("components", [])
        print(f"[profile] {args.profile}")
        for name, component in sorted_components(data, selected):
            print(
                f"{component['install_order']:>3} {name:<12} "
                f"{component['project']}@{component['ref']}"
            )
        return 0

    if args.cmd == "plan-download":
        profiles = data.get("profiles", {})
        if args.profile not in profiles:
            print(f"Error: unknown profile: {args.profile}", file=sys.stderr)
            return 1
        selected = profiles[args.profile].get("components", [])
        print(f"[profile] {args.profile}")
        for name, component in sorted_components(data, selected):
            artifact_url = component.get("artifact_url", "")
            if not artifact_url:
                print(
                    f"Error: component {name} missing artifact_url for download plan",
                    file=sys.stderr,
                )
                return 1
            print(
                f"{component['install_order']:>3} {name:<12} "
                f"{artifact_filename_for(name, component)} "
                f"{artifact_url}"
            )
        return 0

    if args.cmd == "plan-stage":
        profiles = data.get("profiles", {})
        if args.profile not in profiles:
            print(f"Error: unknown profile: {args.profile}", file=sys.stderr)
            return 1
        selected = profiles[args.profile].get("components", [])
        release = data["release"]
        print(f"release_name={release['name']}")
        print(f"release_version={release['version']}")
        print(f"release_channel={release['channel']}")
        print(f"profile={args.profile}")
        for name, component in sorted_components(data, selected):
            print(
                "component="
                f"{name};project={component['project']};ref={component['ref']};"
                f"type={component['type']};install_order={component['install_order']};"
                f"artifact_url={component.get('artifact_url', '')};"
                f"artifact_filename={artifact_filename_for(name, component)};"
                f"stage_assets={encode_list_field(component.get('stage_assets'))};"
                f"post_install_verify={encode_list_field(component.get('post_install_verify'))}"
            )
        return 0

    if args.cmd == "print-json":
        print(json.dumps(data, indent=2, sort_keys=True))
        return 0

    return 1


if __name__ == "__main__":
    raise SystemExit(main())
