#!/bin/sh
set -eu

# Bootstrap is local peer-authenticated and never uses an application credential.
exec psql --no-psqlrc --no-password --host /var/run/postgresql \
  --username postgres --dbname "${PGDATABASE:-${POSTGRES_DB:-cartulary}}" \
  --set ON_ERROR_STOP=on --file /etc/cartulary-postgres/provision.sql
