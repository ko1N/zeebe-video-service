#!/bin/sh

export ZEEBE_HOST="${ZEEBE_HOST:-0.0.0.0:26500}"

export ZEEBE_PLAINTEXT="${ZEEBE_PLAINTEXT:-false}"

cat << EOT > ./config.yml
zeebe:
  host: "$ZEEBE_HOST"
  plaintext: $ZEEBE_PLAINTEXT
EOT

exec "$@"
