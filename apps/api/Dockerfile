FROM golang:1.26-alpine AS build

ARG TARGETOS=linux
ARG TARGETARCH

WORKDIR /repo/apps/api

COPY apps/api/go.mod apps/api/go.sum ./
RUN go mod download

COPY apps/api ./

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH:-amd64} \
    go build -trimpath -ldflags="-s -w" -o bin/api .


FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /repo/apps/api/bin/api /api

# No STORAGE_DIR and no volume: the API keeps no files. Avatars are Porte URLs the browser
# fetches directly, which is what removed the storage directory that used to 404 them.

EXPOSE 4000

USER nonroot:nonroot

ENTRYPOINT ["/api"]
