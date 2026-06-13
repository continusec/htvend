# Building Docker / OCI Images

There are a number of challenges building Docker / OCI Images with respect to getting track of which assets are needed to build them.

> **Just want the Bazel flow?** It runs everything through podman using the published
> tool image, so you only need `podman` + `qemu-user-static` (plus git and Bazel) — none
> of the host `buildah` setup below. See [getting-started.md](./getting-started.md).

This page covers running `build-img-with-proxy` / `buildah` **directly on the host**
(the CLI path). For that you will need `buildah` version
[1.44](https://github.com/podman-container-tools/buildah/releases/tag/v1.44.0) or newer.

### Ubuntu 24.04 dependencies for host `buildah`

To run buildah directly on the host you will likely need some operating-system-specific libraries / configuration installed. At time of writing, the following works for an Ubuntu 24.04 system (tested with <https://lima-vm.io/> on macOS, using template `limactl create --name=u2404 --plain template://ubuntu-24.04 -y`)

```bash
# install Go
curl -L https://go.dev/dl/go1.24.5.linux-arm64.tar.gz | sudo tar -C /usr/local -zx
export PATH=/usr/local/go/bin:$PATH

# install apt dependencies
sudo apt update -y
sudo apt install -y make gcc uidmap pkg-config libseccomp-dev libgpgme-dev netavark runc

# set up a uid/gid range for use by namespaces
sudo usermod --add-subuids 100000-165535 --add-subgids 100000-165535 $USER

# allow unprivileged users to use name-spaces
sudo sysctl -w kernel.apparmor_restrict_unprivileged_userns=0
```

### Building and installing our tools

`build-img-with-proxy` calls upstream `buildah` directly, so install buildah ≥ 1.44
from your distribution (or the wolfi packages), then build and install the htvend
binaries:

```bash
# get our main repo, if not already fetched
git clone https://github.com/continusec/htvend.git
cd htvend/cli

# build and install htvend + build-img-with-proxy to /usr/local/bin
make
sudo make install
```

> Earlier versions of htvend built a fork of buildah to add `--mount`/secret-env
> support to `RUN` instructions. Those patches are now upstream in buildah ≥ 1.44, so
> the fork and the old `make install-patched-buildah` flow have been removed. See
> [../experimental/README.md](../experimental/README.md).

Now try one of the [`examples/`](../examples). Run these from the repo root with
`htvend` on your `PATH`:

```bash
# build the assets.json in that directory, use --clean to start fresh
htvend -C ./examples/alpine-img build --clean -- build-img-with-proxy

# verify all the assets exist
htvend -C ./examples/alpine-img verify --fetch

# run in offline mode, produce an OCI image directory
htvend -C ./examples/alpine-img offline -- build-img-with-proxy
```

To drive the same examples through Bazel (the supported, reusable path), see
[bazel.md](./bazel.md).

## `build-img-with-proxy` script

`build-img-with-proxy` is a wrapper that calls `buildah` with whatever arguments are passed to it, however it detects a number of environment variables including `SSL_CERT_FILE`, `JKS_KEYSTORE_FILE` and uses these to overlay helper files within the context when it runs. (It honours `BUILDAH_BINARY` if you need to point it at a specific buildah.)

See the script itself for which variables it uses, these are at the top.

## `buildah` requirements

Building images this way relies on two `buildah` features that let us automatically
mount files / environment variables into any container used for a `RUN` instruction,
without modifying the original `Dockerfile` and without affecting the final image:

- [feat(build): add --mount option](https://github.com/containers/buildah/pull/6289) — authored by @aeijdenberg
- [feat: support --mount=type=secret,id=foo,env=bar](https://github.com/containers/buildah/pull/6285) — authored by @aeijdenberg

Both are merged upstream and released in
[buildah v1.44.0](https://github.com/podman-container-tools/buildah/releases/tag/v1.44.0),
so any `buildah` ≥ 1.44 works — no fork required.

## Rationale

Here we list the challenges faced, and how we mitigate them:

## Challenges and mitigations

### Build system must be child process of `htvend`

To use the environment configured by `htvend`, the build system must be a child process of `htvend` so that it can use `HTTP_PROXY`, `SSL_CERT_FILE` and other environment variables.

**Mitigation:** We choose to use the [buildah](https://github.com/containers/buildah) tool to build images as it functions without needing a daemon.

### Tokens from registries should not be cached

Most upstream image repos need a simple token to be fetched. This should not be cached, but is needed during a `htvend verify --fetch`.

**Mitigation:** We have special-support for communicating with image registries baked into `htvend`, for example see the `--no-cache-response=^http.*/v2/$` default arguments on `htvend build`

### Build tool caches image layer data and won't re-pull through

Since many upstream image assets are content addressable, many systems cache the file contents locally, such that even with a `--pull=always` (or similar) the actual blobs are not re-fetched.

**Mitigation:** Our [build-img-with-proxy](./scripts/build-img-with-proxy) script wraps `buildah` with `XDG_DATA_DIR=/some/temp/dir` before executing. This prevents `buildah` from seeing any cached layers and will always pull-through again.

### No method to pass `SSL_CERT_FILE` to `RUN` instructions in Dockerfile

Although a `Dockerfile` can use `RUN --mount=type=secret,...` to temporarily mount a file in a process, there's no way to automatically apply this to all `RUN` instructions without needing to modify the original `Dockerfile`.

**Mitigation:** We opened a number of small Pull Requests against `buildah` to add this ability.
