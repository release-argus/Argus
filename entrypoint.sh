#!/bin/sh
set -e

ARGUS_UID=${ARGUS_UID:-911}
ARGUS_GID=${ARGUS_GID:-911}

if [ "$(id -u)" -eq 0 ]; then
  echo "Switching user UID:GID to ${ARGUS_UID}:${ARGUS_GID}"
  sed -i -e "s/^\(argus:[^:]\):[0-9]*:[0-9]*:/\1:${ARGUS_UID}:${ARGUS_GID}:/" /etc/passwd
  sed -i -e "s/^argus:x:[0-9]\+:argus$/argus:x:${ARGUS_GID}:argus:/" /etc/group

  if [ -f "/etc/argus/config.yml" ]; then
    echo "---------------------------------------------------------------------------"
    echo "Warning - creating symlink at /app/config.yml to your /etc/argus/config.yml"
    echo "please move your mount to /app/config.yml (everything should continue"
    echo "working for now, but in future versions this auto-symlink may be removed)"
    echo "---------------------------------------------------------------------------"
    touch /app/config.yml
    rm /app/config.yml >/dev/null
    ln -s /etc/argus/config.yml /app/config.yml
    chown argus:argus /etc/argus/config.yml
  fi

  echo "Applying perms to config.yml and argus.db"
  touch /app/data/argus.db
  chown "${ARGUS_UID}:${ARGUS_GID}" \
    /app \
    /app/data \
    /app/data/argus.db \
    /app/config.yml || \
      echo "WARNING: Changing the ownership of the config/database failed, so some features may not work"

  echo "Starting..."
  su-exec argus /usr/bin/argus "$@"
fi

echo "Starting..."
exec argus "$@"
