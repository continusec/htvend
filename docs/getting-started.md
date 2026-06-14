# Getting started on a clean Ubuntu 24.04

This is a verified, from-scratch runbook for building the [`examples/`](../examples)
with Bazel on a fresh Ubuntu 24.04 host (it was validated end-to-end on an isolated,
no-host-access VM). It uses the fully **local** flow — no S3 and no credentials.

The Bazel rules run everything through **podman** using the published htvend tool
image, so the host only needs podman, git and Bazel (plus qemu for the optional
multi-arch case). You do **not** need to install buildah, Go, libgpgme/libseccomp,
netavark/runc, or configure subuid/subgid or apparmor sysctls — modern Ubuntu + podman
handle it.

## 1. Install the dependencies

```bash
sudo apt-get update
sudo apt-get install -y git podman   # add qemu-user-static for multi-arch (see step 4)

# Bazel, via bazelisk (Ubuntu has no bazel package). Use the right arch suffix:
#   arm64 -> bazelisk-linux-arm64, amd64 -> bazelisk-linux-amd64
sudo curl -fsSL https://github.com/bazelbuild/bazelisk/releases/latest/download/bazelisk-linux-arm64 \
  -o /usr/local/bin/bazel
sudo chmod +x /usr/local/bin/bazel
```

Builds default to your host's architecture, which needs no extra setup. To build
foreign architectures, `sudo apt-get install -y qemu-user-static` registers the binfmt
handlers buildah needs (see step 4).

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
bazel run //alpine-img:image.lock
```

This builds `alpine-img` online inside the tool image, recording every fetched asset
into `alpine-img/assets.json` and storing the content-addressed blobs in your local
htvend cache (`${XDG_DATA_HOME:-$HOME/.local/share}/htvend/cache/blobs`). No S3, no
credentials. The examples don't check `assets.json` in (it's generated on demand, and
the blobs it references aren't in the repo) — you'd commit it in your own project.

By default the lock captures just your host's architecture. To capture more, set the
`platforms` attribute on `htvend_image` (and install `qemu-user-static`) — the assets
for every architecture accumulate into the single `assets.json`.

## 4. Build — replay offline

```bash
bazel build //alpine-img:image
```

This builds the image **offline** from the `assets.json` you just generated plus the
blobs (read via the directory backend, `@alpine_img_blobs`). The result is an OCI layout
(host architecture by default):

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

- To share blobs across machines/CI instead of a local directory, use the **S3** blob
  backend — see the appendix in [bazel.md](./bazel.md) for the tweag credential-helper
  setup that flow needs.
- The tool image is pinned by digest in
  [`../rules/image.bzl`](../rules/image.bzl) (`DEFAULT_HTVEND_IMAGE`); podman pulls it
  on first use.
