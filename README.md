# VeilKey Self-Hosted

`veilkey-selfhosted` is a unified source tree for the self-hosted VeilKey product surface.

It groups the active self-hosted repositories under one top-level workspace while preserving the existing component boundaries as folders.

## Product Split

VeilKey is organized in two product domains:

- `managed`
  - `veilkey-docs`
  - `veilkey-homepage`
- `self-hosted`
  - `installer`
  - `keycenter`
  - `localvault`
  - `cli`
  - `proxy`

This repository contains only the `self-hosted` domain.

## Layout

- `installer/`
  - packaging, install profiles, Proxmox wrappers, health checks
- `keycenter/`
  - central control plane, inventory, policy, lifecycle orchestration
- `localvault/`
  - node-local runtime for secrets, configs, identity, and apply execution
- `cli/`
  - operator CLI, secure terminal wrapping, and the `veilroot` host boundary
- `proxy/`
  - outbound enforcement, session egress control, and rewrite auditing

## Runtime Model

The active runtime model is:

- `keycenter`
  - central control plane
- `localvault`
  - node-local agent
- `cli`
  - operator-facing entrypoint
- `proxy`
  - outbound enforcement layer
- `installer`
  - installation and verification layer

## Scope

This repository is intended to keep the self-hosted VeilKey surface in one place without flattening component responsibilities.

Each folder remains the owner of its own source, tests, and operational contracts.
