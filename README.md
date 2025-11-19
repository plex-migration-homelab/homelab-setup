# Homelab Setup CLI

Go-based interactive setup wizard for configuring homelabs on Fedora CoreOS / UBlue uCore. Built for my NAB9 mini PC homelab—tunnels traffic through WireGuard to a VPS and mounts media from the backend file server over NFS.

## Scope & assumptions
- Single-node helper meant for my own homelab. If you grab it, expect "works on my LAN" defaults.
- Menu-based interactive wizard for post-installation configuration of containerized services.
- Inputs are trusted. The wizard validates obvious pitfalls but intentionally avoids enterprise-grade policy layers.
- **Target OS**: Fedora CoreOS and UBlue uCore only (other distributions are untested and unsupported).


## What it configures
- **Media:** Plex and Jellyfin with Intel QuickSync for hardware transcodes.
- **Portals:** Overseerr, Wizarr, and Nginx Proxy Manager on the VPS for public access.
- **Cloud:** Nextcloud + Collabora + Redis + PostgreSQL and Immich for photos.
- **Infrastructure:** WireGuard VPN tunneling, NFS mounts, Podman/Docker compose stacks, and systemd service units.

## Quick start
1. **Install Fedora CoreOS / UBlue uCore** on your target system.
2. **Copy the binary** to your system: `scp homelab-setup user@host:~`
3. **Run the wizard** as a regular user: `./homelab-setup`
4. **Follow the interactive menu** to configure user accounts, WireGuard VPN, NFS mounts, compose secrets, and deploy services.

The setup wizard creates systemd units, generates configuration files under `/srv/containers/`, and starts your containerized services.

## Documentation
- [`docs/getting-started.md`](docs/getting-started.md): Walkthrough for using the setup wizard.
- [`docs/reference/homelab-setup-cli.md`](docs/reference/homelab-setup-cli.md): Detailed CLI reference and setup steps explanation.
- [`homelab-setup/README.md`](homelab-setup/README.md): Go binary development guide.

## Building from source
```bash
cd homelab-setup/
make deps    # Download dependencies
make test    # Run tests
make build   # Build binary (output: bin/homelab-setup)
```
