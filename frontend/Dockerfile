FROM node:22-alpine@sha256:0a7108bf6c7bf5de370ffb1a3ed6be93d405b43ff159f681a8d18c0e2bc2e402 AS builder

WORKDIR /app

COPY package.json package-lock.json ./
RUN npm ci

COPY . .
RUN npm run build

FROM nginxinc/nginx-unprivileged:alpine@sha256:26b0bf6fbf07297983cb341998d79c831508787de26627dd2a112321b9c3a4af

COPY docker/nginx.conf /etc/nginx/conf.d/default.conf
COPY --chmod=755 docker/40-app-config.sh /docker-entrypoint.d/40-app-config.sh
# Owned by the nginx user (101) so the entrypoint can rewrite config.js without root.
COPY --from=builder --chown=101:101 /app/dist /usr/share/nginx/html

EXPOSE 8080
