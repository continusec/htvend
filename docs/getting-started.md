# Getting started on a clean Ubuntu 24.04

This is a verified, from-scratch runbook for building the [`examples/`](../examples)
with Bazel on a fresh Ubuntu 24.04 host (it was validated end-to-end on an isolated,
no-host-access VM). It uses the fully **local** flow — no S3 and no credentials.

The Bazel rules run everything through **podman** using the published htvend tool
image, so the host only needs podman, qemu (for multi-arch), git and Bazel. You do
**not** need to install buildah, Go, libgpgme/libseccomp, netavark/runc, or configure
subuid/subgid or apparmor sysctls — modern Ubuntu + podman handle it.

## 1. Install the dependencies

```bash
sudo apt-get update
sudo apt-get install -y git podman qemu-user-static

# Bazel, via bazelisk (Ubuntu has no bazel package). Use the right arch suffix:
#   arm64 -> bazelisk-linux-arm64, amd64 -> bazelisk-linux-amd64
sudo curl -fsSL https://github.com/bazelbuild/bazelisk/releases/latest/download/bazelisk-linux-arm64 \
  -o /usr/local/bin/bazel
sudo chmod +x /usr/local/bin/bazel
```

`qemu-user-static` automatically registers the binfmt handlers that let buildah build
foreign-architecture images, so multi-arch "just works".

> **Why so little?** podman rootless works out of the box on Ubuntu 24.04: `/etc/subuid`
> and `/etc/subgid` are pre-populated, and even with
> `kernel.apparmor_restrict_unprivileged_userns=1` the packaged podman apparmor profile
> is allowed to create user namespaces. (The heavier setup in
> [oci-image-building.md](./oci-image-building.md) is only needed to run `buildah`
> directly on the host, not for this podman-based Bazel flow.)

## 2. Get the source

```bash
git clone https://github.com/continusec/htvend.git
cd htvend/examples
```

## 3. Lock — capture assets (online, once)

```bash
bazel run //alpine-img:lock
```

This builds `alpine-img` online inside the tool image, recording every fetched asset
into `alpine-img/assets.json` and storing the content-addressed blobs in your local
htvend cache (`${XDG_DATA_HOME:-$HOME/.local/share}/htvend/cache/blobs`). No S3, no
credentials. Commit the updated `assets.json`.

For multi-architecture images, run the lock once per architecture — the assets
accumulate into the single `assets.json`.

## 4. Build — replay offline

```bash
bazel build //alpine-img:image
```

This builds the image **offline** from the checked-in `assets.json` plus the blobs
(read via the directory backend, `@alpine_img_blobs`). The result is a multi-arch OCI
layout:

```
bazel-bin/alpine-img/image.oci
```

## 5. Verify

```bash
img=$(podman pull oci:bazel-bin/alpine-img/image.oci | tail -1)
podman run --rm "$img" curl --version
# curl 8.14.1 (aarch64-alpine-linux-musl) ...
```

## Notes

- The other examples (`python3-img`, `ubuntu-img`, `java-spring-img`) use the **S3**
  blob backend instead of the local directory — see [bazel.md](./bazel.md) for the
  tweag credential-helper setup that flow needs.
- The tool image is pinned by digest in
  [`../rules/image.bzl`](../rules/image.bzl) (`DEFAULT_HTVEND_IMAGE`); podman pulls it
  on first use.
