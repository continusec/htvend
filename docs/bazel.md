# Building OCI images in Bazel with `rules_htvend`

`rules_htvend` lets Bazel projects build `Dockerfile`/OCI images **hermetically and
offline**. The first time you build, every upstream asset (base-image layers,
`apk`/`apt`/`pip` packages, downloaded files) is captured into a checked-in lockfile
(`assets.json`) and a content-addressed blob store. Every build after that replays from
those — no network, byte-for-byte reproducible — until *you* decide to re-capture.

It does this without a Docker daemon: the rules run the published **htvend tool image**
(which bundles `htvend`, `buildah`, and the `build-img-with-proxy` wrapper) via
**podman**.

This guide is a hands-on walkthrough: you'll build a real image from an empty directory,
watch the lockfile and blob store get created, then prove the build is offline and
reproducible. Everything here is the fully-local **directory backend** — no S3, no
credentials. S3 is an [appendix](#appendix-using-s3-as-the-blob-store) for when you want
to share blobs across a team or CI.

> Every step below is run for you by [`../e2e-bazel-test.sh`](../e2e-bazel-test.sh),
> which drives the matching [`../examples/alpine-img`](../examples/alpine-img): it locks
> the image, builds it, then cleans and rebuilds to assert the image digest is unchanged.
> It's the executable version of this document.

## Prerequisites

The rules shell out to **podman** on a Linux host (or a Linux VM such as
[Lima](https://lima-vm.io/) on macOS) — plus Bazel (via bazelisk). On a stock Ubuntu
24.04 the whole dependency list is:

```bash
sudo apt-get install -y git podman   # + Bazel via bazelisk
```

podman rootless works out of the box (no subuid/subgid or apparmor tweaks). The build
tooling — `buildah`, `build-img-with-proxy` — lives **inside the tool image**, so you
don't install it. (Multi-arch builds additionally need `qemu-user-static`; the default
single-arch build does not — see [Multi-architecture images](#multi-architecture-images).)
For the verified from-scratch runbook see [getting-started.md](./getting-started.md).

## Walkthrough: your first image

We'll build this tiny project:

```
my-workspace/
├── MODULE.bazel
└── app/
    ├── BUILD.bazel
    └── Dockerfile
```

### Step 1 — `MODULE.bazel`

Depend on `rules_htvend`, then declare a **blob backend** — the repository that makes the
lockfile's blobs available to Bazel. We use the directory backend, which reads (and the
lock run writes) a plain local directory.

```python
bazel_dep(name = "rules_htvend", version = "0.0.0")

# Until rules_htvend is on the Bazel Central Registry, point at the repo. The module
# lives in the rules/ sub-directory.
git_override(
    module_name = "rules_htvend",
    remote = "https://github.com/continusec/htvend.git",
    commit = "<pin a commit>",
    strip_prefix = "rules",
)

htvend_blobs_dir_repository = use_repo_rule("@rules_htvend//:blobs_dir_repository.bzl", "htvend_blobs_dir_repository")

htvend_blobs_dir_repository(
    name = "app_blobs",
    assets_json = "//app:assets.json",
    # Where blobs live. Optional — defaults to $HTVEND_BLOBS_DIR, then the shared
    # htvend cache at ${XDG_DATA_HOME:-$HOME/.local/share}/htvend/cache/blobs.
    blobs_dir = "/srv/shared/htvend-blobs",
)
```

`assets_json` points at the lockfile this image will use. It doesn't exist yet — that's
fine, the backend treats a missing lockfile as empty so the first lock run can create it.

### Step 2 — `app/Dockerfile` and `app/BUILD.bazel`

An ordinary Dockerfile:

```dockerfile
FROM alpine:3.21

RUN apk --no-cache add curl
```

A single `htvend_image` call wires it up:

```python
load("@rules_htvend//:defs.bzl", "htvend_image")

htvend_image(
    name = "image",
    blobs = "@app_blobs//:blobs",
)
```

That one macro creates **two targets**:

| Target | Command | Network | What it does |
|--------|---------|---------|--------------|
| `//app:image.lock` | `bazel run` | **online** | Build the image while recording every fetched asset into `assets.json` and every blob into `blobs_dir`. |
| `//app:image` | `bazel build` | **offline** | Build the image from the checked-in `assets.json` + blobs. Emits an OCI layout. |

### Step 3 — lock it (online, once)

```console
$ bazel run //app:image.lock
```

This runs the Dockerfile build inside the tool image with an htvend proxy in front of it,
recording everything it pulls. You'll see the assets stream past as they're fetched:

```
--- Building linux/arm64 ---
Fetching URL: GET https://registry-1.docker.io/v2/library/alpine/manifests/3.21
Fetching URL: GET https://registry-1.docker.io/v2/library/alpine/blobs/sha256:2155344e…
Fetching URL: GET https://dl-cdn.alpinelinux.org/alpine/v3.21/main/aarch64/APKINDEX.tar.gz
Fetching URL: GET https://dl-cdn.alpinelinux.org/alpine/v3.21/main/aarch64/curl-8.14.1-r2.apk
…
Writing manifest list to image destination
loading assets file from: assets.json     # ← the export step copies blobs to blobs_dir
```

> **It built for `linux/arm64` only** — the host's architecture. By default `htvend_image`
> targets just the machine you're on, so getting started needs no qemu/binfmt setup. Ask
> for more with the `platforms` attribute (see
> [Multi-architecture images](#multi-architecture-images)).

Two things now exist that didn't before.

**`app/assets.json`** — the lockfile, written back into your source tree to commit. Every
URL the build fetched, with its content hash and headers:

```console
$ jq 'length' app/assets.json
15

$ jq 'to_entries[0]' app/assets.json
{
  "key": "https://dl-cdn.alpinelinux.org/alpine/v3.21/community/aarch64/APKINDEX.tar.gz",
  "value": {
    "Sha256": "e7328007eaba3996f18cdacfc24f5b538dfa4546454ab6a528eb2cf4256ff8b8",
    "Headers": {
      "Content-Length": "1974329",
      "Content-Type": "application/octet-stream"
    }
  }
}
```

**`blobs_dir`** — the actual bytes, named by sha256 (content-addressed, so it can be
shared by many images):

```console
$ ls /srv/shared/htvend-blobs | head -3
1699d1701542e1e446b082c6a4c966422b558df84b45c201be60dd384b8c0f21
1832327faf048390adc33852575d37c7ba155e064a339e78b9bd81983a8c7a00
189be98c9ce9ec5bd6709ea489bd236f5a6c971bf70e074f9a3ea24585f9d5ec
$ ls /srv/shared/htvend-blobs | wc -l
15
```

Commit `app/assets.json`. The blobs you keep in `blobs_dir` (or re-export / fetch from S3
— see the appendix); they're what makes the offline build possible. See
[Don't forget the blobs](../README.md#dont-forget-the-blobs) in the top-level README for
why the lockfile alone isn't enough.

### Step 4 — build it (offline)

```console
$ bazel build //app:image
…
--- Building linux/arm64 ---
Writing manifest to image destination
Target //app:image up-to-date:
  bazel-bin/app/image.oci
INFO: Build completed successfully, 3 total actions
```

No `Fetching URL` lines this time: the build runs `htvend offline`, serving every asset
from the blob store and **failing closed** if one is missing. In podman mode the
container even runs with `--network=none`, so a plain `bazel build //app:image` is itself
a no-network test.

The output is a standard OCI layout under `bazel-bin/app/image.oci`:

```console
$ jq '.manifests[0].digest' bazel-bin/app/image.oci/index.json
"sha256:b4746f429335270ee0f367e04eb63b8a26cb6210cb76899fb679830a4eb192f4"
```

Feed that into another rule (e.g. an `oci_push` from rules_oci / rules_img) or load it
directly: `podman pull oci:bazel-bin/app/image.oci`.

### Step 5 — prove it's reproducible

The whole point is that the offline build is deterministic. Wipe Bazel's cache and build
again; the image digest is identical:

```console
$ bazel clean
$ bazel build //app:image
$ jq -r '.manifests[0].digest' bazel-bin/app/image.oci/index.json
sha256:b4746f429335270ee0f367e04eb63b8a26cb6210cb76899fb679830a4eb192f4   # ← unchanged
```

That digest check is exactly what [`../e2e-bazel-test.sh`](../e2e-bazel-test.sh) asserts.

## The mental model

- **`assets.json` is a lockfile** — URLs, headers, sha256s; no bytes. **Blobs** are the
  bytes, content-addressed, stored separately (a directory here, S3 in the appendix).
- **You re-lock on your schedule, not upstream's.** `bazel run :image.lock` only needs
  re-running when an *external* dependency changes — a new/updated base image, a version
  bump in `requirements.txt`/`pom.xml`, a new package in the `Dockerfile`. Editing your
  own application source doesn't need a re-lock; `bazel build :image` just replays the
  existing lockfile + blobs.
- **One source of truth for where blobs live.** The backend repository (`app_blobs`) both
  *reads* blobs for the offline build and tells the lock run *where to write* them — via a
  `:blobs_info` target generated alongside `:blobs`. You set `blobs_dir` once, on the
  repository.

## Build configuration

Pass build-time options through `htvend_image` attributes:

- **`dockerfile`** — which Dockerfile to build (default `"Dockerfile"`); set it when a
  package has more than one.
- **`env`** — environment variables for the build, e.g. to tell `build-img-with-proxy`
  where to mount a CA truststore or Maven settings file:

  ```python
  htvend_image(
      name = "image",
      blobs = "@app_blobs//:blobs",
      dockerfile = "Dockerfile.app",
      env = {
          "JKS_KEYSTORE_PATH": "/etc/ssl/certs/java/cacerts",
          "MVN_SETTINGS_PATH": "/root/.m2/settings.xml",
      },
  )
  ```

  Most builds need none of these — see the header of
  [`cli/scripts/build-img-with-proxy`](../cli/scripts/build-img-with-proxy) for the full
  list of variables it reads.

- **`storage_driver`** — buildah storage driver for the offline build. Empty (default)
  uses each mode's natural default: **direct/RBE mode uses `vfs`** so the worker needs no
  `/dev/fuse`; podman mode uses the tool image's overlay (with the `/dev/fuse` device it
  already passes). Set `"overlay"` or `"vfs"` to override. See
  [Remote execution (RBE)](#remote-execution-rbe).

### Multiple images in one package

`htvend_image` takes a `lockfile_name` (default `assets.json`). To host several images in
one package, give each its own name, lockfile and Dockerfile:

```python
htvend_image(
    name = "a",
    blobs = "@a_blobs//:blobs",
    dockerfile = "Dockerfile.a",
    lockfile_name = "a.assets.json",
)
# -> //pkg:a and //pkg:a.lock
```

The matching blob backend must set `assets_json` to the same `lockfile_name`.

### Multi-architecture images

By default a build targets only the **host architecture**, so getting started needs no
cross-arch setup. To build a multi-arch manifest, list the `os/arch` platforms — the lock
captures all of them in one run, and the offline build replays them into a manifest list:

```python
htvend_image(
    name = "image",
    blobs = "@app_blobs//:blobs",
    platforms = ["linux/amd64", "linux/arm64"],
)
```

> Foreign architectures need the matching `binfmt` handlers on the host —
> `apt-get install qemu-user-static` registers them. A single `assets.json` holds the
> assets for every architecture (different arches just reference different URLs); commit
> the combined lockfile and `bazel build :image` produces the multi-arch OCI layout
> offline.
>
> **Why an attribute and not Bazel `--platforms`?** The fan-out happens *inside* buildah
> (one action emitting a manifest list), not via Bazel per-platform transitions, so the
> natural knob is this list of buildah `os/arch` strings.

### Pinning the tool image

`htvend_image` defaults to the published `ghcr.io/continusec/htvend` image, pinned by
digest (`DEFAULT_HTVEND_IMAGE` in [`../rules/image.bzl`](../rules/image.bzl)). podman uses
a matching local image if present (e.g. after `cd cli && make image`), otherwise pulls it.
Override per-target with `image = "ghcr.io/continusec/htvend@sha256:..."`; pin by digest
for fully reproducible builds.

## Remote execution (RBE)

By default `bazel build :image` runs under **podman** on the local machine, with
`--network=none`. That path is deliberately **local-only**: the rule tags the action
`local` + `no-sandbox` because podman/buildah need real host devices (`/dev/fuse`), their
own user+mount namespaces, and `$HOME` container storage — none of which survive Bazel's
sandbox or a remote worker.

For RBE, switch to **direct mode**:

```bash
bazel build //app:image --@rules_htvend//:exec_mode=direct
```

In direct mode the rule doesn't shell out to podman. It runs `htvend`, `buildah`, and
`build-img-with-proxy` straight from `PATH`, with the build context and blobs as declared
Bazel inputs and no network — so the action is sandbox- and remote-eligible (the
`local`/`no-sandbox` tags drop). You can also set `exec_mode = "direct"` per target.

### A `--config=rbe` bazelrc convention

`exec_mode` only controls *how the action runs*; RBE itself still needs
`--remote_executor` etc. Bundle both into a `build:rbe` config group so they stay in sync —
users without RBE get plain `bazel build` (podman, local); users with RBE add
`--config=rbe`:

```ini
# .bazelrc
build:rbe --@rules_htvend//:exec_mode=direct
build:rbe --remote_executor=grpc://rbe.example.com:443
build:rbe --remote_instance_name=...
```

```bash
bazel build //app:image              # local, podman
bazel build //app:image --config=rbe # remote, direct mode
```

See [`../examples/.bazelrc`](../examples/.bazelrc) for a working block.

### Providing the tooling on the worker

Direct mode expects `htvend`/`buildah`/`build-img-with-proxy` to already be on the
worker's `PATH`. The simplest, most robust way is to **make the worker's container image
the htvend tool image itself** (it bundles all three).

`htvend_image` sets this up for you: it points the target's `exec_properties` at the same
image used for the podman path —

```python
{
    "container-image": "docker://" + image,  # = DEFAULT_HTVEND_IMAGE unless overridden
    "OSFamily": "linux",
}
```

— so there's one source of truth for the image digest. This only takes effect under
`--@rules_htvend//:exec_mode=direct`. If your backend uses different platform keys, or you
want a different worker pool, pass your own `exec_properties` to override the default
entirely. The action runs with `use_default_shell_env = True`, so the three tools must be
on the action's `PATH` (extend with `--action_env=PATH=...` if needed; in the tool image
they're in `/usr/local/bin`).

buildah still needs **mount privilege** on the worker to set up each `RUN` container —
that's buildah's nature, not something Bazel removes. The next section pins down exactly
what the worker pool must allow.

### Worker privilege requirements

The custom worker image is the *easy* part: `container-image` is a standard REAPI platform
property, and pointing a worker pool at the tool image is an ordinary Buildbarn operation.
What an RBE team actually has to sign off on is the **privilege buildah needs to execute
`RUN` instructions**. Direct mode keeps that as small as practical:

- **Storage driver defaults to `vfs`**, so the worker needs **no `/dev/fuse`** (overlay's
  fuse-overlayfs is what would otherwise need it). The output image is byte-identical to
  the overlay build — vfs just copies where overlay would mount, trading some speed/disk.
  Override with the `storage_driver` attribute (e.g. `"overlay"` if your worker *does*
  provide a fuse device).
- **A non-overlay writable `/var/tmp`** (e.g. `tmpfs: [/var/tmp:exec]`) — buildah stages an
  overlay mount of the *build context* there, and overlay-on-overlay (the container's own
  rootfs) is rejected by the kernel. Needed even with the vfs driver.
- **Raised `RLIMIT_NOFILE`** (e.g. `nofile: 1048576`) — buildah raises it for build
  containers; start high so it needn't hold `CAP_SYS_RESOURCE`.
- **Mount privilege for the `RUN` containers** — irreducible (every `RUN` bind-mounts
  `/proc`, `/dev`, secrets, …). Grant it one of two equivalent ways:

  | | added caps | seccomp | apparmor¹ | how `RUN`'s mounts happen |
  |---|---|---|---|---|
  | **Strategy P** | `SYS_ADMIN` | **default** (filter stays on) | `unconfined` | real `CAP_SYS_ADMIN` — assumes the action runs as root² |
  | **Strategy U** | none | `unconfined` | `unconfined` | buildah creates a user namespace (`unshare(CLONE_NEWUSER)`, which the default seccomp profile blocks) |

  ¹ `apparmor=unconfined` is only needed where the runtime applies a restrictive default
  LSM profile (e.g. Docker's `docker-default`, which denies the mounts). A custom apparmor
  profile allowing just buildah's mounts is a finer-grained alternative; on hosts without
  apparmor it doesn't apply at all.
  ² true on the bb-deployments runner, whose image is `User=0`; bb_runner sets no uid.

  Strategy P generally sits better with security teams — it keeps the seccomp filter on;
  Strategy U avoids granting `CAP_SYS_ADMIN`. **Neither runs on a fully locked-down generic
  runner** — some mount privilege is unavoidable for image builds that execute `RUN`.

### Verified against a local Buildbarn cluster

Both strategies are verified end-to-end against `buildbarn/bb-deployments`' docker-compose
deployment, using its `*-hardlinking-ubuntu22-04` worker/runner pair with the runner's
image swapped for the htvend tool image and its `container-image` property set to the exact
`DEFAULT_HTVEND_IMAGE` (including digest):

```bash
bazel build //app:image \
    --@rules_htvend//:exec_mode=direct \
    --remote_executor=grpc://localhost:8980 \
    --remote_instance_name=hardlinking
```

The action dispatches to the runner ("Runner: remote"), which has `network_mode: none` —
no network interfaces at all — and the three tools on `PATH` from the image. Both runner
privilege sets below produced the **byte-identical** OCI image (`--network=none`, vfs
driver, **no `/dev/fuse`**):

- **Strategy P** — `cap_add: [SYS_ADMIN]`, `security_opt: [apparmor=unconfined]` (default
  seccomp), `tmpfs: [/var/tmp:exec]`, `ulimits.nofile: 1048576`.
- **Strategy U** — `security_opt: [seccomp=unconfined, apparmor=unconfined]` (no added
  caps), `tmpfs: [/var/tmp:exec]`, `ulimits.nofile: 1048576`.

These are properties of "a container that runs buildah", independent of `rules_htvend`.

## FAQ

### Why aren't the `assets.json` lockfiles checked into this repo?

Two reasons:

1. **A lockfile isn't enough on its own.** `assets.json` only records URLs, headers and
   sha256s — not the bytes. Building offline also needs the **blobs** in a blob store
   (see [Don't forget the blobs](../README.md#dont-forget-the-blobs)). This is a *public*
   repo, and the shared blobs for these examples live in a private S3 bucket we're not
   going to publish here — so a checked-in lockfile would only be half the picture anyway.

2. **Without the blobs, a committed lockfile would force a re-lock anyway.** Upstream
   assets change over time — the Alpine package indexes (`APKINDEX.tar.gz`) in the
   walkthrough are republished at least daily, so a day-old lockfile already points at
   content that's been replaced upstream. With the blobs saved this is a *non-issue* —
   indeed it's the whole point of htvend: `assets.json` plus the captured blobs let you
   keep serving that exact (now "stale") content into your build for as long as you like,
   completely insulated from upstream churn, until *you* decide to take the updates by
   re-running the lock. But here, with no saved blobs to fall back on, a checked-in
   lockfile referencing vanished upstream content would just send you back to
   `bazel run //…:image.lock` before you could build.

`htvend verify --fetch` tries to re-fetch anything missing from where it used to live, and
`--repair` goes further and updates the lockfile to whatever upstream serves *now* (it
implies `--fetch`) — but once the bytes are gone upstream, neither can bring them back.
That's exactly the situation htvend is built to avoid: capture the blobs once and you
never depend on upstream still having them.

In **your own** project you generally *do* commit `assets.json`, paired with a blob store
you control (a directory you ship alongside the repo, or your own S3 bucket) — and then
staleness works *for* you: the build keeps replaying the pinned content indefinitely until
you choose to re-lock. The examples here skip the checked-in lockfile precisely because
that blob store is private; `bazel run //alpine-img:image.lock` regenerates it on demand.

---

## Appendix: using S3 as the blob store

Everything above used the local directory backend. For sharing blobs across a team or CI —
so a clean checkout can build offline without first re-running the lock — swap the backend
for **`htvend_blobs_s3_repository`**. The image-side wiring is unchanged; only `MODULE.bazel`
and credentials differ.

```python
htvend_blobs_s3_repository = use_repo_rule("@rules_htvend//:blobs_s3_repository.bzl", "htvend_blobs_s3_repository")

htvend_blobs_s3_repository(
    name = "app_blobs",
    assets_json = "//app:assets.json",
    s3_bucket = "your-bucket",
    s3_prefix = "blobs/",
)
```

`htvend_image(name = "image", blobs = "@app_blobs//:blobs")` is exactly as before:
`:image.lock` exports captured blobs **to** `your-bucket`/`blobs/`, and `:image` downloads
them **from** there — both read the bucket/prefix from the backend's `:blobs_info`, so it's
specified only once.

Blobs are fetched by sha256 from `https://<bucket>.s3.amazonaws.com/<prefix><sha>` and
hash-verified; already-present blobs are skipped. Auth uses the
[tweag credential helper](https://github.com/tweag/credential-helper), which reads the
standard AWS chain. Wire it up once in the consumer repo:

```ini
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
bazel run @tweag-credential-helper//installer   # once per machine
```

The lock run's export to S3 uses your `~/.aws` credentials (mounted into the tool image).
For the fully-local directory flow end to end, the [`examples/`](../examples) workspace
and [getting-started.md](./getting-started.md) have you covered.
