FROM node:20.10.0-slim AS base
WORKDIR /app

ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
RUN corepack enable

FROM base AS dependencies

COPY package.json pnpm-lock.yaml ./

RUN pnpm install --frozen-lockfile

FROM base AS builder

ARG VITE_BACKEND_ENDPOINT
ENV VITE_BACKEND_ENDPOINT=$VITE_BACKEND_ENDPOINT

COPY . .
COPY --from=dependencies /app/node_modules ./node_modules

RUN pnpm build

FROM caddy:2.10.2-alpine
WORKDIR /var/www

COPY ./Caddyfile /etc/caddy/Caddyfile
COPY --from=builder /app/dist .
