# Building OCI images in Bazel with `rules_htvend`

`rules_htvend` lets other Bazel projects build `Dockerfile`/OCI images **hermetically
and offline**: every upstream asset (base image layers, `apt`/`pip`/`apk` packages,
downloaded files) is captured once into a checked-in lockfile and a content-addressed
blob store, then replayed on every subsequent build with no network access.

It solves the "build a Dockerfile under Bazel" problem without a Docker daemon: the
rules run the published **htvend tool image** (which bundles `htvend`, `buildah`, and
the `build-img-with-proxy` wrapper) via **podman**.

## How it fits together

One macro, **`htvend_image`**, creates the pair of targets every image needs:

- **`:<name>.lock`** — a `bazel run` target. Online, on demand. Builds the image while
  recording every fetched asset into a lockfile (`assets.json` by default), stores the
  blobs (locally, and optionally to S3), and writes the updated lockfile back into your
  source tree to be checked in.
- **`:<name>`** — a `bazel build` target. Offline. Builds the image from the checked-in
  lockfile + blobs and emits an OCI layout directory.

Supporting them:

- **`htvend_blobs_repository`** (S3) / **`htvend_blobs_dir_repository`** (local
  directory) — repository rules that make the lockfile's blobs available to Bazel.

## Prerequisites

The rules shell out to **podman** on a Linux host (or a Linux VM such as
[Lima](https://lima-vm.io/) on macOS). The build tooling — `buildah` and
`build-img-with-proxy` — lives **inside the tool image**, so you do *not* install them
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

A single `htvend_image` call creates the pair of targets every image needs:

```python
load("@rules_htvend//:defs.bzl", "htvend_image")

htvend_image(
    name = "image",
    blobs = "@my_app_blobs//:blobs",
)
```

`:image.lock` will export blobs to `your-bucket`/`blobs/` automatically — it reads
those from `@my_app_blobs`'s own `:blobs_info` (generated alongside `:blobs`), so the
S3 location is only specified once, on `htvend_blobs_repository` above. For the
local-only (directory) flow, just point `blobs` at an `htvend_blobs_dir_repository`
instead — see "Blob backends" below.

This produces:

- `//path/to/my-app:image` — `bazel build`, the OCI image built offline;
- `//path/to/my-app:image.lock` — `bazel run`, regenerates the lockfile + blobs.

Pass build-time configuration through attributes:

- `dockerfile` — which Dockerfile to build (default `"Dockerfile"`); set it when a
  package has more than one.
- `env` — environment variables to set during the build, e.g. to tell
  `build-img-with-proxy` where to mount the CA truststore or a maven settings file:

  ```python
  htvend_image(
      name = "image",
      blobs = "@my_app_blobs//:blobs",
      dockerfile = "Dockerfile.app",
      env = {
          "JKS_KEYSTORE_PATH": "/etc/ssl/certs/java/cacerts",
          "MVN_SETTINGS_PATH": "/root/.m2/settings.xml",
      },
  )
  ```

  Most builds need none of these. They're for specialty cases (e.g. Maven/Java
  needing a truststore or `settings.xml` inside the build container) — see the
  header comments in
  [`cli/scripts/build-img-with-proxy`](../cli/scripts/build-img-with-proxy) for the
  full list of variables it reads.

### Day-to-day

```bash
# (re)generate the lockfile and populate the blob store — online, on demand
bazel run //path/to/my-app:image.lock

# build the image — offline, hermetic, cached by Bazel
bazel build //path/to/my-app:image
```

`:image.lock` only needs to be re-run when an *external* dependency changes —
e.g. a new/updated base image, a version bump in a `pom.xml` or `requirements.txt`,
or a new package added to the `Dockerfile`. Changes to your application's own
source (that don't pull in new upstream assets) don't need a re-lock; `:image`
rebuilds from the existing lockfile + blobs as usual.

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
    assets_json = "//path/to/my-app:assets.json",
    # optional; defaults to $HTVEND_BLOBS_DIR, then
    # ${XDG_DATA_HOME:-$HOME/.local/share}/htvend/cache/blobs
    blobs_dir = "/srv/shared/htvend-blobs",
)
```

The directory may hold blobs for many images (it's content-addressed by sha256);
`assets_json` limits what this repository exposes to just the blobs `:my-app:image`
needs, so each image's `@..._blobs//:blobs` only contains its own dependencies.

A directory-backed `:blobs_info` reports no S3 bucket, so `:image.lock` just stores
blobs into the local directory (defaulting to the htvend cache) with no S3 export and
no credentials — no extra attributes needed:

```python
htvend_image(name = "image", blobs = "@my_app_blobs//:blobs")
```

`examples/alpine-img` uses exactly this fully-local backend; see
[getting-started.md](./getting-started.md) for the end-to-end commands.

## Multiple targets in one directory

`htvend_image` takes a `lockfile_name` (default `assets.json`). To host several images
in one package, give each its own name and lockfile (and point each at the right
`Dockerfile`):

```python
htvend_image(
    name = "a",
    blobs = "@a_blobs//:blobs",
    dockerfile = "Dockerfile.a",
    lockfile_name = "a.assets.json",
)
# -> //pkg:a and //pkg:a.lock
```

The matching `htvend_blobs_repository` should set `assets_json` to the same
`lockfile_name`.

## Multi-architecture images

`build-img-with-proxy` builds a multi-arch manifest in one shot, looping over a list
of `os/arch` platforms. Control that list with the `platforms` attribute (default
`["linux/amd64", "linux/arm64"]`):

```python
htvend_image(
    name = "image",
    blobs = "@my_app_blobs//:blobs",
    platforms = ["linux/amd64", "linux/arm64", "linux/arm/v7"],
)
```

> The host needs the matching `binfmt` handlers to build foreign architectures —
> `apt-get install qemu-user-static` registers them (see
> [getting-started.md](./getting-started.md)).

A single `assets.json` can hold the assets for every architecture — different
architectures simply reference different URLs, and a build ignores any it doesn't
need. The lock captures all the platforms in `platforms` in one run; commit the
combined `assets.json`, and `bazel build :image` then produces the multi-arch OCI
layout offline.

> **Why an attribute and not Bazel `--platforms`?** The multi-arch fan-out happens
> *inside* buildah (one action emitting a manifest list), not via Bazel per-platform
> transitions, so the natural knob is this list of buildah `os/arch` strings rather
> than `@platforms//` constraint values.

## Remote execution (RBE)

By default `bazel build :image` runs the build under **podman** on the local machine,
with `--network=none` (the container gets no external network — buildah's inner
`--network=host` containers share that same netns, so loopback to the htvend proxy
still works). That path is deliberately **local-only**: the rule tags the action
`local` + `no-sandbox` because podman/buildah need real host devices (`/dev/fuse`),
their own user+mount namespaces, and `$HOME` container storage — none of which survive
Bazel's sandbox or a remote worker.

For RBE, switch the build to **direct mode**:

```bash
bazel build //path/to/my-app:image --@rules_htvend//:exec_mode=direct
```

In direct mode the rule does **not** shell out to podman. It runs `htvend`, `buildah`,
and `build-img-with-proxy` straight from `PATH`, with the build context and blobs as
declared Bazel inputs and no network access — so the action is both sandboxable and
remote-eligible (the `local`/`no-sandbox` tags are dropped). You can also set it
per-target with `exec_mode = "direct"` on `htvend_image`.

### Recommended: a `--config=rbe` bazelrc convention

`exec_mode` only controls *how the action runs*; it doesn't configure RBE itself
(`--remote_executor`, `--remote_default_exec_properties`, etc. are still needed). Bundle
both into a `build:rbe` config group in your `.bazelrc` so the two switches stay in sync —
users without RBE configured get plain `bazel build` (podman, local); users with RBE
configured add `--config=rbe` and get direct mode pointed at the cluster:

```ini
# .bazelrc
build:rbe --@rules_htvend//:exec_mode=direct
build:rbe --remote_executor=grpc://rbe.example.com:443
build:rbe --remote_instance_name=...
```

`htvend_image` already sets `exec_properties` (`container-image`, `OSFamily`) to match
the tool image it's built against, so you don't need to repeat those here — see
"Providing the tooling on the worker" below.

```bash
bazel build //path/to/my-app:image            # local, podman
bazel build //path/to/my-app:image --config=rbe  # remote, direct mode
```

See [`../examples/.bazelrc`](../examples/.bazelrc) for a working `build:rbe` block (pointed
at a local Buildbarn cluster, per the "Verified against a local Buildbarn cluster" section
below).

The action runs with `use_default_shell_env = True`, so the three tools must be on the
action's `PATH` (extend it if needed with `--action_env=PATH=/usr/local/bin:/usr/bin:/bin`).
In the tool image they live in `/usr/local/bin`, which a normal shell `PATH` already
covers.

### Providing the tooling on the worker

Direct mode expects `htvend`/`buildah`/`build-img-with-proxy` to already be present in
the execution environment. The simplest, most robust way to guarantee that on RBE is to
**make the worker's container image the htvend tool image itself** (it already bundles
all three).

`htvend_image` does this selection for you automatically: it sets the target's
`exec_properties` to

```python
{
    "container-image": "docker://" + image,  # image = DEFAULT_HTVEND_IMAGE unless overridden
    "OSFamily": "linux",
}
```

so the worker image is derived from the same `image`/`DEFAULT_HTVEND_IMAGE` used for
the podman path — one source of truth, no separate digest to keep in sync. This only
takes effect under `--@rules_htvend//:exec_mode=direct`; it's harmless otherwise.

The exact `exec_properties` keys are backend-specific; the defaults above match
Buildbarn's `container-image`/`OSFamily` platform properties (see "Verified against a
local Buildbarn cluster" below). If your backend uses different keys, or you want a
different worker pool than the image podman pulls, pass your own `exec_properties` to
`htvend_image` to override the default entirely:

```python
htvend_image(
    name = "image",
    blobs = "@my_app_blobs//:blobs",
    exec_mode = "direct",
    exec_properties = {
        "container-image": "docker://ghcr.io/continusec/htvend@sha256:...",
        "OSFamily": "linux",
    },
)
```

buildah still needs user-namespace + fuse-overlayfs support on the worker — that's
buildah's nature, not something Bazel can remove; ensure your worker pool provides it.

### Testing readiness without a full RBE cluster

The cheapest check needs no setup at all: the **default podman path already runs with
`--network=none`** (see above), so a plain

```bash
bazel build //path/to/my-app:image
```

is itself a no-external-network build. The real hermeticity guarantee underneath is
htvend `offline` mode, which serves strictly from the captured blobs and **fails
closed**: remove a blob from the blob set and the build errors rather than reaching the
internet.

The next rung up is a real local RBE server (e.g. Buildbarn's `bb-deployments`) with a
tool-image worker, run with `--@rules_htvend//:exec_mode=direct
--remote_executor=grpc://localhost:...`. There's no supported way to exercise
direct/RBE mode without such a worker: it expects `htvend`, `buildah` and
`build-img-with-proxy` to already be on `PATH`, and distro-packaged `buildah` (e.g.
`apt-get install buildah`) is typically too old to support the flags these rules
rely on — only the published tool image is supported.

### Verified against a local Buildbarn cluster

This has been verified end-to-end against `buildbarn/bb-deployments`' docker-compose
deployment, using its `*-hardlinking-ubuntu22-04` worker/runner pair with the runner's image
swapped for the htvend tool image (`ghcr.io/continusec/htvend`) and its `container-image`
platform property set to the exact `DEFAULT_HTVEND_IMAGE` (including digest), matching
what `htvend_image` now sets in `exec_properties` automatically:

```bash
bazel build //path/to/my-app:image \
    --@rules_htvend//:exec_mode=direct \
    --remote_executor=grpc://localhost:8980 \
    --remote_instance_name=hardlinking
```

The action dispatches to the runner ("Runner: remote"), which has `network_mode: none` — no
network interfaces at all, a stronger guarantee than `--sandbox_default_allow_network=false` —
and `htvend`/`buildah`/`build-img-with-proxy` already on `PATH` from the image. The build
produced a valid multi-arch OCI layout. Three small additions were needed on the runner
container beyond the reference compose file, all because the default Docker security profile
is tighter than a bare sandbox/host process:

- `cap_add: [SYS_ADMIN]` + `security_opt: [seccomp=unconfined, apparmor=unconfined]` — buildah
  calls `unshare(CLONE_NEWUSER)` even under `BUILDAH_ISOLATION=chroot`.
- `tmpfs: [/var/tmp:exec]` — buildah stages an overlay mount for the build context under
  `/var/tmp`; overlay-on-overlay (the container's own root is overlay2) is rejected by the
  kernel, so `/var/tmp` needs a non-overlay filesystem.
- `devices: [/dev/fuse:/dev/fuse]` + `ulimits.nofile` raised to 1048576 — the same
  `/dev/fuse` and `RLIMIT_NOFILE` needs the podman path already covers via `--device /dev/fuse`.

These are properties of "a container that runs buildah", independent of `rules_htvend` — an
RBE worker pool built from the tool image should grant the same.

## Pinning the tool image

`htvend_image` defaults to the published `ghcr.io/continusec/htvend` image, pinned by
digest (see `DEFAULT_HTVEND_IMAGE` in [`../rules/image.bzl`](../rules/image.bzl)).
podman uses a
matching local image if present (e.g. after `cd cli && make image`), otherwise pulls
it. Override per-target with `image = "ghcr.io/continusec/htvend@sha256:..."`, and
pin by digest for fully reproducible builds.
