#!/bin/sh
# Writes runtime API addresses for the SPA (12-factor III: config in the environment).
# Runs from the nginx image entrypoint before nginx starts.
set -eu

target=/usr/share/nginx/html/config.js

for name in AUTH_URL REGISTRATION_URL BRANCHES_URL BOOKING_URL; do
    eval "value=\${$name:-}"
    if [ -z "$value" ]; then
        echo "40-app-config.sh: $name is required" >&2
        exit 1
    fi
    case "$value" in
        *\"* | *\\*)
            echo "40-app-config.sh: $name must not contain quotes or backslashes" >&2
            exit 1
            ;;
    esac
done

cat > "$target" <<CONFIG
window.__APP_CONFIG__ = {
  AUTH_URL: "$AUTH_URL",
  REGISTRATION_URL: "$REGISTRATION_URL",
  BRANCHES_URL: "$BRANCHES_URL",
  BOOKING_URL: "$BOOKING_URL"
};
CONFIG

echo "40-app-config.sh: wrote $target"
