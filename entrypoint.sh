#!/bin/sh

ARGUS_UID=${ARGUS_UID:-911}
ARGUS_GID=${ARGUS_GID:-911}

echo "Switching user UID:GID to ${ARGUS_UID}:${ARGUS_GID}"
groupmod -o -g "$ARGUS_GID" argus >/dev/null
usermod -o -u "$ARGUS_UID" argus >/dev/null

echo "Applying perms to config.yml and argus.db"
chown argus:argus /app/config.yml
touch /app/data/argus.db
chown argus:argus /app/data/argus.db

su - argus
echo "Starting..."
argus "$@"
