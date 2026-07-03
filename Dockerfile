# syntax=docker/dockerfile:1
# Multi-stage: base (deps) → dev (toolbox for containerized dev/test/lint)
#                          → build → cli (distroless wowapi CLI image)

FROM golang:1.26-alpine AS base
RUN apk add --no-cache git make bash grep
WORKDIR /src
COPY go.mod ./
# go.sum appears once external deps land (Phase 1+); wildcard keeps this stage valid now.
COPY go.su[m] ./
RUN go mod download
COPY . .

# Toolbox used by deployments/compose.yaml `tools` service:
#   docker compose -f deployments/compose.yaml run --rm tools make ci
FROM base AS dev
CMD ["sleep", "infinity"]

FROM base AS build
ARG VERSION=devel
RUN CGO_ENABLED=0 go build -trimpath \
    -ldflags "-s -w -X github.com/qatoolist/wowapi/internal/buildinfo.version=${VERSION}" \
    -o /out/wowapi ./cmd/wowapi

FROM gcr.io/distroless/static-debian12:nonroot AS cli
COPY --from=build /out/wowapi /usr/local/bin/wowapi
ENTRYPOINT ["/usr/local/bin/wowapi"]
CMD ["help"]
