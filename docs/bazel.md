# Building OCI images in Bazel with `rules_htvend`

`rules_htvend` lets other Bazel projects build `Dockerfile`/OCI images **hermetically
and offline**: every upstream asset (base image layers, `apt`/`pip`/`apk` packages,
downloaded files) is captured once into a checked-in lockfile and a content-addressed
blob store, then replayed on every subsequent build with no network access.

It solves the "build a Dockerfile under Bazel" problem without a Docker daemon: the
rules run the published **htvend tool image** (which bundles `htvend`, `buildah`, and
the `build-img-with-proxy` wrapper) via **podman**.

## How it fits together

- **`htvend_lock`** — a `bazel run` target. Online, one-off / on-demand. Builds the
  image while recording every fetched asset into a lockfile (`assets.json` by
  default), pushes the blobs to S3, and writes the updated lockfile back into your
  source tree to be checked in.
- **`htvend_blobs_repository`** (S3) / **`htvend_blobs_dir_repository`** (local
  directory) — repository rules that make the lockfile's blobs available to Bazel.
- **`htvend_image`** — a `bazel build` target. Offline. Builds the image from the
  checked-in lockfile + blobs and emits an OCI layout directory.

## Prerequisites

The rules shell out to **podman** on a Linux host (or a Linux VM such as
[Lima](https://lima-vm.io/) on macOS). The build tooling — `buildah`, `make`,
`build-img-with-proxy` — lives **inside the tool image**, so you do *not* install it
yourself. On a stock Ubuntu 24.04 the entire dependency list is:

```bash
sudo apt-get install -y git podman qemu-user-static   # + Bazel via bazelisk
```

`qemu-user-static` registers the binfmt handlers for multi-arch builds; podman rootless
otherwise works out of the box (no subuid/subgid or apparmor tweaks needed). For the
complete, verified from-scratch runbook see
[getting-started.md](./getting-started.md). The underlying mechanics (and the heavier
setup needed only to run `buildah` directly on the host) are in
[oci-image-building.md](./oci-image-building.md).

## Wiring it into a consumer project

### `MODULE.bazel`

```python
bazel_dep(name = "rules_htvend", version = "0.0.0")

# Until rules_htvend is published to the Bazel Central Registry, point at the repo:
git_override(
    module_name = "rules_htvend",
    remote = "https://github.com/continusec/htvend.git",
    commit = "<pin a commit>",
    # the module lives in the rules/ sub-directory of the repo
    strip_prefix = "rules",
)

# Blob backends. Pick S3 or directory per image (see below).
htvend_blobs_repository = use_repo_rule("@rules_htvend//:blobs_repository.bzl", "htvend_blobs_repository")
htvend_blobs_repository(
    name = "my_app_blobs",
    assets_json = "//path/to/my-app:assets.json",
    s3_bucket = "your-bucket",
    s3_prefix = "blobs/",
)
```

### `BUILD.bazel` (next to your `Dockerfile`)

```python
load("@rules_htvend//:image.bzl", "htvend_image")
load("@rules_htvend//:lock.bzl", "htvend_lock")

htvend_lock(
    name = "lock",
    s3_bucket = "your-bucket",
    s3_prefix = "blobs/",
)

htvend_image(
    name = "image",
    blobs = "@my_app_blobs//:blobs",
)
```

Each package also needs a tiny `Makefile` and a `Dockerfile`. The rules run
`make -B` inside the tool image; the default `Makefile` just invokes the wrapper:

```makefile
assets.json:
	build-img-with-proxy
```

### Day-to-day

```bash
# (re)generate the lockfile and populate the blob store — online, on demand
bazel run //path/to/my-app:lock

# build the image — offline, hermetic, cached by Bazel
bazel build //path/to/my-app:image
```

`bazel build :image` produces an OCI layout directory under `bazel-bin/...` that you
can feed into other rules (e.g. an `oci_push` from rules_oci/rules_img) or load with
`podman pull oci:bazel-bin/.../image.oci`.

## Blob backends

Both backends expose the same `@<name>//:blobs` target, so they're interchangeable in
the `blobs` attribute of `htvend_image`.

### S3 (`htvend_blobs_repository`)

Blobs are downloaded by sha256 from `https://<bucket>.s3.amazonaws.com/<prefix><sha>`
and hash-verified. Auth is handled by the
[tweag credential helper](https://github.com/tweag/credential-helper), which reads the
standard AWS chain — wire it up once in the consumer repo:

```
# .bazelrc
common --credential_helper=s3.amazonaws.com=%workspace%/tools/credential-helper
common --credential_helper=*.s3.amazonaws.com=%workspace%/tools/credential-helper
```

```json
// .tweag-credential-helper.json
{
  "urls": [
    { "scheme": "https", "host": "*.s3.amazonaws.com", "helper": "s3",
      "config": { "region": "us-east-2" } }
  ]
}
```

```bash
# install once per machine
bazel run @tweag-credential-helper//installer
```

`ctx.download` skips blobs already present with the matching sha256, so repeated builds
don't re-fetch. See the [`examples/`](../examples) root for a complete working setup.

### Local directory (`htvend_blobs_dir_repository`)

For setups without S3 — blobs on a local disk, an NFS mount, or just the local htvend
cache. No credentials, no network.

```python
htvend_blobs_dir_repository = use_repo_rule("@rules_htvend//:blobs_dir_repository.bzl", "htvend_blobs_dir_repository")

htvend_blobs_dir_repository(
    name = "my_app_blobs",
    # optional; defaults to $HTVEND_BLOBS_DIR, then
    # ${XDG_DATA_HOME:-$HOME/.local/share}/htvend/cache/blobs
    blobs_dir = "/srv/shared/htvend-blobs",
)
```

Pair it with an `htvend_lock` that has **no** `s3_bucket` — the lock then just stores
blobs into the local directory (defaulting to the htvend cache) with no S3 export and
no credentials:

```python
htvend_lock(name = "lock")          # writes blobs to the local cache only
htvend_image(name = "image", blobs = "@my_app_blobs//:blobs")
```

`examples/alpine-img` uses exactly this fully-local backend; see
[getting-started.md](./getting-started.md) for the end-to-end commands.

## Multiple targets in one directory

Each rule (and the `blobs_repository` `assets_json` attribute) takes a
`lockfile_name` (default `assets.json`). To host several images in one package, give
each its own lockfile:

```python
htvend_lock(name = "lock_a", lockfile_name = "a.assets.json", s3_bucket = "...", s3_prefix = "blobs/")
htvend_image(name = "image_a", lockfile_name = "a.assets.json", blobs = "@a_blobs//:blobs")
```

## Multi-architecture images

A single `assets.json` can hold the assets for several architectures — different
architectures simply reference different URLs, and a build ignores any assets it
doesn't need. Run the lock step once per architecture to accumulate them all into the
one shared lockfile:

```bash
bazel run //path/to/my-app:lock   # on / for linux/amd64
bazel run //path/to/my-app:lock   # on / for linux/arm64
```

then commit the combined `assets.json`. `build-img-with-proxy` builds a multi-arch
manifest, so a single `bazel build :image` produces the multi-arch OCI layout.

## Pinning the tool image

`htvend_image` / `htvend_lock` default to `ghcr.io/continusec/htvend:1.6` (see
`DEFAULT_HTVEND_IMAGE` in [`../rules/image.bzl`](../rules/image.bzl)). podman uses a
matching local image if present (e.g. after `cd cli && make image`), otherwise pulls
it. Override per-target with `image = "ghcr.io/continusec/htvend@sha256:..."`, and
pin by digest for fully reproducible builds.
