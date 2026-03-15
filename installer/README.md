# VeilKey Installer

`installer` is the self-hosted VeilKey installation and packaging component.

It assembles tested component versions, stages runtime assets, renders host-specific scaffolding, and verifies that the installed node is healthy.

## Product Position

VeilKey is split into:

- `managed`
  - `veilkey-docs`
  - `veilkey-homepage`
- `self-hosted`
  - `installer`
  - `keycenter`
  - `localvault`
  - `cli`
  - `proxy`

## Responsibilities

This component owns:

- install profiles
- component bundle layout
- Proxmox wrapper commands
- activation and health verification
- bootstrap export flows for initial installs

It does not own long-lived runtime business logic for KeyCenter, LocalVault, CLI, or Proxy.

## Canonical Targets

The active install targets are:

- `proxmox-host`
- `proxmox-host-cli`
- `proxmox-host-localvault`
- `proxmox-lxc-allinone`
- `proxmox-lxc-runtime`

## Main Documents

- `INSTALL.md`
  - operator install flow
- `profiles/`
  - install target inputs
- `scripts/`
  - wrapper commands
- `validation-logs/`
  - command-object validation records
- post-install validation

This repository is not the source of truth for:

- KeyCenter runtime code
- LocalVault runtime code
- long-term mirrors of component repositories

## Canonical Profiles

Only the following profiles are active:

- `proxmox-host`
- `proxmox-host-cli`
- `proxmox-host-localvault`
- `proxmox-lxc-allinone`
- `proxmox-lxc-runtime`

## Canonical Runtime Shapes

### `proxmox-host-localvault`

Purpose:

- install a host-side LocalVault node
- connect it to KeyCenter
- support install, health, and purge with explicit wrappers

Validated flow:

- install
- activate
- KeyCenter registration
- purge
- KeyCenter unregister

### `proxmox-lxc-allinone`

Purpose:

- install an all-in-one LXC with both KeyCenter and LocalVault
- expose both services by IP
- generate bootstrap SSH material during first install
- export bootstrap public and encrypted artifacts to the host

Validated flow:

- fresh one-shot install
- IP access to KeyCenter and LocalVault
- LocalVault registration inside KeyCenter
- bootstrap export
- health
- purge

## Architecture Summary

```text
Operator / CLI / UI
        |
        v
   VeilKey KeyCenter
        |
        +---- LocalVault (node A)
        +---- LocalVault (node B)
        +---- LocalVault (host)
```

Installer provisions the runtime shape. KeyCenter and LocalVault carry the runtime state.

## Command Model

Top-level commands:

```bash
./install.sh init
./install.sh validate
./install.sh doctor
./install.sh detect-os
./install.sh profiles
./install.sh plan-install <profile> <root>
./install.sh download <profile>
./install.sh stage <profile>
./install.sh bundle <profile>
./install.sh install <profile> <root>
./install.sh configure <profile> <root>
./install.sh install-profile <profile> <root>
./install.sh plan-activate <root>
./install.sh activate
./install.sh post-install-health <root>
```

Operator-facing wrappers:

```bash
./scripts/proxmox-host-install.sh /
./scripts/proxmox-host-localvault/install.sh --activate /
./scripts/proxmox-host-localvault/health.sh /
./scripts/proxmox-host-localvault/purge.sh /

./scripts/proxmox-lxc-allinone-install.sh --activate /
./scripts/proxmox-lxc-allinone-health.sh /
./scripts/proxmox-lxc-allinone-purge.sh /
./scripts/proxmox-lxc-allinone-export-bootstrap.sh <vmid> [dest_root]
```

## Bundle and Install Model

The installer works in four stages:

1. plan
2. download
3. stage / bundle
4. install / configure / activate

Important behavior:

- `install-profile` reuses an existing `bundle_root` when present
- fresh installs can download artifacts directly from the internal GitLab HTTPS source
- `post-install-health` validates the installed scaffold
- wrapper commands add target-specific runtime checks on top

## Bootstrap SSH Export

`proxmox-lxc-allinone` generates bootstrap SSH material on first install:

- private key
- public key
- encrypted export

The runtime stores bootstrap material inside the LXC, and the export wrapper copies the public and encrypted artifacts to the host.

Validated host export layout:

- `/opt/veilkey/bootstrap-exports/<vmid>-<hostname>/veilkey-admin.pub`
- `/opt/veilkey/bootstrap-exports/<vmid>-<hostname>/veilkey-admin.enc`
- `/opt/veilkey/bootstrap-exports/<vmid>-<hostname>/manifest.json`

## Validation

Validation logs are tracked as command objects under:

- [`validation-logs/validated`](./validation-logs/validated)
- [`validation-logs/pending`](./validation-logs/pending)

Each validated object log records:

- command
- target
- expected result
- observed result
- observed time
- proof
- artifacts
- exit code

## Related Repositories

- `veilkey-keycenter`
  - control plane
- `veilkey-localvault`
  - node runtime

## Installation Guide

For copy-paste installation steps, see [`INSTALL.md`](./INSTALL.md).

## Status

Current canonical state:

- active installer profiles are Proxmox-only
- the active remote branch surface is `main`
- the active host path uses the installer and localvault source repositories only

## License

See [`LICENSE`](./LICENSE).
