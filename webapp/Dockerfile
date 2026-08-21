# Static export: the build stage produces files and the runtime stage is a file
# server. No Node runs in production — which on a 2 GB VPS already hosting the
# bot, Postgres and two other projects is the difference that matters.
FROM node:24-alpine AS builder
WORKDIR /app
COPY package.json package-lock.json ./
RUN npm ci
COPY . .
# Baked at build time: a static export has no runtime environment. Both values
# are public — a path prefix and a t.me link — so baking them costs nothing.
ARG NEXT_PUBLIC_API_BASE=/api
ARG NEXT_PUBLIC_TG_LINK
ENV NEXT_PUBLIC_API_BASE=$NEXT_PUBLIC_API_BASE \
    NEXT_PUBLIC_TG_LINK=$NEXT_PUBLIC_TG_LINK
RUN npm run build

FROM caddy:2-alpine
COPY Caddyfile /etc/caddy/Caddyfile
COPY --from=builder /app/out /srv
