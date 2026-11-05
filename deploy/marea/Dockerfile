FROM oven/bun:1.3.11 AS frontend
WORKDIR /repo
COPY . .
RUN bun install --frozen-lockfile && bun run build -- frontend

FROM golang:1.26.10-alpine AS backend
RUN apk add --no-cache gcc musl-dev sqlite-dev
WORKDIR /repo
COPY apps/server/go.mod apps/server/go.sum ./apps/server/
RUN cd apps/server && go mod download
COPY apps/server/ ./apps/server/
COPY --from=frontend /repo/apps/server/cmd/openpost/public ./apps/server/cmd/openpost/public
RUN cd apps/server && CGO_ENABLED=1 GOOS=linux go build -buildvcs=false -ldflags="-s -w" -o openpost ./cmd/openpost

FROM alpine:3.24.3 AS runtime
RUN apk add --no-cache ca-certificates ffmpeg tzdata sqlite \
    'libcrypto3>=3.5.8-r0' 'libssl3>=3.5.8-r0' \
    && adduser -D -u 1000 marea
WORKDIR /app
COPY --from=backend /repo/apps/server/openpost ./openpost
RUN mkdir -p /data/media /data/db && chown -R marea:marea /data
USER marea
ENV OPENPOST_PORT=8080 \
    OPENPOST_EDITION=selfhost \
    OPENPOST_DATABASE_DRIVER=sqlite \
    OPENPOST_DATABASE_PATH=/data/db/marea.db \
    OPENPOST_STORAGE_DRIVER=local \
    OPENPOST_MEDIA_PATH=/data/media \
    OPENPOST_DIAGNOSTICS_STATE_FILE=/data/diagnostics-installation-id
EXPOSE 8080
HEALTHCHECK --interval=30s --timeout=3s --start-period=10s --retries=3 \
    CMD wget --no-verbose --tries=1 --spider http://localhost:8080/api/v1/health || exit 1
CMD ["./openpost", "all"]
