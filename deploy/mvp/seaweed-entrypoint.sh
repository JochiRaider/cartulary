#!/bin/sh
set -eu
umask 077

# The upstream entrypoint drops to seaweed. Copy only this service's identity.
mkdir -p /run/cartulary-seaweed
cp /etc/cartulary-seaweed/server.crt /run/cartulary-seaweed/server.crt
cp /etc/cartulary-seaweed/server.key /run/cartulary-seaweed/server.key
chown -R seaweed:seaweed /run/cartulary-seaweed
chmod 0700 /run/cartulary-seaweed
chmod 0600 /run/cartulary-seaweed/server.*
exec /entrypoint.sh "$@"
