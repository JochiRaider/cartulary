#!/usr/bin/env bash
# Private composition function; caller holds package exclusion.
package_verify() {
  OBJECT_INIT_OUTPUT=/dev/null
  fail() { package_fail "$1"; }
  target_db="$RESTORE_VERIFY_POSTGRES_DB"
  if [[ ! "$target_db" =~ ^[A-Za-z_][A-Za-z0-9_]*$ ]]; then
    fail "RESTORE_VERIFY_POSTGRES_DB must be a simple database identifier"
  fi
  for name in \
    CARTULARY_S3_RESTORE_VERIFY_ENDPOINT \
    CARTULARY_S3_RESTORE_VERIFY_ACCESS_KEY_ID \
    CARTULARY_S3_RESTORE_VERIFY_SECRET_ACCESS_KEY \
    CARTULARY_S3_RESTORE_VERIFY_BUCKET; do
    if [[ -z "${!name:-}" || "${!name:-}" == replace-* ]]; then
      fail "set ${name} in ${ENV_FILE}"
    fi
  done
  
  if [[ ! -d "$TARGET_ROOT_HOST" ]]; then
    mkdir -p "$TARGET_ROOT_HOST"
    compose run --rm --no-deps --user 0 \
      --volume "${TARGET_ROOT_HOST}:/target" --entrypoint /bin/sh postgres \
      -c 'chown 65532:65532 /target && chmod 0700 /target'
  fi
  if ! compose exec -T --user postgres postgres psql -U postgres -d postgres -Atc "SELECT 1 FROM pg_database WHERE datname = '${target_db}'" | grep -qx "1"; then
    compose exec -T --user postgres postgres createdb -U postgres "$target_db"
    compose exec -T --user postgres -e "PGDATABASE=${target_db}" postgres /docker-entrypoint-initdb.d/010-cartulary-provision.sh >/dev/null

  compose run --rm --no-deps \
    --volume "${TARGET_CONFIG_HOST}:${TARGET_CONFIG_CONTAINER}:ro" \
    --volume "${TARGET_ROOT_HOST}:${TARGET_ROOT_CONTAINER}:ro" \
    --entrypoint /usr/local/bin/cartulary-migrate \
    restore-verify-migrate up >/dev/null
  fi
  
  package_initialize_objects run --rm --no-deps \
    --volume "${TARGET_CONFIG_HOST}:${TARGET_CONFIG_CONTAINER}:ro" \
    --volume "${TARGET_ROOT_HOST}:${TARGET_ROOT_CONTAINER}" \
    --entrypoint /usr/local/bin/cartulary-operator \
    restore-verify-operator object-store init -config "$TARGET_CONFIG_CONTAINER" >"$OBJECT_INIT_OUTPUT"
  
  compose run --rm --no-deps \
    --volume "${SOURCE_CONFIG_HOST}:${SOURCE_CONFIG_CONTAINER}:ro" \
    --volume "${TARGET_CONFIG_HOST}:${TARGET_CONFIG_CONTAINER}:ro" \
    --volume "${TARGET_ROOT_HOST}:${TARGET_ROOT_CONTAINER}" \
    --entrypoint /usr/local/bin/cartulary-operator \
    restore-verify-operator restore-verify due \
    --source-config-file "$SOURCE_CONFIG_CONTAINER" \
    --target-config-file "$TARGET_CONFIG_CONTAINER"
}
