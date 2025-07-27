# Introduction

`htvend` is a tool to help you capture any internet dependencies needed in order to perform a task.

It builds a manifest of internet assets needed, which you can check-in with your project.

The idea being that this serves as an upstream package lock file for any asset type, and that you can re-use this to rebuild your application if the upstream assets are removed, or if you are without internet connectivity.

## Installation

To build just `htvend`, you need [Go](https://go.dev/dl/) installed:

```bash
make target/htvend
```

Run:

```bash
./target/htvend --help
```

This repo also includes a patched `buildah` (`htvend-buildah`) and helper script `htvend-buildah-build`. The patched `buildah` includes a number of small PRs that are not yet merged upstream.

To install these you will need the additional dependencies listed in the [buildah instllation instructions](https://github.com/containers/buildah/blob/main/install.md#installation-from-github), and these are easy to install with your regular distribution package manager.

```bash
make
sudo make install # copies binaries to /usr/local/bin/
```

## Quickstart

Here's a simple example:

```bash
htvend build -- curl https://www.google.com.au
```

Creates `blobs.yml` in your directory, with contents:

```bash
https://www.google.com.au/:
  Headers:
    Content-Type: text/html; charset=ISO-8859-1
  Sha256: bd9c1762501a93c9ed806477a7fbf3db427aa7b8929d8b4d9839ca89ea560f24
```

If you now disconnect your internet (for example turn wifi off on your laptop), you can run:

```bash
htvend offline -- curl https://www.google.com.au
```

and the same contents are output but without making upstream connections to the internet.

But if instead you run:

```bash
htvend offline -- curl https://www.bing.com
```

Then a 404 will be returned, and a log message printed:

```
WARN[0000] missing asset for URL: https://www.bing.com/ 
```

To package up the assets (e.g. to transfer to a different environment), run:

```bash
htvend export
```

and this creates:

```
assets
└── bd9c1762501a93c9ed806477a7fbf3db427aa7b8929d8b4d9839ca89ea560f24
```

## How does it work?

When invoked as `htvend build` or `htvend offline` it creates a local HTTP and HTTPS proxy server on a dynamic port, with a self-signed certificate.

It then runs the specified child process with appropriate environment variables specified. For example:

```bash
htvend build -- env
```

Shows:

```bash
https_proxy=http://127.0.0.1:46307
http_proxy=http://127.0.0.1:46307
no_proxy=
HTTPS_PROXY=http://127.0.0.1:46307
HTTP_PROXY=http://127.0.0.1:46307
NO_PROXY=
SSL_CERT_FILE=/tmp/htvend1586023741/cacerts.pem
...
```

When a URL is requested that is found in `blobs.yml`, then that content is served.

If it isn't found, then if invoked as `htvend build`, it will be fetched from upstream, and if invoked as `htvend offline` then an error response will be served.

By default all blobs are saved and retrieved from `${XDG_DATA_HOME}/htvend/blobs` (`XDG_DATA_HOME` defaults to `~/.local/share`). The `htvend export` command demonstrated above copied any references the current dir `blobs.yml` to an `assets` directory in the current dir.

A cache `blobs.yml` is also saved at `${XDG_DATA_HOME}/htvend/cache.yml`, and this is useful during rebuilds of `blobs.yml` to avoid needing to connect to upstream servers more than neccessary.

## When is this useful?

This is useful for a number of reasons, including:

1. Some environments (such as air-gapped networks) don't have internet access. Here you can supply a directory of blobs instead.
2. Assets on the internet often change, and not always on a schedule that supports your team.
3. Assets on the internet can become unavailable due to commercial, geopolitical or other reasons (e.g. Dockerhub rate limits), or a maintainer simply deleting their repository.

Perhaps most importantly, this lets you accept changes on your schedule. If you have to make a small change to a script that lives inside of an image to address a production issue, this makes it easy to make that change without inavertently bringing in additional changes due to other upstream changes that are pulled in via an otherwise uncontrolled image build process.

## Does this work with building Docker / OCI images?

Yes. Packaging software into OCI Images is a very useful way to distribute software.

We have built special support into `htvend` to make it straight-forward to use with the `buildah` tool to create OCI Image archives, which can be imported or pushed to your container management system. The special support is primarily around making the `SSL_CERT_FILE` mount available to the `RUN` instructions inside the `Dockerfile` without needing to modify the `Dockerfile`.

Building images has 2 challenges:

1. Fetching upstream base images from Docker registries such as docker.io.
2. Propagating proxy/certificate information to `RUN` instructions inside of a `Dockerfile`.

Our intent is to work without needing to make changes to a `Dockerfile` that otherwise can build with normal internet connectivity.

### Non-caching of registry /v2/ tokens

Here we make 2 main accommodations:

1. Don't cache registry token end-points during `build` mode. Controlled with `--no-cache-response=` flag. See `htvend build --help` for default list.
2. Do serve dummy 200 OK response for registry endpoint during `offline` mode. Controlled with `--dummy-ok-response=`. See `htvend offline --help` for default list.

The first is necessary so that we don't cache and save authentication information.

The second is necessary because some registry clients (including that used by `buildah`) will always perform a `GET` against the `/v2/` endpoint to check if authentication is required for furture connection (and during `offline` mode it isn't). By returning a dummy `200 OK` subsequent calls can work without putting weird data in the `blobs.yml`.

### Custom `RUN` mount to make `SSL_CERT_FILE` and other files available to build process

While `docker` and `buildah` automatically propagate `https_proxy` and other proxy variables to `RUN` commands without affecting the final image, they don't have a similar method for propagating `SSL_CERT_FILE`.

However if we set this as a Docker "secret", then effectively run `RUN --mount=secret,id=xx,env=SSL_CERT_FILE` we can make that data available to `RUN` commands.

This is awkward however as it requires modifying a `Dockerfile`, so we have contributed a patch to `buildah` that does so in a ephemeral manner, such that additional files/environment variables can be made available during each `RUN` invocation during a build, but without baking this into or otherwise affecting the final image.

We have the following pull requests open against `buildah`:

<https://github.com/containers/buildah/pull/6289>
<https://github.com/containers/buildah/pull/6285>

Until the above are merged, we have a branch with these patches at:
<https://github.com/aeijdenberg/buildah/tree/continusecbuild>

### Building Java projects

Maven is a popular tool for building Java projects. It does not use the standard `HTTP_PROXY` and related variables.

We instead create a temporary `settings.xml` file for `mvn` and put it in a temporary file. `MAVEN_SETTINGS_FILE` is set to the path of this file.

Likewise we create a custom JKS truststore containing the self-signed certificate, and put this in a temporary file. `JAVA_TRUST_STORE_FILE` is set to the path of this file.

## Frequently asked questions

### When I run `htvend verify` the upstream asset now has a different SHA256, how can I get the original?

As part of your workflow it's important to ensure that the assets that you care about are saved.

When `htvend build` is run, assets are saved to `~/.local/share/htvend/blobs` rather than your working directory (which is where `blobs.yml` is saved).

This is so that `blobs.yml` can be easily commited with your source code. To save the other assets, run `htvend export` (see `--help`) to collect the referenced assets in an `assets/` directory.

To "repair" or update the `blobs.yml` to re-point to latest images, you can run as `htvend build --force-refresh`.

For example in this repo, the `examples/alpine-img/blobs.yml` file is likely of out-of-date from when we saved it.

Run:

```bash
cd examples/alpine-img
htvend verify --fetch
```

Results in error:

```
FATA[0000] error fetching https://dl-cdn.alpinelinux.org/alpine/v3.21/community/aarch64/APKINDEX.tar.gz: error updating asset file: wrong SHA256 for https://dl-cdn.alpinelinux.org/alpine/v3.21/community/aarch64/APKINDEX.tar.gz: expected: 7d7cfb8cdf852f2b1ccc887b624dcdefbba67cf0b35a4fff571811fa9b21f3c0 received: d4958fd1f4d5130459d8c0fa0e39861f6dc723ff341d3778b6b662de11550715 
```

Rebuild with latest with:

```bash
htvend build --force-refresh -- htvend-buildah-build
```

Or, for example if you only wanted to re-fetch that particular file:

```bash
htvend build --force-refresh -- curl https://dl-cdn.alpinelinux.org/alpine/v3.21/community/aarch64/APKINDEX.tar.gz 
```

### Why is this needed, can't we just ship built images around?

Shipping built images around might work well for your use-case.

This tool recognises that many projects end up being a combination of public upstream images / packages / assets and private application source code.

The intent is to help make it easier to make changes to the private application part without pulling in any other changes from the internet.

### Can specialised pull through caches like Artifactory and Nexus server the same purpose?

Yes, they likely can. However they can be tricky to setup and may require specialist configuration for each package type (e.g. Maven vs Docker vs apt vs Python) and modification of each `Dockerfile` to use.

This project tests the hypothesis that we can do this at a simple HTTP layer.

### Is enterprise support available?

Yes. Please contact [info@continusec.com](mailto:info@continusec.com) for information and pricing for enterprise support by our Australia-based local team.
