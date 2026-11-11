FROM oven/bun:1.3.11 AS client-build
WORKDIR /client
COPY apps/client/package.json apps/client/bun.lock ./
RUN bun install --frozen-lockfile
COPY apps/client/ .
RUN bun run build

FROM golang:1.26.11-alpine AS api-build

ARG TARGETOS=linux
ARG TARGETARCH

WORKDIR /repo/apps/api

COPY apps/api/go.mod apps/api/go.sum ./
RUN go mod download

COPY apps/api ./
COPY apps/client/package.json /repo/apps/client/package.json
COPY scripts/version.sh /repo/scripts/version.sh

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH:-amd64} \
    go build -trimpath -ldflags="-s -w -X main.version=$(sh /repo/scripts/version.sh)" -o bin/api .

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=api-build /repo/apps/api/bin/api /api
COPY --from=client-build /client/build /client

# The :nonroot base sets its own WorkingDir, so a relative ./client would resolve under
# /home/nonroot and the SPA would silently not be served. Name the directory.
ENV CLIENT_DIR=/client

EXPOSE 4000

USER nonroot:nonroot

ENTRYPOINT ["/api"]
