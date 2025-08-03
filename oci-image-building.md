# Building Docker / OCI Images

There are a number of challenges building Docker / OCI Images with respect to getting track of which assets are needed to build them.

Here we have a small number of patches on the `buildah` tool that allow it to work well to service this use-case.

## Quickstart

Our patched `buildah` can be found at: <https://github.com/aeijdenberg/buildah/tree/continusecbuild>

See further down for rationale why, but here is how to get started with a patched `buildah`.

### Ubuntu 24.04 dependencies for `buildah`

You will likely need some operating system specific libraries / configuration be installed. At time of writing, the following works for an Ubuntu 24.04 system (tested with <https://lima-vm.io/> on macOS, using template `limactl create --name=u2404 --plain template://ubuntu-24.04 -y`)

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

### Building our tools including `patched-buildah`

The following will fetch and run our tools:

```bash
# get our main repo, if not already fetched
git clone https://github.com/continusec/htvend.git
cd htvend

# fetch and build patched version locally, as well as our other binaries
make patched-buildah-src/bin/buildah all

# install patched-buildah and our other binaries to /usr/local/bin/patched-buildah
sudo make install-patched-buildah install
```

Now go and try the examples. For example:

```bash
# build the assets.json in that directory, use the --clean
htvend -C ./examples/alpine-img build --clean -- build-img-with-proxy

# verify all the assets exist
htvend -C ./examples/alpine-img verify --fetch

# run in offline mode, produce img.tar file
htvend -C ./examples/alpine-img offline -- build-img-with-proxy --tag oci-archive:img.tar
```

The following `make` targets will run the above for each example:

```bash
make img-manifests img-blobs img-tarballs sha256sums
```

## `build-img-with-proxy` script

`build-img-with-proxy` is a wrapper that called `patched-buildah` with whatever arguments are passed to it, however it detects a number of environment variables including `SSL_CERT_FILE`, `JKS_KEYSTORE_FILE` and uses these to overlay helpers files within the context when it runs.

See the script itself for which variables it uses, these are at the top.

## Patched `buildah`

The examples in this repository apply some small patches to the [buildah](https://github.com/containers/buildah) tool so that we can easily build images, while pulling through upstream assets in such a way that they can be saved out, and then replayed later to facilitate new image builds.

The patches that we apply to `buildah` are, at time of writing:

- [feat(build): add --run-mount option](https://github.com/containers/buildah/pull/6289) - authored by @aeijdenberg
- [feat: support --mount=type=secret,id=foo,env=bar](https://github.com/containers/buildah/pull/6285) - authored by @aeijdenberg
- [fix(build): make --tag oci-archive:xxx.tar work with simple images](https://github.com/containers/buildah/pull/6284) - authored by @aeijdenberg
- [build,add: add support for corporate proxies](https://github.com/containers/buildah/pull/6274) - authored by @userid0x0

These patches make it possible to invoke `buildah` with options to automatically mount various files / environment variables in any containers used for `RUN` instructions, in such a way that they don't affect the final image.

Hopefully the above (or equivalent) are merged over time, but until then we have a branch with these patches at: https://github.com/aeijdenberg/buildah/tree/continusecbuild

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
