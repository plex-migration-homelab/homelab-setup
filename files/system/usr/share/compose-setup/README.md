# Docker Compose Templates for Homelab Setup

This directory contains Docker Compose templates for deploying containerized services on Fedora CoreOS / UBlue uCore systems.

## Available Stacks

- **`media.yml`** - Plex, Jellyfin, Tautulli (media streaming with Intel QuickSync transcoding)
- **`web.yml`** - Nginx Proxy Manager, Overseerr, Wizarr, Organizr, Homepage
- **`cloud.yml`** - Nextcloud, Collabora, Immich (personal cloud and photo management)

## Critical: SELinux Volume Labels

**Fedora CoreOS has SELinux enabled by default.** All volume mounts in Docker Compose files **MUST** include proper SELinux context labels, or containers will fail with "Permission denied" errors.

### SELinux Label Types

#### `:Z` (Private Label - Exclusive Access)

Use for application data directories that are accessed by **only one container**.

**Examples:**
```yaml
volumes:
  - ${APPDATA_PATH}/plex:/config:Z
  - ${APPDATA_PATH}/jellyfin/config:/config:Z
  - ${APPDATA_PATH}/nextcloud/db:/var/lib/postgresql/data:Z
```

**When to use:**
- Container-specific configuration directories
- Database data directories (PostgreSQL, Redis)
- Application caches and temporary files

#### `:z` (Shared Label - Multi-Container Access)

Use for directories that are accessed by **multiple containers** or shared storage.

**Examples:**
```yaml
volumes:
  - /mnt/nas-media:/media:ro,z
  - /mnt/nas-nextcloud/data:/var/www/html/data:z
  - /mnt/nas-immich:/usr/src/app/upload:z
```

**When to use:**
- NFS-mounted media libraries
- Shared storage accessed by multiple services
- Directories that need to be accessed by the host and containers

#### No Label (Special Cases Only)

Some mounts don't need SELinux labels:

```yaml
volumes:
  - /dev/dri:/dev/dri  # Hardware devices
  - /etc/localtime:/etc/localtime:ro  # Read-only system files
  - /var/run/docker.sock:/var/run/docker.sock:ro  # Docker socket
```

**When to use:**
- Device files (`/dev/*`)
- Read-only system files (`/etc/localtime`)
- Unix sockets (`/var/run/docker.sock`)

### Syntax Examples

Combine SELinux labels with other mount options:

```yaml
volumes:
  # Read-only + shared label
  - /mnt/nas-media:/media:ro,z

  # Cached mount + shared label (for devcontainers)
  - ..:/workspace:cached,z

  # Private label (most common for app data)
  - ${APPDATA_PATH}/service:/config:Z
```

## What Happens Without SELinux Labels?

If you forget to add SELinux labels to volume mounts:

1. **Containers fail to start** with permission errors
2. **Volumes appear mounted** but are inaccessible
3. **SELinux blocks access** even with correct UID/GID
4. **Audit logs fill up** with AVC denial messages

**Example error:**
```
Error: failed to create shim task: OCI runtime create failed:
container_linux.go:380: starting container process caused:
process_linux.go:545: container init caused:
rootfs_linux.go:76: mounting "/path/to/data" to rootfs at "/config"
caused: permission denied
```

## Verifying SELinux Labels

After starting containers, verify SELinux labels are correct:

```bash
# Check container mounts
ls -laZ /path/to/mounted/directory

# Expected output for :Z (private)
drwxr-xr-x. user user system_u:object_r:container_file_t:s0:c123,c456 config

# Expected output for :z (shared)
drwxr-xr-x. user user system_u:object_r:container_file_t:s0 data
```

## Troubleshooting

### Permission Denied Errors

If containers fail with permission errors:

1. **Check SELinux labels** - Verify `:Z` or `:z` is present on all volume mounts
2. **Check SELinux mode** - Run `getenforce` (should be "Enforcing" on Fedora CoreOS)
3. **Check audit logs** - Look for AVC denials: `sudo ausearch -m avc -ts recent`

### AVC Denial Messages

If you see AVC denials in logs:

```bash
# View recent SELinux denials
sudo ausearch -m avc -ts recent

# Check if a specific path is causing issues
sudo ausearch -m avc | grep "/path/to/directory"
```

**Solution:** Add the appropriate SELinux label (`:Z` or `:z`) to the volume mount.

### Fixing Existing Deployments

If you have running containers without SELinux labels:

1. **Stop the containers:**
   ```bash
   docker-compose down
   # or
   podman-compose down
   ```

2. **Update compose file** with proper `:Z` or `:z` labels

3. **Restart containers:**
   ```bash
   docker-compose up -d
   # or
   podman-compose up -d
   ```

## Environment Variables

These templates use environment variables for configuration. Create a `.env` file in the same directory as your compose file:

```bash
# User/Group Configuration
PUID=1001
PGID=1001
TZ=America/Chicago

# Paths
APPDATA_PATH=/var/lib/containers/appdata

# Service-specific variables
# (see individual compose files for required variables)
```

The `homelab-setup` CLI tool will generate `.env` files automatically with the correct values.

## Additional Resources

- **Fedora CoreOS Documentation:** https://docs.fedoraproject.org/en-US/fedora-coreos/
- **SELinux and Containers:** https://www.redhat.com/en/blog/volume-mounts-selinux
- **Docker SELinux Labels:** https://docs.docker.com/storage/bind-mounts/#configure-the-selinux-label
- **Podman SELinux:** https://docs.podman.io/en/latest/markdown/podman-run.1.html#volume-v-source-volume-host-dir-container-dir-options

## Template Maintenance

When modifying these compose templates:

1. ✅ **ALWAYS** add `:Z` or `:z` to volume mounts
2. ✅ Use `:Z` for private container data
3. ✅ Use `:z` for shared/NFS mounts
4. ✅ Test on a Fedora CoreOS system with SELinux enforcing
5. ✅ Document any new environment variables in comments

**This is not optional.** Missing SELinux labels will cause deployment failures on Fedora CoreOS.
