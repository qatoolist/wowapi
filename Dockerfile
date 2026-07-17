# syntax=docker/dockerfile:1
# Multi-stage: base (deps) → dev (toolbox for containerized dev/test/lint)
#                          → build → cli (distroless wowapi CLI image)

FROM golang:1.26-alpine@sha256:0178a641fbb4858c5f1b48e34bdaabe0350a330a1b1149aabd498d0699ff5fb2 AS base
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
# `go test -race` needs cgo, and Allure Report 2 needs Java plus the
# `allure-commandline` npm distribution. golurectl converts `go test -json`
# into Allure result files; its version is pinned by go.mod's tool directive.
RUN apk add --no-cache gcc musl-dev nodejs npm openjdk21-jre-headless python3 \
    && mkdir -p /opt/allure
COPY tools/allure/package.json tools/allure/package-lock.json /opt/allure/
RUN npm ci --omit=dev --prefix /opt/allure \
    && ln -s /opt/allure/node_modules/.bin/allure /usr/local/bin/allure \
    && go build -o /usr/local/bin/golurectl github.com/robotomize/go-allure/cmd/golurectl \
    && allure --version \
    && golurectl version
# The repo is bind-mounted at /src owned by the host/runner uid, while this
# container runs as root — git would refuse it as "dubious ownership", making
# `go` VCS stamping fail ("error obtaining VCS status: exit status 128") and
# breaking lint-boundaries/build in the containerized gate. Trust the mount.
RUN git config --global --add safe.directory '*'
ENV CGO_ENABLED=1
CMD ["sleep", "infinity"]

FROM base AS build
ARG VERSION=devel
RUN CGO_ENABLED=0 go build -trimpath \
    -ldflags "-s -w -X github.com/qatoolist/wowapi/internal/buildinfo.version=${VERSION}" \
    -o /out/wowapi ./cmd/wowapi

FROM gcr.io/distroless/static-debian12:nonroot@sha256:b7bb25d9f7c31d2bdd1982feb4dafcaf137703c7075dbe2febb41c24212b946f AS cli
COPY --from=build /out/wowapi /usr/local/bin/wowapi
ENTRYPOINT ["/usr/local/bin/wowapi"]
CMD ["help"]
