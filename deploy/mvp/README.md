# Cartulary MVP On-Prem Stand-Up Package

This package is an MVP on-prem local stand-up package. It runs one Cartulary application image with companion Postgres and SeaweedFS S3-compatible object-storage services.

It is not disconnected-profile conformance. Operational recovery for this package is deployment-local and operator-facing; it is not exposed through public backup, restore, or restore-verification route families.

## Contents

- `Containerfile` builds the app image from Make-built `server`, `migrate`, and `operator` binaries.
- `docker-compose.yml` starts `app`, `postgres`, `seaweedfs-s3`, one-shot `migrate`, and one-shot `object-store-init`.
- `config.toml.example` is the deployment config template mounted at `/etc/cartulary/config.toml`.
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
cp deploy/mvp/restore-verification-target.toml.example deploy/mvp/restore-verification-target.toml
```

Before starting, replace the placeholder passwords, S3 credentials, `CARTULARY_AUTH_MASTER_KEY`, `CARTULARY_RECOVERY_MASTER_KEY`, bootstrap admin password, `CARTULARY_RECOVERY_DEPLOYMENT_ADMIN_EMAIL`, and restore-verification target values. The two Cartulary master keys must be base64-encoded values that decode to at least 32 bytes.

The config template uses `deployment_profile = "on_prem"` with managed service refs:

- `roots.database_storage.service_ref = "primary"` selects `CARTULARY_POSTGRES_PRIMARY_DSN`.
- `roots.object_storage.service_ref = "primary"` selects `CARTULARY_S3_PRIMARY_*`.

The compose file mounts `config.toml` at `/etc/cartulary/config.toml` and sets absolute `CARTULARY_CONFIG_FILE=/etc/cartulary/config.toml`.

The restore-verification target template uses separate `restore_verify` service refs for Postgres and object storage. Keep `RESTORE_VERIFY_POSTGRES_DB`, `CARTULARY_POSTGRES_RESTORE_VERIFY_DSN`, and `CARTULARY_S3_RESTORE_VERIFY_*` isolated from the source database and source bucket.

## Build

Run from the repository root after configuration:

```sh
make build
docker compose --env-file deploy/mvp/.env -f deploy/mvp/docker-compose.yml build app
```

The final image contains only `cartulary-server`, `cartulary-migrate`, and `cartulary-operator` plus runtime base image files. It must not contain Node, pnpm, Vite, `apps/web`, `db/migrations` source files, or a repository checkout.

## Start

The `app` service depends on healthy Postgres, successful migration, and successful object-store initialization. Starting `app` is the normal package path:

```sh
docker compose --env-file deploy/mvp/.env -f deploy/mvp/docker-compose.yml up -d app
```

After startup, check:

```sh
curl -fsS http://127.0.0.1:8080/healthz
curl -fsS http://127.0.0.1:8080/readyz
curl -fsS http://127.0.0.1:8080/
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

- `cartulary-postgres-data`
- `cartulary-seaweedfs-data`
- `cartulary-backups`
- `cartulary-reference-packs`
- `cartulary-tmp`
- `cartulary-exports`

These roots are persistent package storage, not source-tree runtime paths.

The image pre-creates the app runtime-root mount targets as non-root-owned directories so Docker named volumes are writable by the non-root app process.

## Operational Recovery

Backup creation and restore verification run through `cartulary-operator` inside the package image. The Core logical backup creation command is `operator backup create`; this MVP package script currently wraps the package-local `backup capture` compatibility alias for that logical command. They require `CARTULARY_RECOVERY_MASTER_KEY` and an active `deployment_admin` whose email is set in `CARTULARY_RECOVERY_DEPLOYMENT_ADMIN_EMAIL`.

Manual backup creation:

```sh
mkdir -p deploy/mvp/runtime
deploy/mvp/scripts/backup-capture.sh > deploy/mvp/runtime/backup-capture.json
```

The backup script stops the `app` service, writes a local quiescence proof, runs the package compatibility alias for logical `operator backup create`, and restarts `app` in cleanup. The JSON result includes the `backup_set_id`, `consistency_point_at`, restore anchors, artifact hashes, integrity manifest proof, and retention timestamps. If the recovery key is missing, if existing encrypted backup artifacts cannot be read with the supplied key, or if publication fails before success, the operator fails closed and any candidate remains diagnostic-only rather than a successful retained backup.

Inspect latest backup metadata:

The direct example below uses the current package compatibility alias for logical `operator backup inspect latest`.

```sh
set -a
. deploy/mvp/.env
set +a

docker compose --env-file deploy/mvp/.env -f deploy/mvp/docker-compose.yml run --rm --no-deps \
  --entrypoint /usr/local/bin/cartulary-operator \
  app backup-metadata latest \
  -source-config /etc/cartulary/config.toml \
  -deployment-admin-email "$CARTULARY_RECOVERY_DEPLOYMENT_ADMIN_EMAIL"
```

Manual due restore verification:

```sh
mkdir -p deploy/mvp/runtime
deploy/mvp/scripts/restore-verify-due.sh > deploy/mvp/runtime/restore-verify-due.json
```

The restore-verification script creates or confirms the target database, initializes the target object-store bucket, writes the `cartulary.restore_verification_target.v1` marker under the target backup root, migrates the target database, and runs `cartulary-operator restore-verify due`. The target config, target root, target database, and target bucket must remain isolated from production state. Unsafe or unmarked targets are rejected before mutation.

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
sudo systemctl enable --now cartulary-backup.timer cartulary-restore-verify.timer
systemctl list-timers 'cartulary-*'
```

`cartulary-backup.timer` runs backup creation every 6 hours. That interval is recommended operator practice, not a Core conformance interval. Deployment-owned scheduling must still run `operator backup create` or this package wrapper often enough to keep at least one successful retained backup no older than 24 hours. `cartulary-restore-verify.timer` runs due restore verification daily.

## Cron Scheduling

A cron deployment can run the package-local wrapper instead of systemd. Keep the secret environment file outside the checkout and point the script at that file explicitly:

```cron
SHELL=/bin/sh
0 */6 * * * cartulary CARTULARY_MVP_DIR=/opt/cartulary/deploy/mvp CARTULARY_MVP_ENV_FILE=/etc/cartulary/mvp.env CARTULARY_SOURCE_CONFIG=/opt/cartulary/deploy/mvp/config.toml /opt/cartulary/deploy/mvp/scripts/backup-capture.sh >>/var/log/cartulary/backup-create.log 2>&1
```

Use the same restore-verification target isolation described above before adding a cron entry for `scripts/restore-verify-due.sh`.

## Container-Job Scheduling

Container schedulers should run either the package-local wrapper or the Core logical command shape. A generic one-shot backup creation container job should resolve the effective source deployment configuration through the mounted deployment config and secret references, then run:

```sh
/usr/local/bin/cartulary-operator backup create --output=json --progress=jsonl
```

For the MVP Compose package, prefer the host-side wrapper because it stops and restarts the `app` service and writes the quiescence proof required by the current package realization:

```sh
CARTULARY_MVP_DIR=/opt/cartulary/deploy/mvp \
CARTULARY_MVP_ENV_FILE=/etc/cartulary/mvp.env \
CARTULARY_SOURCE_CONFIG=/opt/cartulary/deploy/mvp/config.toml \
/opt/cartulary/deploy/mvp/scripts/backup-capture.sh
```

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

It builds and runs the MVP Compose package, creates a backup, inspects latest metadata, checks the retention floor, runs due restore verification against an isolated target, proves the public backup/restore route families are absent, and retains summary artifacts. It is operational package evidence only and is not disconnected-profile conformance.

## Troubleshooting

- If `app` restarts with `path_not_writable`, confirm the package image is current and the runtime roots are Docker named volumes, not host paths from the source tree.
- If `object-store-init` fails, check `CARTULARY_S3_PRIMARY_ENDPOINT`, `CARTULARY_S3_PRIMARY_SECURE`, credentials, and whether `seaweedfs-s3` is running.
- If `/readyz` returns a non-200 response, inspect the structured readiness status and the `postgres` and `seaweedfs-s3` service health.
- If migration fails, inspect `docker compose logs migrate postgres` and verify `CARTULARY_POSTGRES_PRIMARY_DSN` resolves to the package Postgres service.
- If browser WebSocket requests fail with HTTP 403, verify `CARTULARY_PUBLIC_ORIGIN` exactly matches the browser origin used to reach the app.
- If backup creation fails, inspect the script stderr and confirm the configured deployment admin is active, the app service can be stopped and restarted, and `CARTULARY_RECOVERY_MASTER_KEY` matches existing encrypted backup artifacts.
- If restore verification fails before mutation, confirm the target config differs from the source config, the target database and object-store bucket are isolated, and the target marker JSON is present under the target backup root.
- If restore verification reports a failed item, retain the JSON output and inspect the target `postgres`, object-store, migration, and operator logs before deleting target state.
