# Cartulary MVP On-Prem Stand-Up Package

This package is an MVP on-prem local stand-up package. It runs one Cartulary application image with companion Postgres and SeaweedFS S3-compatible object-storage services.

It is not disconnected-profile conformance. It allocates a persistent backup root, but it does not claim backup conformance because backup capture, retention, and restore-verification scheduling are separate requirements.

## Contents

- `Containerfile` builds the app image from Make-built `server`, `migrate`, and `operator` binaries.
- `docker-compose.yml` starts `app`, `postgres`, `seaweedfs-s3`, one-shot `migrate`, and one-shot `object-store-init`.
- `config.toml.example` is the deployment config template mounted at `/etc/cartulary/config.toml`.
- `.env.example` carries service-binding environment names and placeholder values.
- `bootstrap-admin.json.example` is the first deployment-admin bootstrap manifest template.

## Configure

```sh
cp deploy/mvp/.env.example deploy/mvp/.env
cp deploy/mvp/config.toml.example deploy/mvp/config.toml
cp deploy/mvp/bootstrap-admin.json.example deploy/mvp/bootstrap-admin.json
```

Before starting, replace the placeholder passwords, S3 credentials, `CARTULARY_AUTH_MASTER_KEY`, `CARTULARY_RECOVERY_MASTER_KEY`, and bootstrap admin password. The two Cartulary master keys must be base64-encoded values that decode to at least 32 bytes.

The config template uses `deployment_profile = "on_prem"` with managed service refs:

- `roots.database_storage.service_ref = "primary"` selects `CARTULARY_POSTGRES_PRIMARY_DSN`.
- `roots.object_storage.service_ref = "primary"` selects `CARTULARY_S3_PRIMARY_*`.

The compose file mounts `config.toml` at `/etc/cartulary/config.toml` and sets absolute `CARTULARY_CONFIG_FILE=/etc/cartulary/config.toml`.

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

## Package Validation

The repository-owned smoke gate is:

```sh
make standup-package-smoke
```

It builds the image, runs the Compose topology, applies migrations, initializes the object store, checks `/healthz` and `/readyz`, verifies embedded `/` and `/assets/*`, checks persistent Docker-volume roots, proves no Vite/source-tree runtime dependency, and checks WebSocket Origin behavior. It is package smoke evidence only. It is not disconnected-profile conformance and is not backup/restore conformance.

`make deployable-shape` remains the narrower static deployable-shape check.

## Troubleshooting

- If `app` restarts with `path_not_writable`, confirm the package image is current and the runtime roots are Docker named volumes, not host paths from the source tree.
- If `object-store-init` fails, check `CARTULARY_S3_PRIMARY_ENDPOINT`, `CARTULARY_S3_PRIMARY_SECURE`, credentials, and whether `seaweedfs-s3` is running.
- If `/readyz` returns a non-200 response, inspect the structured readiness status and the `postgres` and `seaweedfs-s3` service health.
- If migration fails, inspect `docker compose logs migrate postgres` and verify `CARTULARY_POSTGRES_PRIMARY_DSN` resolves to the package Postgres service.
- If browser WebSocket requests fail with HTTP 403, verify `CARTULARY_PUBLIC_ORIGIN` exactly matches the browser origin used to reach the app.
