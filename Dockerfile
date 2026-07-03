# syntax=docker/dockerfile:1

FROM --platform=${BUILDPLATFORM} golang:1.25-bookworm AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN go generate -tags tools ./...

ARG TARGETOS
ARG TARGETARCH

RUN CGO_ENABLED=0 GOOS="${TARGETOS}" GOARCH="${TARGETARCH}" \
    go build -ldflags="-w -s -buildid=" -trimpath -o /out/app main.go

# distroless/static: no shell, no package manager, no libc - just
# ca-certificates, tzdata and a nonroot user. Works because the binary
# above is built with CGO_ENABLED=0 (fully static, no glibc needed).
# curl/procps from the old debian-slim base weren't used by the app or
# any health check (all probes are httpGet, executed by kubelet from
# outside the container), so nothing here depends on them.
FROM gcr.io/distroless/static-debian12:nonroot@sha256:d093aa3e30dbadd3efe1310db061a14da60299baff8450a17fe0ccc514a16639

# numeric uid:gid (distroless's "nonroot") instead of the symbolic name -
# --chown with a name + --link fails to resolve on cross-platform/emulated
# builds ("invalid user index: -1"); numeric IDs skip that lookup entirely.
COPY --link --from=builder --chown=65532:65532 /out/app /opt/app

USER 65532:65532

ENTRYPOINT [ "/opt/app" ]
