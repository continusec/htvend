# Tool image for htvend: bundles the htvend binary, the build-img-with-proxy
# wrapper, and an upstream buildah (>= 1.44, which carries the --mount patches
# that used to require our fork) on a small wolfi base.
#
# This is the image the Bazel rules (../rules) pull and run via podman. Build and
# publish it with `make image` / `make image-push`.

# ---- build htvend (cross-compiled per target arch, no QEMU needed) ----
FROM --platform=$BUILDPLATFORM docker.io/library/golang:1.25-trixie AS builder

ARG TARGETOS
ARG TARGETARCH

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=$TARGETOS GOARCH=$TARGETARCH \
    go build -trimpath -ldflags=-buildid= -o /out/htvend ./cmd/htvend

# ---- final image ----
FROM cgr.dev/chainguard/wolfi-base

RUN apk add --no-cache bash buildah make

COPY --from=builder /out/htvend /usr/local/bin/htvend
COPY scripts/build-img-with-proxy /usr/local/bin/build-img-with-proxy

# chroot isolation avoids needing extra capabilities when run rootless in a container
ENV BUILDAH_ISOLATION=chroot

WORKDIR /workspace
ENTRYPOINT ["/usr/local/bin/htvend"]

LABEL org.opencontainers.image.source=https://github.com/continusec/htvend
