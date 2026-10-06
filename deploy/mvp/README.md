# Cartulary local package — Windows 11 / WSL2

The supported package runs on operator-started Windows 11 with WSL2 Ubuntu and
Docker Desktop's WSL2 backend. It contains one application image (server,
migrate, operator) and pinned PostgreSQL and SeaweedFS images. Native Linux
qualification, registry publication and automatic upgrades are outside scope.

## Install the matching release

Obtain the archive and its expected SHA-256 through your trusted delivery
channel. A checksum establishes identity; a checksum supplied alongside an
untrusted archive does not establish trust. Keep that exact archive with each
independent backup set. No checkout, Go, Node, pnpm or network image pull is
required to install it. The WSL guest needs Bash, Docker/Compose, jq, curl,
GNU coreutils/findutils/tar, util-linux (`flock`) and a running Docker Desktop
linux/amd64 backend. Provision those host prerequisites before offline use.

Verify the delivered archive checksum before extraction. Extract into a private,
empty directory on the WSL filesystem, then run the installer from that archive:

```sh
sha256sum /absolute/delivery/PACKAGE_SHA256.tar
tar --no-same-owner --same-permissions -xf /absolute/delivery/PACKAGE_SHA256.tar -C /absolute/private/extraction
/absolute/private/extraction/release/assets/scripts/install.sh /opt/cartulary
```

Use an existing canonical parent and an unoccupied installation path. The installer
verifies all bound assets, inventories and receipts, loads the included images,
checks their identities and platform, and publishes a new installation. It rejects
occupied destinations and contaminated immutable image references. Failed attempts
never replace an existing installation. Loaded immutable images may remain for a
retry; they are release artifacts, not application state.

`/opt/cartulary/release/` contains immutable payload bytes. Do not edit them or
substitute images. `/opt/cartulary/.env`, configuration, bootstrap/key-ring files,
TLS material and `/opt/cartulary/runtime/` are installation state outside that
payload. Keep the release manifest, image archives, receipts and inventories.
Every package operation verifies the release again before changing services.

For maintainers, `make package-release` produces a manifest-addressed archive under
`.cartulary/release-artifacts/packages/`; its canonical run receipt is
`package-release/artifacts/package.json`. `make package-inspect` rehearses the
installer against that exact producer output. Consult `make help-all` for the
current public qualification routes. Qualification evidence stays outside the
release manifest, avoiding circular identity. These commands never read this guide.

## Configure one deployment

The installer copies examples to the installation root. Replace all placeholders
before first start: separate random authentication and Recovery master keys,
Revisions token material, S3 credentials, first administrator credentials, public
origin, purpose TLS identities and source/verification namespace names. Protect
settings and keys with owner-only host permissions and the specific container
read permissions described below. Do not ship configured secrets in a release.

Mounted configuration, bootstrap and key-ring files must be readable by container
GID 65532 (for example owner:65532 with mode 0640). Keep the installation directory
private and `.env` owner-only; it is read by the host entrypoint, not mounted.

Settings are literal `KEY=value` lines in `.env`, with blank lines and whole-line
`#` comments. Never source this file. Shell expansion, duplicate or unknown keys,
placeholder values and ambient overrides are rejected. There is no configurable
application tag, arbitrary Compose-file override, or compatibility parser.
Set `CARTULARY_MVP_COMPOSE_PROJECT_NAME` once to a unique deployment name;
changing it selects different state. Use distinct `POSTGRES_DB` and
`RESTORE_VERIFY_POSTGRES_DB`, distinct source/verification buckets, the shipped
`primary` and `restore_verify` bindings, and canonical non-overlapping roots.
Preflight loads application configuration through its owner before provisioning.

`CARTULARY_REFERENCE_PACKS_ENABLED=true` selects the shipped Reference Pack
overlay for every operation. Configure its administration settings and approved
`reference-pack-trust.json` in the installation root before enabling it. No test
trust is shipped. Required built-in packs, approved historical trust, provenance
and pins retain their existing owner semantics; this package adds no UI workflow.

The authentication and Recovery master keys each decode to 32 bytes. Keep them
separate from each other and from service credentials. Revisions token material
is unpadded base64url encoding of 32 independent random bytes. Database connections
use certificate-only purpose logins; password, passfile, service-file and inherited
`PG*` fallbacks are rejected. Do not bypass owner admission to initialize retained
or partially initialized state.

## Certificates and fresh provisioning

Provision certificates in an absolute private directory on the WSL guest
filesystem, set `CARTULARY_TLS_DIR`, and retain its signing authority outside
the package. Do not place keys in a Windows-mounted shared source directory.
Use the admitted P-256/P-384 or RSA parameters and SHA-256-or-stronger signature
policy. Configure the HTTPS public origin to a DNS name or IP in the application
certificate. Each private key is distinct:

| File stem | Purpose and certificate identity |
| --- | --- |
| `application` | Server authentication; SAN includes `localhost` for package readiness and the operator’s HTTPS/WSS origin hostname. |
| `postgres` | Server authentication; SAN includes `postgres`. |
| `seaweed` | Server authentication; SAN includes `seaweedfs-s3`. |
| `migration` | Client authentication; exact CN `migration`. |
| `runtime` | Client authentication; exact CN `runtime`. |
| `recovery` | Client authentication; exact CN `recovery`. |
| `restore-migration` | Client authentication; exact CN `restore-migration`. |
| `restore-recovery` | Client authentication; exact CN `restore-recovery`. |

Each stem has `.crt` (PEM leaf plus intermediates) and `.key` files; `ca.pem`
contains the admitted trust roots. Provision the directory as root-owned 0700.
Application-mounted private keys must be readable by container GID 65532
(for example root:65532, mode 0440); the directory remains private on the host.
PostgreSQL and SeaweedFS entrypoints copy only their own identity into private
service-owned tmpfs. All source binds are read-only and fail if absent.
The app continues to run as nonroot and receives only its runtime and server keys.

The fixed PostgreSQL HBA permits local peer authentication only for OS/database
user `postgres`; every network application login requires its mapped certificate.
All application login passwords are NULL. The upstream image’s required temporary
initialization password is kernel-random and removed by provisioning before the
ordinary listener starts; it is never accepted over the network. Ordinary startup
never reapplies role provisioning. A partially initialized database must be
investigated and reprovisioned as fresh; do not stamp or repair it into compatibility.

SeaweedFS serves verified HTTPS on its native S3 listener, with no plaintext S3
alternative. Public access remains application-mediated. Restore-target S3 trust
must be included in the scoped `ca.pem` bundle. External storage credentials must
have access only to the intended source or verification bucket.

The packaged SeaweedFS node uses the stable `seaweedfs-s3` Compose DNS name as
its persisted server identity. Preserve that identity and its named volume during
renewal or recreation; a detected container IP is not a durable node identity.
Never reset Raft metadata or recreate the bucket to conceal a restart failure.
Compose waits for the container-local master readiness probe before admitting S3
dependencies; S3 transport is independently verified over TLS by the application.

For certificate renewal, issue replacement identities with the same purpose and
required SANs, verify validity and chain, stop maintenance timers and run the
installed `package.sh stop`. Replace the files atomically, then run `package.sh start`.
Under its deployment lock, startup recreates the companion containers with their
existing named volumes and reopens certificate bind files before protection and
readiness gates. Resume timers only after successful verified readiness. A root rotation first deploys an overlapping root bundle,
then replaces leaves, then removes the retired root. Use the restart checks below.
Expired, untrusted, wrong-name or wrong-purpose identities fail closed. Never
substitute a verification bypass. Test-only PKI is short-lived and is destroyed
with its disposable workspace; it is not production provisioning.

## Start, stop and first protection

Use the installed entrypoint for lifecycle and maintenance:

```sh
/opt/cartulary/release/assets/scripts/package.sh start
/opt/cartulary/release/assets/scripts/package.sh stop
```

Object initialization waits at most 120 seconds for exhausted transient metadata
attempts; configuration and cryptographic rejection remain immediate. Failed
startup reports its gate and cleanup outcome and leaves serving stopped.

Startup acquires deployment-scoped exclusion, admits configuration and release
identity, provisions dependencies, and privately initializes/bootstrap-checks the
application without publishing its port. It stops capture-sensitive work, creates
a missing or stale backup, runs due verification, then starts serving and requires
verified HTTPS readiness. A failed gate leaves serving stopped and identifies the
failed operation. A missing or older-than-24-hour backup is never accepted as a
successful start. This includes first installation and startup after an outage.

The application port publishes only after these gates pass. Use the configured
HTTPS origin in Windows; WSS uses that same trusted origin. Install the intended
public CA trust through your organization's Windows process. Expired certificates,
wrong names or disabled cryptographic execution must be corrected at their owners.

Stop the application before intentionally stopping WSL or Docker Desktop. After
sleep, host restart or Docker restart, start Docker Desktop and WSL explicitly,
then run `package.sh start` and inspect its result and HTTPS readiness. Guest timers
do not wake Windows or turn WSL into an unattended appliance. Keep intended service
and volume identities stable across restarts; do not delete retained state to
conceal a failed readiness or compatibility check.

## Routine maintenance and failed operations

```sh
/opt/cartulary/release/assets/scripts/package.sh backup-create
/opt/cartulary/release/assets/scripts/package.sh restore-verify-due
```

Backup creation stops the exact running application container, captures through
Recovery, then restarts only that container if it was running originally.
An initially stopped application remains stopped. Capture, restart and cleanup
outcomes are separate in the terminal package result. A published successful backup
remains valid when a later service restart fails; inspect both results before retrying.

All installed lifecycle operations share one deployment lock. Recovery also owns
its database/target exclusions. Concurrent operations reject. After process death,
the pending operation record blocks overlap with any surviving operation container.
Inspect the owned container and finish or stop that exact operation before retrying;
never delete its pending record merely to force another operation through.
Cancellation cannot authorize a second capture or reset an indeterminate restore.

Recovery issues and renews target proofs after actual state inspection under its
exclusive target lease. A matching retry keeps generation identity; a due no-op
does not rewrite proof. Successfully verified disposable targets may be reset by
the existing lease-controlled Recovery path. Partial or indeterminate targets are
preserved for investigation. There is no supported marker/hash override or manual
proof construction. Use a separate fresh target when admission rejects retained state.

Copy the non-secret service/timer templates from `release/assets/systemd/` to your
WSL systemd configuration and adjust the installation path if necessary. Enable
timers only for the intended operator-started guest session. Their commands use
this same entrypoint and settings. Stop maintenance timers during certificate or
configuration replacement. Keep secrets in the protected installation file, not
in systemd unit text. Review the timer results after resuming the host.

## Key custody and supported rotation

Keep the Recovery key and matching release separately from exported backups.
An intact encrypted backup without its key is unusable. Keep the authentication
key needed by restored sealed authentication state, and the other owner-required
keys/configuration in separately secured custody. A bundle deliberately excludes
provisioning credentials and deployment keys. Test access to both independent
storage and key custody before a real incident.

Auth and Recovery master keys are deployment bindings, not rotating key rings.
This release provides no in-place master-key converter or historical reader.
Replacing their environment values does not re-encrypt existing MFA, transactions
or backups. Retain original keys and the exact original release for retained backups.
A fresh deployment using new bindings needs an explicit owner-supported data
migration; domain pack export/import does not transfer authentication credentials.

Revisions key-ring v2 permits exactly one active key and at most seven decrypt-only
keys. Preserve the old key ID/material and add canonical UTC deactivation/retirement
metadata when making it decrypt-only; retirement must be at least 31 minutes later
and still in the future. Add independent active material and restart through the
normal protection gates. Outstanding cft4 tokens expire after 30 minutes. Follow
Network Flow's owner rules for its own active/decrypt-only lifetimes. No package
command bypasses these owners or adds legacy emission modes.

## Export an independent encrypted backup

Create an existing private destination parent on genuinely independent storage,
then choose a new absolute directory name:

```sh
/opt/cartulary/release/assets/scripts/package.sh backup-export /absolute/offline-storage/new-bundle
```

The export selects the newest intact successful retained backup, regardless of
operational age. It copies complete existing encrypted artifacts and authenticates
selection, release/catalog/codec identities, sizes and digests in a versioned
manifest. It validates readback and atomically publishes a complete directory;
occupied destinations reject. Failed export cannot change retained source selection.
Copy the entire completed bundle. Never rearrange members or edit the manifest.

The destination must be writable by the application container's nonroot identity
(UID/GID 65532) while remaining private. Source retention expiry does not invalidate
a completed export. A second volume or directory in the same WSL/Docker storage is
not independent disaster protection. Place exports and a verified copy of the exact
release archive outside that failure domain, and test a restore with source services
unavailable. Export success proves byte closure, not the physical independence of
a customer's storage choice.

## Restore after source loss

Install the exact matching archive into a new installation with isolated database,
object and filesystem namespaces. Provision fresh target credentials/certificates
and the separately retained keys/configuration. Leave the application stopped.
Do not run normal startup first: that would bootstrap application state into the
otherwise fresh restore target. Use the backup UUID from the completed export
result and a recorded operation UUID for all retries of this attempt:

```sh
/opt/cartulary/release/assets/scripts/package.sh restore-bundle /absolute/offline-storage/bundle BACKUP_UUID --operation-id OPERATION_UUID
```

For a consistency point strictly older than 24 hours, explicitly acknowledge that
exact backup ID by adding `--acknowledge-stale-backup BACKUP_UUID`. Acknowledgement
accepts data loss since that point; it never bypasses integrity, release, key,
confirmation or target admission. Exactly 24 hours does not require acknowledgement.
Incorrect confirmation or acknowledgement rejects. Completed exports remain usable
after their original source retention deadline.

Portable restore opens transfer storage and admitted target resources only. Source
database, object service, source configuration and original volumes may all be lost.
Restoration invalidates existing login sessions. Sign in again with your retained
credentials and enrolled MFA factor after the fresh-backup readiness gate passes.
Recovery first authenticates and captures immutable input, then uses its one restore
engine. It restores authoritative state and complete object families, rebuilds
projections, and probes workbook usability. Its encrypted local intent/completion
journal stays outside restored database contents. Successful completion reconciles
safe target audit evidence. Preserve that journal and target generation.

Repeat the identical bundle, target, backup ID and operation ID after a lost response.
A completed retry returns original terminal evidence without repeating committed
restore work. Missing/corrupt terminal evidence or incompatible generation may leave
an indeterminate target isolated; preserve it and provision another fresh target.
Never restamp or partially clear a target to make its proof look compatible.

A successful restore leaves the application stopped. Run the normal `package.sh start`
to establish a new fresh backup and required verification before serving. Restore
does not renew the exported consistency point or manufacture fresh verification.
Retain the original bundle and keys until the new installation's independent backup
protection has been checked.

## Replacement and acceptance

Retain the matching release archive for every retained backup generation. There is
no automatic release switching, historical upgrade/converter or compatibility alias.
Replacing lost infrastructure with the same release is a fresh installation followed
by admitted restore. Moving to a different release requires an explicit future
owner-defined migration and renewed qualification; editing tags or receipts is not
an upgrade mechanism.

Inventories scan actual shipped image archives with the pinned Syft tool. Notice
files are included when discoverable; unknown license metadata and unresolved
component/notice association remain explicit review findings. Technical inventory
completeness does not grant distribution permission or replace security/licensing
review. A missing/stale inventory fails technical acceptance.

Record implementation completion, Windows/WSL2 package acceptance, formal CMVP
applicability, specification adoption and customer deployment approval separately.
An enabled cryptographic module and passing engineering tests do not establish
formal environment applicability. This guide is operating support; the controlling
remediation tracker records actual candidate identities, failures, cleanup and
readiness dispositions. No new customer deployment is automatically approved.
