#!/bin/sh
set -e

DATA_DIR="/data"
SALT_FILE="$DATA_DIR/salt"

if [ ! -f "$SALT_FILE" ]; then
  if [ -z "$VEILKEY_PASSWORD" ]; then
    echo "ERROR: VEILKEY_PASSWORD required for first run."
    exit 1
  fi

  echo "=== VeilKey Agent Init ==="
  echo "$VEILKEY_PASSWORD" | veilkey-localvault init --root

  echo "Init complete."
fi

exec veilkey-localvault "$@"
