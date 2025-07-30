# Introduction

`htvend` is a tool to help you capture any internet dependencies needed in order to perform a task.

It builds a manifest of internet assets needed, which you can check-in with your project.

The idea being that this serves as an upstream package lock file for any asset type, and that you can re-use this to rebuild your application if the upstream assets are removed, or if you are without internet connectivity.

## Installation

To build just `htvend`, you need [Go](https://go.dev/dl/) installed and then:

```bash
make

# optional, copies target/htvend to /usr/local/bin
sudo make install
```

Run `htvend --help`:

```
Usage:
  htvend [OPTIONS] <command>

Application Options:
  -C, --chdir=   Directory to change to before running. (default: .)
  -v, --verbose  Set for verbose output. Equivalent to setting LOG_LEVEL=debug

Help Options:
  -h, --help     Show this help message

Available commands:
  build    Run command to create/update the manifest file
  export   Export referenced assets to directory
  offline  Serve assets to command, don't allow other outbound requests
  verify   Verify and fetch any missing assets in the manifest file
```

## Quickstart

Here's a simple example:

```bash
htvend build -- curl https://www.google.com.au
```

Creates `assets.json` in your directory, with contents:

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
```

When the proxy server receives a URL that is found in `assets.json`, then that content is served.

If it isn't found, then if invoked as `htvend build`, it will be fetched from upstream, and if invoked as `htvend offline`, an error response will be served.

By default all blobs are saved to and retrieved from `${XDG_DATA_HOME}/htvend/blobs` (`XDG_DATA_HOME` defaults to `~/.local/share`). The `htvend export` command demonstrated above copies any references in the current dir `assets.json` to an `assets` directory in the current dir.

A cache `assets.json` is also saved at `${XDG_DATA_HOME}/htvend/cache.yml`, and this is useful during rebuilds of `assets.json` to avoid needing to connect to upstream servers more than neccessary.

## When is this useful?

This is useful for a number of reasons, including:

1. Some environments (such as air-gapped networks) don't have internet access. Here you can supply a directory of blobs instead.
2. Assets on the internet often change, and not always on a schedule that supports your team.
3. Assets on the internet can become unavailable due to commercial, geopolitical or other reasons (e.g. Dockerhub rate limits), or a maintainer simply deleting their repository.

Perhaps most importantly, this lets you accept changes on your schedule. If you have to make a small change to a script that lives inside of an image to address a production issue, this makes it easy to make that change without inavertently bringing in additional changes due to other upstream changes that are pulled in via an otherwise uncontrolled image build process.

## Does this work with building Docker / OCI images?

Yes. Packaging software into OCI Images is a very useful way to distribute software.

Further using a `Dockerfile` to populate `assets.json` is an excellent way to ensure that a build is done from scratch (that is, it pulls through all needed assets) and thus is a great way of producing a canonical `assets.json` file.

However at time of writing none of the image building tools evaluated make full and effective use of `HTTP_PROXY` and `SSL_CERT_FILE` values.

We have a (temporary, until PRs are accepted) fork of the `buildah` tool that has a number of small patches that enable it to work in this manner, and we have that packaged up here:
<https://github.com/aeijdenberg/buildah>

See [README-oci-image-building.md](./README-oci-image-building.md) for details on how to use this.

## Frequently asked questions

### When I run `htvend verify` the upstream asset now has a different SHA256, how can I get the original?

As part of your workflow it's important to ensure that the assets that you care about are saved.

When `htvend build` is run, assets are saved to `~/.local/share/htvend/blobs` rather than your working directory (which is where `assets.json` is saved).

This is so that `assets.json` can be easily commited with your source code. To save the other assets, run `htvend export` (see `--help`) to collect the referenced assets in an `assets/` directory.

We have 2 options for updating a `assets.json`.

1. Full rebuild, run: `htvend build --force-refresh --clean -- <your build command>` The `--force-refresh` tell it to always pull a new asset from upstream, and `--clean` tell it to clear the `assets.json` before starting.
2. Update the hashes to current values, run: `htvend verify --fetch --repair` - this will attempt to fetch any missing assets, then rather than complain about incorrect hashes, it will replace the value in `assets.json` with the new hash.

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
