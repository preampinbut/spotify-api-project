FROM oven/bun:1-alpine AS base
WORKDIR /app

FROM base AS prune
RUN bun install --global turbo@2.6.2
COPY --exclude=apps/backend/ . .
RUN turbo prune frontend --docker

FROM base AS dependencies
ARG VITE_BACKEND_ENDPOINT
ENV VITE_BACKEND_ENDPOINT=$VITE_BACKEND_ENDPOINT
COPY --from=prune /app/out/json/ .
RUN bun install

FROM base AS builder
COPY --from=dependencies /app/ .
COPY --from=prune /app/out/full/ .
RUN bun turbo build

FROM caddy:2.10.2-alpine
WORKDIR /var/www

COPY ./apps/frontend/Caddyfile /etc/caddy/Caddyfile
COPY --from=builder /app/apps/frontend/dist .
