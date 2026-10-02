#!/bin/sh
# Builds .env for CI from .env.example (single source of config keys, 12-factor X)
# and overrides only the values that differ in CI. Real secrets, if ever needed,
# should come from GitHub Secrets via the step's env, not from this file.
set -eu

cp .env.example .env

while IFS='=' read -r key value; do
    [ -z "$key" ] && continue
    if grep -q "^${key}=" .env; then
        sed -i "s|^${key}=.*|${key}=${value}|" .env
    else
        printf '%s=%s\n' "$key" "$value" >> .env
    fi
done <<'OVERRIDES'
APP_ENV=test
POSTGRES_USER=postgres
POSTGRES_PASSWORD=postgres
DB_USER=postgres
DB_PASSWORD=postgres
REDIS_PASSWORD=redis_password
JWT_ACCESS_SECRET=test_access_secret
JWT_REFRESH_SECRET=test_refresh_secret
SMTP_HOST=localhost
SMTP_PORT=1025
SMTP_USER=test
SMTP_PASS=test
OVERRIDES

echo "CI .env generated from .env.example"
