# first, build the binaries we need
FROM library/golang:1.25-trixie AS builder

# add dependencies required by buildah for building
RUN apt-get update -y && \
    apt-get install -y --no-install-recommends \
        libgpgme-dev \
        libseccomp-dev && \
    rm -rf /var/lib/apt/lists/* && \
    mkdir /go/src/buildah /go/src/htvend

# pull in our buildah branch (until the PRs are merged)
ADD https://api.github.com/repos/aeijdenberg/buildah/tarball/continusecbuild /buildah.tar.gz

# pull in htvend source
ADD . /go/src/htvend

# untar and build
RUN tar -C /go/src/buildah --strip-components=1 -zxf /buildah.tar.gz && \
    make -C /go/src/buildah GIT_COMMIT=continusecbuild bin/buildah && \
    make -C /go/src/htvend all && \
    mkdir /result && \
    mv /go/src/buildah/bin/buildah /result/patched-buildah && \
    mv \
        /go/src/htvend/target/build-img-with-proxy \
        /go/src/htvend/target/htvend \
        /result/

# now copy into final image
FROM library/debian:trixie-slim

# install some base packages we need and normal podman/buildah configs
RUN apt-get update -y && \
    DEBIAN_FRONTEND=noninteractive apt-get install -y --no-install-recommends \
        ca-certificates \
        fuse-overlayfs \
        libgpgme11 \
        netavark \
        runc && \
    rm -rf /var/lib/apt/lists/* && \
    usermod --add-subuids 1-65535 --add-subgids 1-65535 root && \
    mkdir /etc/containers && \
    echo 'unqualified-search-registries = ["docker.io"]' > /etc/containers/registries.conf && \
    echo '{"default":[{"type":"insecureAcceptAnything"}]}' > /etc/containers/policy.json

# then copy in our binaries
COPY --from=builder /result/* /usr/local/bin/

ENTRYPOINT ["/usr/local/bin/htvend"]
