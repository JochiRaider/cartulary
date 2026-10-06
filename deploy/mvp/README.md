# Cartulary MVP On-Prem Stand-Up Package

This package targets operator-started Windows 11 / WSL2 Ubuntu with Docker Desktop’s WSL2 backend. Native Linux qualification is deferred indefinitely. It runs one Cartulary application image with companion Postgres and SeaweedFS S3-compatible object-storage services.

It is not disconnected-profile conformance. Operational recovery for this package is deployment-local and operator-facing; it is not exposed through public backup, restore, or restore-verification route families.

## Contents

- `Containerfile` builds the app image from Make-built `server`, `migrate`, and `operator` binaries under `build/bin`.
- `docker-compose.yml` starts `app`, `postgres`, `seaweedfs-s3`, one-shot `migrate`, and one-shot `object-store-init`.
- `config.toml.example` is the deployment config template mounted at `/etc/cartulary/config.toml`.
- `revisions-conflict-token-key-ring.json.example` is the dedicated sealed conflict-token key-ring template.
- `.env.example` carries service-binding environment names and placeholder values.
- `bootstrap-admin.json.example` is the first deployment-admin bootstrap manifest template.
- `scripts/backup-capture.sh` runs deployment-local backup creation through the package image.
- `scripts/restore-verify-due.sh` runs due restore verification against an isolated target.
- `restore-verification-target.toml.example` and `restore-verification-target.marker.json.example` are the isolated target examples.
- `systemd/` contains non-secret service and timer templates for package-local scheduling.

## Configure

```sh
cp deploy/mvp/.env.example deploy/mvp/.env
cp deploy/mvp/config.toml.example deploy/mvp/config.toml
cp deploy/mvp/bootstrap-admin.json.example deploy/mvp/bootstrap-admin.json
cp deploy/mvp/revisions-conflict-token-key-ring.json.example deploy/mvp/revisions-conflict-token-key-ring.json
cp deploy/mvp/restore-verification-target.toml.example deploy/mvp/restore-verification-target.toml
```

Before starting, replace the S3 credentials, `CARTULARY_AUTH_MASTER_KEY`, `CARTULARY_RECOVERY_MASTER_KEY`, `CARTULARY_SECRET_REVISIONS_CONFLICT_TOKEN_ACTIVE`, bootstrap admin password, and restore-verification target values. The authentication master key must decode to exactly 32 bytes. The Recovery key also targets exactly 32 bytes in the current-format release; historical Recovery state requires its matching release. The Revisions conflict-token secret must be unpadded base64url that decodes to exactly 32 bytes and must not reuse authentication, recovery, storage, or another subsystem's material.

The config template uses `deployment_profile = "on_prem"` with managed service refs:

- `roots.database_storage.service_ref = "primary"` selects
  `CARTULARY_POSTGRES_PRIMARY_MIGRATION_DSN`,
  `CARTULARY_POSTGRES_PRIMARY_RUNTIME_DSN`, or
  `CARTULARY_POSTGRES_PRIMARY_RECOVERY_DSN` according to the process purpose.
- `roots.object_storage.service_ref = "primary"` selects `CARTULARY_S3_PRIMARY_*`.

The compose file mounts `config.toml` and the Revisions key-ring manifest under `/etc/cartulary` and sets absolute `CARTULARY_CONFIG_FILE=/etc/cartulary/config.toml`. Conflict-token rotation requires exactly one `active` key and at most seven `decrypt_only` keys. A decrypt-only entry records canonical UTC `deactivated_at` and `retire_at`; `retire_at` must be at least 31 minutes later and remain in the future. Replace active material by adding the old key as decrypt-only with its unchanged key ID and secret reference, adding a new active key, and restarting. Outstanding tokens expire after 30 minutes; only cft4 tokens and key-ring v2 are accepted.

The Auth and Recovery master keys are deployment bindings, not rotating key rings.
Do not replace either environment value in place and expect existing sealed MFA,
enterprise transactions or backups to remain readable. This release has no master
key conversion or overlapping legacy reader. Provision independent random keys on
a fresh deployment when changing these bindings, and retain the old release and
keys in the separately secured environment that owns its historical backups.
Use supported domain export/import only for data it explicitly carries; it does
not transfer authentication credentials or make old encrypted backups compatible.
Network Flow and Revisions key-ring rotation follows their owner-specific active
and decrypt-only lifetimes; clients reload when short-lived tokens expire.

The restore-verification target template uses separate `restore_verify` service
refs for Postgres and object storage. Keep `RESTORE_VERIFY_POSTGRES_DB` and
`CARTULARY_S3_RESTORE_VERIFY_*` isolated from the source database and source
bucket. Compose derives purpose-specific target migration and Recovery DSNs
from the target database name and distinct client certificate bindings. Passwords,
passfiles, service files and inherited `PG*` connection variables are rejected.

Administrators must provision the fixed `NOLOGIN` roles
`cartulary_schema_owner`, `cartulary_runtime`, and `cartulary_recovery`, plus
one `NOINHERIT` deployment login for each purpose. The package provisioning
script also installs exact `public` prerequisites `pgcrypto` 1.3 and `citext`
1.6, transfers `public` to the schema owner, and closes database, schema,
extension, and default privileges before migration. Application migrations
validate these prerequisites but never create extensions or roles.

## Certificates and fresh provisioning

Provision certificates in an absolute private directory on the WSL guest
filesystem, set `CARTULARY_TLS_DIR`, and retain its signing authority outside
the package. Do not place keys in a Windows-mounted shared source directory.
Use the admitted P-256/P-384 or RSA parameters and SHA-256-or-stronger signature
policy. Configure the HTTPS public origin to a DNS name or IP in the application
certificate. Each private key is distinct:

| File stem | Purpose and certificate identity |
| --- | --- |
| `application` | Server authentication; SAN covers the operator’s HTTPS/WSS origin. |
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
required SANs, verify validity and chain, stop the affected service, replace the
files atomically, and recreate that service so bind mounts and captured TLS
identities are renewed. A root rotation first deploys an overlapping root bundle,
then replaces leaves, then removes the retired root. Use the restart checks below.
Expired, untrusted, wrong-name or wrong-purpose identities fail closed. Never
substitute a verification bypass. Test-only PKI is short-lived and is destroyed
with its disposable workspace; it is not production provisioning.

## Build

Run from the repository root after configuration:

```sh
make build
docker compose --env-file deploy/mvp/.env -f deploy/mvp/docker-compose.yml build app
```

Supported Make builds pin Go 1.27.1 and cryptographic module selector
`v1.0.0-c2097c7c`, verify the archive digest, and write a binary-bound
`.crypto.json` receipt. Each facade independently admits its actual executable
identity and enabled mode before loading configuration or opening services.
`GODEBUG=fips140=off`, ordinary builds and mismatched metadata reject; there is
no runtime bypass. Strict mode is a diagnostic, not the deployment policy.

`migrate up` establishes `cartulary.application_crypto_format.v1` only after
serialized fresh initialization, with read-only inspection of the actual database,
object bucket and all four filesystem roots before any migration writes. The same
exclusion remains held through migration and identity commit. Server and operator
commands require this committed identity; they never create or repair it.
A missing identity with retained state, including an interrupted initialization,
rejects before leases, bootstrap, bucket creation or jobs. Preserve that state for
investigation or its matching historical release; provision a separate fresh
target. An interruption that left the entire target empty may retry. Module
upgrades require renewed qualification and never redefine the stored-format ID.

The final image contains only `cartulary-server`, `cartulary-migrate`, and `cartulary-operator` plus runtime base image files. It must not contain Node, pnpm, Vite, `apps/web`, `db/migrations` source files, or a repository checkout.

## Start

The `app` service depends on healthy Postgres, successful migration, and successful object-store initialization. Starting `app` is the normal package path:

The packaged database baseline is exact PostgreSQL 18.6 from
`docker.io/library/postgres:18.6-alpine@sha256:d3e1620b530c944afa6e887d22eb899824da68e19c52024bf98f5220c88a65b2`.
Compose mounts the versioned `cartulary-postgres-data-v18` volume at the
PostgreSQL parent directory `/var/lib/postgresql`, while the server stores the
cluster under `PGDATA=/var/lib/postgresql/18/docker`. Initialization enables
data checksums and certificate-only host authentication over verified TLS 1.3; platform admission rejects a
different server patch, checksums-off cluster, or wrong-purpose login before
startup work proceeds.

```sh
docker compose --env-file deploy/mvp/.env -f deploy/mvp/docker-compose.yml up -d app
```

After startup, check:

```sh
curl --cacert "$CARTULARY_TLS_DIR/ca.pem" --tlsv1.3 -fsS https://localhost:8080/healthz
curl --cacert "$CARTULARY_TLS_DIR/ca.pem" --tlsv1.3 -fsS https://localhost:8080/readyz
curl --cacert "$CARTULARY_TLS_DIR/ca.pem" --tlsv1.3 -fsS https://localhost:8080/
```

`/healthz` is process liveness. `/readyz` is structured readiness and returns HTTP 200 only when active dependencies are ready.

To rerun the deployment-local object-store initialization explicitly:

```sh
docker compose --env-file deploy/mvp/.env -f deploy/mvp/docker-compose.yml run --rm object-store-init
```

This command creates or confirms the configured bucket. App startup still fails closed when the configured managed-service bucket is missing.

## Stop

```sh
docker compose --env-file deploy/mvp/.env -f deploy/mvp/docker-compose.yml down
```

Use `docker compose --env-file deploy/mvp/.env -f deploy/mvp/docker-compose.yml down -v` only when intentionally deleting package data volumes.

## Persistent Roots

The package persists state in Docker-managed named volumes:

- `cartulary-postgres-data-v18`
- `cartulary-seaweedfs-data`
- `cartulary-backups`
- `cartulary-reference-packs`
- `cartulary-tmp`
- `cartulary-exports`

These roots are persistent package storage, not source-tree runtime paths.

The image contains empty nonroot-owned runtime-root directories so Docker named volumes are writable by the nonroot app process. Its package-specific build context excludes source-control `.keep` files. Migration receives those same volumes read-only and the source object-store binding for compatibility inspection; it never treats private image paths as deployment storage. Restore-target migration likewise receives the target storage binding and a read-only target-root mount.

## Fresh database provisioning and rejected state

This release requires a fresh exact PostgreSQL 18.6 target and current application
cryptographic-format identity. It has no in-place PostgreSQL cutover, historical
application-state conversion or compatibility stamping procedure. Keep historical
deployments and their backups with their matching release and keys.

When startup rejects an existing or interrupted target, preserve its bounded
error and state for investigation. Correct provisioning inputs and create a
separate, genuinely empty target. Do not run a historical baseline-reset procedure,
relabel a format record, or reuse nonempty object and filesystem roots to bypass
admission. Delete failed initialization state only when it is explicitly disposable
and every database, volume and bucket is confirmed to belong to that attempt.
Ordinary startup never performs this deletion.

Current-format logical Recovery artifacts restore only into pristine targets
admitted by the current release, with separate purpose certificates and confined
storage. A PostgreSQL data directory is never an application recovery artifact.

## Optional Reference Pack Administration

Base registries load without claiming administration. To enable administration,
append `reference-pack-administration.toml.example` to `config.toml` once and
supply `reference-pack-trust.json` using the
`cartulary.reference_pack_trust_bootstrap.v1` contract. Install only roots whose
repository identity and signing keys you have independently approved. The
package ships no trust roots or signing keys. Set `clock_trusted = true` only
after establishing trusted UTC; the example deliberately leaves it false.

Use `docker-compose.reference-packs.yml` together with the base Compose file
for every operation on this deployment. Create `reference-pack-incoming` before
running the operator. Trust, config, and incoming files must be regular files,
readable by container UID 65532; incoming directories must be traversable by
that UID. Grant only the required read access. Bind mounts reject absent host
paths; trust and incoming bundles are read-only inside the containers. Published
content and temporary work use the same named volumes as the server.

```sh
docker compose --env-file deploy/mvp/.env \
  -f deploy/mvp/docker-compose.yml \
  -f deploy/mvp/docker-compose.reference-packs.yml up -d app

docker compose --env-file deploy/mvp/.env \
  -f deploy/mvp/docker-compose.yml \
  -f deploy/mvp/docker-compose.reference-packs.yml run --rm --no-deps \
  reference-pack-operator reference-pack import approved-pack.zip
```

The server must be running and ready: its durable Job worker verifies admitted
imports. The operator accepts a confined bundle filename from the incoming
mount, observes that Job, and returns `cartulary.reference_pack_operator_result.v1`.
Import does not activate a pack. An authorized administrator must explicitly
activate a verified version through the existing administration interface.

For backup/restore wrapper scripts, set `CARTULARY_MVP_COMPOSE_OVERLAY` to the
absolute overlay path in the deployment environment so app restart preserves
these mounts. Stop and restart with the same Compose file pair. Retain the
configured roots, trust history, and backup artifacts together; removing a
claim or changing bootstrap roots is not a conversion of retained state.

## Operational Recovery

Backup creation and restore verification run through `cartulary-operator` inside the package image using the Core logical commands. They require `CARTULARY_RECOVERY_MASTER_KEY`; recovery CLI invocation is deployment-local operator behavior and is not authorized through a runtime `deployment_admin`.

All encrypted backup artifacts, including small metadata files, use
`cartulary.backup_artifact_envelope.v3`. Each artifact derives a fresh key;
4-MiB chunks authenticate order, length and finality, with at most 1,048,576
chunks. Reads authenticate one complete input into private confined staging
before exposing content. The backup filesystem must support private unnamed
files; the reference profile uses guest ext4, not a Windows-mounted path.

Integrity manifests use v4, target markers and verification proofs use v5,
and journal envelopes use v2 with payload v5. All current proofs bind
`cartulary.application_crypto_format.v1`. Historical backups, deployments and
keys must stay with their matching release. There is no converter, legacy
reader or in-place upgrade. Relabeling a marker or manifest cannot establish
compatibility. Preserve the release image and key material needed to read each
retained backup. Replacing a Recovery key requires a separately provisioned
fresh deployment/storage set and a new verified backup; do not discard the
old key while its matching backups remain retained.

The logical PostgreSQL payload keeps
`cartulary.postgres_snapshot_artifact.v2` inside the current authenticated
format. That inner schema does not make historical envelopes compatible and
does not promise cross-engine portability. Target engine, checksum, and purpose-role admission completes before the
first restore mutation. A rejected or interrupted target remains non-serving
and must be cleaned or reinitialized before reuse.

Manual backup creation:

```sh
mkdir -p deploy/mvp/runtime
deploy/mvp/scripts/backup-capture.sh > deploy/mvp/runtime/backup-capture.json
```

The backup script stops the `app` service, runs `operator backup create --source-config-file /etc/cartulary/config.toml`, and restarts the same stopped `app` container in cleanup only if it was running on entry. The JSON result is a single `cartulary.operator_recovery_result.v1` object with the `backup_set_id`, `consistency_point_at`, and non-secret logical artifact references. If the recovery key is missing, if existing encrypted backup artifacts cannot be read with the supplied key, or if publication fails before success, the operator fails closed and any candidate remains diagnostic-only rather than a successful retained backup.

Inspect latest backup metadata:

```sh
set -a
. deploy/mvp/.env
set +a

docker compose --env-file deploy/mvp/.env -f deploy/mvp/docker-compose.yml run --rm --no-deps \
  --entrypoint /usr/local/bin/cartulary-operator \
  recovery-operator backup inspect latest \
  --source-config-file /etc/cartulary/config.toml
```

Manual due restore verification:

```sh
mkdir -p deploy/mvp/runtime
deploy/mvp/scripts/restore-verify-due.sh > deploy/mvp/runtime/restore-verify-due.json
```

The restore-verification script creates or confirms the target database,
migrates and admits the fresh target database, initializes the target object-store bucket, and
then writes a fresh target-generation proof plus a bound
`cartulary.restore_target_marker.v5` under the target backup root before it
runs `cartulary-operator restore-verify due`. The target config, target root,
target database, and target bucket must remain isolated from production state.
The v5 marker binds the application cryptographic format, database, object store, Reference Pack storage, and export-output storage.
For a customized target reference root, set
`CARTULARY_RESTORE_VERIFY_REFERENCE_PACK_BINDING_IDENTITY` to
`filesystem_root:` followed by its canonical container path. Unsafe, expired,
wrongly bound, or unmarked targets are rejected before mutation.

If wrapper scripts must join an existing non-default Compose project, set `CARTULARY_MVP_COMPOSE_PROJECT_NAME` before invoking them.

## Systemd Scheduling

The systemd templates are examples and contain no secrets. Adjust `/opt/cartulary/deploy/mvp` paths if the package is installed elsewhere, and store the secret environment file outside the repository checkout:

```sh
sudo install -D -m 0600 deploy/mvp/.env /etc/cartulary/mvp.env
sudo install -D -m 0644 deploy/mvp/systemd/cartulary-backup.service /etc/systemd/system/cartulary-backup.service
sudo install -D -m 0644 deploy/mvp/systemd/cartulary-backup.timer /etc/systemd/system/cartulary-backup.timer
sudo install -D -m 0644 deploy/mvp/systemd/cartulary-restore-verify.service /etc/systemd/system/cartulary-restore-verify.service
sudo install -D -m 0644 deploy/mvp/systemd/cartulary-restore-verify.timer /etc/systemd/system/cartulary-restore-verify.timer
sudo systemctl daemon-reload
sudo systemctl start cartulary-backup.timer cartulary-restore-verify.timer
systemctl list-timers 'cartulary-*'
```

`cartulary-backup.timer` runs backup creation every 6 hours. That interval is recommended operator practice, not a Core conformance interval. Deployment-owned scheduling must still run `operator backup create` or this package wrapper often enough to keep at least one successful retained backup no older than 24 hours. `cartulary-restore-verify.timer` runs due restore verification daily.

## Operator restart and overdue work

The timers have no boot-install target; start them explicitly for each operating
session after checks pass. They do not depend on a guest `docker.service`.
Stop both timers before intentionally stopping the package. Do not configure
Windows startup tasks, cron or automatic guest boot launch for this package.
Docker Desktop must already be running with WSL integration available.

After sleep, shutdown, guest restart or Docker Desktop interruption, run:

```sh
deploy/mvp/scripts/operation-start.sh
```

This starts the selected package, checks `operator backup inspect latest` and
runs due restore verification. It returns failure if either owner check fails;
downtime does not extend the 24-hour successful-backup freshness limit. On a
fresh deployment, capture the first backup and rerun the checks before starting
timers. On an overdue or failed deployment, inspect the bounded owner error,
resolve the cause, capture a current backup when needed and rerun due verification.
A returned failure is not an accepted operating state. Do not delete old backups
or advance verification records to clear the failure. Historical backups require
their matching release; the current codec has no legacy reader or converter.

Tests interrupt only services they own. They never reboot Windows, shut down the
user’s WSL VM, or alter global Windows trust or firewall settings.

## Package Validation

The package-shape smoke gate is:

```sh
make standup-package-smoke
```

It builds the image, runs the Compose topology, applies migrations, initializes the object store, checks `/healthz` and `/readyz`, verifies embedded `/` and `/assets/*`, checks persistent Docker-volume roots, proves no Vite/source-tree runtime dependency, and checks WebSocket Origin behavior. It is package smoke evidence only. It is not disconnected-profile conformance and is not backup/restore conformance.

`make deployable-shape` remains the narrower static deployable-shape check.

The operational recovery smoke gate is:

```sh
make standup-operational-recovery-smoke
```

It builds and runs the MVP Compose package, creates a backup, inspects latest metadata through the canonical result envelope, runs due restore verification against an isolated target, proves the public backup/restore route families are absent, and retains summary artifacts. It is operational package evidence only and is not disconnected-profile conformance.

## Troubleshooting

- If `app` restarts with `path_not_writable`, confirm the package image is current and the runtime roots are Docker named volumes, not host paths from the source tree.
- If `object-store-init` fails, check `CARTULARY_S3_PRIMARY_ENDPOINT`, `CARTULARY_S3_PRIMARY_SECURE`, credentials, and whether `seaweedfs-s3` is running.
- If `/readyz` returns a non-200 response, inspect the structured readiness status and the `postgres` and `seaweedfs-s3` service health.
- If migration fails, inspect `docker compose logs migrate postgres` and verify
  `CARTULARY_POSTGRES_PRIMARY_MIGRATION_DSN` resolves to the package Postgres
  service. A `prod_ddl_rebaseline_v2` report rejects an incompatible database; preserve
  retained state with its matching release and provision a separate fresh target.
  A server-version or checksum admission error requires a fresh exact
  PostgreSQL 18.6 baseline, not a compatibility override.
- If browser WebSocket requests fail with HTTP 403, verify `CARTULARY_PUBLIC_ORIGIN` exactly matches the browser origin used to reach the app.
- If startup reports a `revisions_conflict_token_*` diagnostic, verify the key-ring mount, exact manifest schema, one-active-key rotation state, unique IDs and secret references, and the 32-byte unpadded-base64url secret. Startup intentionally fails before listeners when this credential is unavailable.
- If backup creation fails, inspect the script stderr and confirm the Recovery certificate has the correct purpose, the app service can be stopped and restarted, and `CARTULARY_RECOVERY_MASTER_KEY` matches existing encrypted backup artifacts.
- If restore verification fails before mutation, confirm the target config
  differs from the source config, the target database and object-store bucket
  are isolated, and both `restore-target-marker.json` and
  `restore-target-generation` are present under the target backup root.
- If restore verification reports a failed item, retain the JSON output and inspect the target `postgres`, object-store, migration, and operator logs before deleting target state.

Restore verification markers bind the database, object store, Reference Pack root, and export-output root. Create current markers only for freshly initialized, admitted targets; historical state and markers require their matching release. A custom export root uses `CARTULARY_RESTORE_VERIFY_EXPORT_BINDING_IDENTITY` with its `filesystem_root:` identity. Recovery captures exported Incident Bundles from the configured export root and restores them into the isolated target export root.

## Windows 11 / WSL2 qualification and upgrades

Run the repository's full `make release-check` on the selected Windows 11 host,
Ubuntu guest and Docker Desktop WSL2 backend. Native Linux qualification is
indefinitely deferred. The release requires all three package smokes, actual
binary admission and fixed credential-capacity assessments, plus the full owner,
browser, accessibility, visual and release membership. The configured Markdown
lint does not include this README or the remediation handoff; review them directly.

Package evidence records Windows build, WSL/kernel, Ubuntu, CPU features, observed
VM/container limits, Docker Desktop/Engine/Compose/daemon identity, guest ext4
storage, WSL network mode, resolved images and the three binary/module/archive
receipts. Zero container limits mean no narrower container limit than the observed
VM; the credential assessment separately requires effective two-CPU/two-GiB
cgroup limits. The Make-built Windows executable uses only an in-memory private CA
pool supplied over stdin. Its HTTPS/WSS evidence is separate from Linux browser
results. It tests chain/name/expiry and TLS-version rejection without changing
Windows trust or the host clock. Qualification replaces purpose certificates and
recreates only disposable owned services, then checks current-state operation.

Keep the checkout, private key material and package storage on the guest ext4
filesystem, outside Windows-mounted drives. Docker Desktop named volumes reside
in the recorded daemon's managed storage. Back up through the application's
current authenticated codec; copying live database or object-store directories
is not a recovery procedure. Preserve the exact release and key material with
historical backups. Do not stamp, migrate or reset a rejected retained deployment
with this release; use its matching historical release in a separately isolated
environment. If fresh initialization was interrupted after any retained state
was created, diagnose the failure and provision a genuinely new, empty target.
Only explicitly disposable failed-initialization state may be destroyed.

After host sleep or shutdown, first confirm Docker Desktop availability and
Windows/guest time synchronization. Run the operator-started checks above before
starting guest timers. A stale backup or failed due verification blocks accepted
operation; downtime grants no freshness exception. Renewal requires new keys and
purpose-correct certificates, renewed mounts and service recreation. Root changes
require an overlap period and verification from both Windows and guest clients.

Any Windows, WSL, Docker Desktop, daemon platform, PostgreSQL, SeaweedFS, Go or
module upgrade creates a new qualification candidate. Verify upstream identity
and support, update authored pins and downstream generation, then repeat the
service/verifier/package and complete release gates on fresh disposable state.
Keep application-format identity independent of module version. One successful
source test or image label cannot qualify a platform upgrade. The current single
policy permits certificate-only PostgreSQL with TLS 1.3, HTTPS S3 with SigV4 and
SHA-256 checksums, the bounded OIDC/SAML algorithms, and verified HTTPS telemetry;
there is no password, plaintext, legacy codec or disabled-module fallback.

Engineering completion, WSL2 package acceptance, formal CMVP applicability,
specification adoption and customer deployment approval are separate decisions.
A package pass does not establish formal CMVP applicability or authorize a customer
deployment. Formal applicability remains unestablished unless independently
reviewed for the exact module, approved services and observed environment.
