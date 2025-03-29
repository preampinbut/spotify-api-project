FROM node:20.10.0-slim AS base
ENV PNPM_HOME="/pnpm"
ENV PATH="$PNPM_HOME:$PATH"
RUN corepack enable
WORKDIR /app

FROM base AS dependencies
COPY package.json pnpm-lock.yaml ./
RUN pnpm install --frozen-lockfile

FROM base AS builder
ARG VITE_BACKEND_ENDPOINT
ENV VITE_BACKEND_ENDPOINT=$VITE_BACKEND_ENDPOINT
COPY . .
COPY --from=dependencies /app/node_modules ./node_modules
RUN pnpm build

FROM nginx:alpine
WORKDIR /usr/share/nginx/html
COPY ./nginx.conf /etc/nginx/conf.d/default.conf
RUN rm -rf ./*
COPY --from=builder /app/dist .
CMD [ "nginx", "-g", "daemon off;" ]
