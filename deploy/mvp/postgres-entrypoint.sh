#!/bin/sh
set -eu
umask 077

# Read-only host credentials stay untouched. PostgreSQL requires an owned key.
mkdir -p /run/cartulary-postgres
cp /etc/cartulary-postgres/tls/server.crt /run/cartulary-postgres/server.crt
cp /etc/cartulary-postgres/tls/server.key /run/cartulary-postgres/server.key
chown -R postgres:postgres /run/cartulary-postgres
chmod 0700 /run/cartulary-postgres
chmod 0600 /run/cartulary-postgres/server.*

# The upstream entrypoint requires an initialization password. This ephemeral
# value is never accepted over a network: the fixed HBA permits only certificates.
# Provisioning removes the verifier before the ordinary listener starts.
if [ ! -s "$PGDATA/PG_VERSION" ]; then
  POSTGRES_PASSWORD="$(od -An -N32 -tx1 /dev/urandom | tr -d ' \n')"
  export POSTGRES_PASSWORD
fi
# Preserve the upstream directory-creation contract: it secures PGDATA itself.
# A 077 mask would strand a root-owned parent above PGDATA.
umask 022
exec /usr/local/bin/docker-entrypoint.sh "$@"
