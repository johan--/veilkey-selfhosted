# VeilKey Self-Hosted

`veilkey-selfhosted` is the self-hosted VeilKey product tree.

It packages the runtime services, installer, and operator CLI needed to run VeilKey on your own infrastructure instead of relying on a hosted control plane.

## What VeilKey Is

VeilKey is a self-hosted secret and execution-boundary system for local AI and operator workflows.

The active runtime model is:

- `services/keycenter`
  - central control plane
- `services/localvault`
  - node-local runtime
- `client/cli`
  - operator entrypoint
- `services/proxy`
  - outbound enforcement layer
- `installer`
  - installation and verification layer

## Why Self-Hosted

VeilKey is self-hosted because the main value is control over:

- where ciphertext and runtime state live
- how node identity and policy are enforced
- how Proxmox hosts and LXCs are provisioned
- how secrets are handled inside your own trust boundary

If you need a hosted SaaS secret manager, this repository is not that.
If you need VeilKey to live on your own host, LXC, and network boundary, this repository is the right surface.

## Quick Start

The fastest operator path is the installer.

```bash
git clone https://github.com/veilkey/veilkey-selfhosted.git
cd veilkey-selfhosted/installer
./install.sh validate
```

Then choose one of the validated install paths:

- all-in-one LXC

```bash
./scripts/proxmox-lxc-allinone-install.sh --activate /
./scripts/proxmox-lxc-allinone-health.sh /
```

- host-side LocalVault

```bash
./scripts/proxmox-host-localvault/install.sh --activate /
./scripts/proxmox-host-localvault/health.sh /
```

The full operator guide lives in [`installer/INSTALL.md`](./installer/INSTALL.md).

## Main Use Cases

- run KeyCenter and LocalVault inside your own Proxmox environment
- keep node-local runtime state under your own control
- use LocalVault as the node-local runtime paired with a central KeyCenter
- stage boundary and bootstrap assets for host companion setups

## How To Read This Repository

- `installer/`
  - install profiles, wrappers, health checks, and packaging
- `services/keycenter/`
  - central control plane
- `services/localvault/`
  - node-local runtime
- `services/proxy/`
  - outbound enforcement layer
- `client/cli/`
  - operator-facing CLI

## Why It Exists As One Repo

This repository keeps the self-hosted VeilKey surface in one place without flattening component responsibilities.

That means:

- install flow changes can ship with runtime changes
- operator docs can stay next to the code they describe
- CI can validate the self-hosted product as one surface

## Comparison Frame

VeilKey is not trying to be a generic password manager or a hosted secret vault.

The practical difference is:

- stronger emphasis on self-hosted runtime identity and node registration
- explicit Proxmox and LXC install paths
- local runtime components such as LocalVault instead of a cloud-only model
- tighter install-to-runtime contract inside one source tree

## Contributing

Start with [`CONTRIBUTING.md`](./CONTRIBUTING.md).

Short version:

- behavior changes need focused regression tests
- user-facing behavior changes need docs updates in the same change
- installer, runtime, and deploy changes should prove one real operator path

## License

This repository is licensed under the MIT License.

See [`LICENSE`](./LICENSE).
