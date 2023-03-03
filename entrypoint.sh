#!/bin/sh

ARGUS_UID=${ARGUS_UID:-911}
ARGUS_GID=${ARGUS_GID:-911}

if [ "$(id -u)" -eq 0 ]; then
  echo "Switching user UID:GID to ${ARGUS_UID}:${ARGUS_GID}"
  groupmod -o -g "$ARGUS_GID" argus >/dev/null
  usermod -o -u "$ARGUS_UID" argus >/dev/null

  if [ -f "/etc/argus/config.yml" ]; then
    echo "---------------------------------------------------------------------"
    echo "Warning - symlinking /app/config.yml to your /etc/argus/config.yml"
    echo "please move your mount to /app/config.yml (everything should continue"
    echo "working for now, but in future versions I may remove this symlinking)"
    echo "---------------------------------------------------------------------"
    touch /app/config.yml
    rm /app/config.yml >/dev/null
    ln -s /etc/argus/config.yml /app/config.yml
    chown argus:argus /etc/argus/config.yml
  fi

  echo "Applying perms to config.yml and argus.db"
  touch /app/data/argus.db
  chown argus:argus /app/data/argus.db

  su - argus
fi

echo "Starting..."
argus "$@"
