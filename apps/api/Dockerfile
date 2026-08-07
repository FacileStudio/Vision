FROM golang:1.25-alpine AS build

ARG TARGETOS=linux
ARG TARGETARCH

WORKDIR /repo/apps/api

COPY apps/api/go.mod apps/api/go.sum ./
RUN go mod download

COPY apps/api ./

RUN CGO_ENABLED=0 GOOS=${TARGETOS} GOARCH=${TARGETARCH:-amd64} \
    go build -trimpath -ldflags="-s -w" -o bin/api .

RUN mkdir -p /seed/data

FROM gcr.io/distroless/static-debian12:nonroot

COPY --from=build /repo/apps/api/bin/api /api

# STORAGE_DIR defaults to a relative ./data, and the :nonroot base sets
# WorkingDir=/home/nonroot — so the default silently moves with the base image.
# Pin it somewhere nonroot can actually create it, the way Capsule pins CLIENT_DIR.
ENV STORAGE_DIR=/home/nonroot/data

# The directory has to exist in the image, owned by nonroot, before compose mounts a volume
# over it: Docker seeds an empty named volume from whatever is at the mount path, ownership
# included, and creates it root-owned 0755 when the path is missing. Distroless has no shell
# to chown at runtime, so uid 65532 would find an unwritable directory and the process would
# exit 1 on its MkdirAll. COPY --chown is the only way to set this here.
COPY --from=build --chown=65532:65532 /seed/data /home/nonroot/data

EXPOSE 4000

USER nonroot:nonroot

ENTRYPOINT ["/api"]
