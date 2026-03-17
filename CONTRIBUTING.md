# Contributing

This repository is the self-hosted VeilKey product surface.

Contributions are welcome, but changes should prove the exact path they touch.

## Ground Rules

1. Runtime, install, deploy, CLI, and API behavior changes must include focused regression tests.
2. User-facing or operator-facing behavior changes must update README or docs in the same change.
3. Installer and deploy changes should verify one real operator path whenever practical.
4. Do not introduce plaintext-secret examples where an existing masked or file-based path already exists.

## Where To Start

- repository overview: [`README.md`](./README.md)
- installer operator path: [`installer/INSTALL.md`](./installer/INSTALL.md)
- installer-specific guard rules: [`installer/CONTRIBUTING.md`](./installer/CONTRIBUTING.md)

## Common Contribution Areas

- installation and Proxmox wrapper improvements
- KeyCenter or LocalVault runtime fixes
- CLI UX and operator workflow improvements
- docs that make the self-hosted model easier to understand

## Minimum Validation

Use the smallest relevant validation first.

- installer

```bash
cd installer
./install.sh validate
```

- Go services or CLI

```bash
go test ./...
```

If your change is narrower than the full suite, add a focused regression test for that path.

## Pull Request Standard

A good contribution makes all of these easy to answer:

- what changed
- why it changed
- how it was verified
- what operator behavior changed, if any

If the answer to one of those is missing, the change is not ready yet.
