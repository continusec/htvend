# htvend

`htvend` captures the internet dependencies needed to perform a task, records them in
a manifest (lock file) you check in alongside your project, and replays them later —
so you can rebuild even when the upstream assets have changed, disappeared, or you're
offline.

Think of it as an **upstream package lock file for any asset type**, captured at the
plain HTTP(S) layer so it works regardless of the package ecosystem (Docker/OCI, apt,
pip, Maven, raw downloads, …).

## Repository layout

| Directory | What it is |
|-----------|------------|
| [`cli/`](./cli) | The standalone Go project — the `htvend` binary. `make install`, no Bazel required. |
| [`rules/`](./rules) | `rules_htvend`: reusable Bazel rules so **other** projects can build OCI images offline. |
| [`examples/`](./examples) | A separate Bazel workspace that consumes `rules/` exactly as an external project would. |
| [`docs/`](./docs) | Documentation (see links below). |
| [`experimental/`](./experimental) | Reference artifacts not part of the active build/CI. |

## Installation

To build just `htvend` you need [Go](https://go.dev/dl/) installed:

```bash
cd cli
make

# optional, copies target/htvend (and helper scripts) to /usr/local/bin
sudo make install
```

## Quickstart

```bash
mkdir test && cd test
htvend build -- curl https://www.google.com.au
```

This creates `assets.json` in your directory:

```json
{
  "https://www.google.com.au/": {
    "Sha256": "89106a37bb7b8e803990c9589d069e4ee06ef045d2692ad78b0462c12b89f59f",
    "Headers": {
      "Content-Type": "text/html; charset=ISO-8859-1"
    }
  }
}
```

Now disconnect your internet (e.g. turn off wifi) and run:

```bash
htvend offline -- curl https://www.google.com.au
```

The same content is served, with no upstream connection. Request something that
*isn't* in the manifest and you get a 404:

```bash
htvend offline -- curl https://www.bing.com
# WARN[0000] missing asset for URL: https://www.bing.com/
```

To package the captured assets up (e.g. to move to another environment):

```bash
htvend export --output-directory=blobs
```

## How does it work?

`htvend build` / `htvend offline` stand up a local HTTP(S) proxy on a dynamic port
with a self-signed CA, then run your sub-process with the relevant environment
variables pointed at it:

```bash
htvend build -- env
# https_proxy=http://127.0.0.1:46307
# HTTPS_PROXY=http://127.0.0.1:46307
# SSL_CERT_FILE=/tmp/htvend.../cacerts.pem
# JKS_KEYSTORE_FILE=/tmp/htvend.../cacerts.jks
# ...
```

When the proxy sees a URL present in `assets.json`, it serves that content (with the
recorded headers). Otherwise, under `build` it fetches from upstream and records it;
under `offline` it returns a 404. Blobs are content-addressed by SHA256.

## Ways to use htvend

- **As a CLI** — `build` / `verify` / `export` / `offline`. Full reference:
  [docs/cli.md](./docs/cli.md).
- **Building Docker / OCI images** — capture and replay everything a `Dockerfile`
  build pulls in: [docs/oci-image-building.md](./docs/oci-image-building.md).
- **In Bazel** — let other projects build Dockerfiles hermetically with the reusable
  `rules_htvend` rules: [docs/bazel.md](./docs/bazel.md).
- **Experimental: feeding k3s** — [docs/k3s-running.md](./docs/k3s-running.md).

## When is this useful?

1. Air-gapped / offline environments — supply a directory of blobs instead of the
   internet.
2. Upstream assets change on someone else's schedule, not yours.
3. Upstream assets become unavailable (commercial, geopolitical, Dockerhub rate
   limits, or a maintainer deleting a repo).

Most importantly, it lets you **accept changes on your schedule**: make a small fix to
something inside an image without inadvertently pulling in unrelated upstream changes
from an otherwise uncontrolled build.

## FAQ

### Can this work with building Docker / OCI images?

Yes — and it's a great fit. Using a `Dockerfile` to populate `assets.json` ensures the
build pulls through everything it needs, producing a canonical lock file. See
[docs/oci-image-building.md](./docs/oci-image-building.md), and
[docs/bazel.md](./docs/bazel.md) for the Bazel rules.

### Isn't `go mod vendor` a better solution for Go code?

Yes it is. `htvend` is most useful for the long tail of assets that *don't* have a
good vendoring story — not all ecosystems are as well-served as Go.

### Why is this needed, can't we just ship built images around?

That might work for your use case. This tool recognises that many projects are a
combination of public upstream images/packages/assets and private application source
code, and aims to make it easy to change the private part without pulling in other
upstream changes.

### Can pull-through caches like Artifactory and Nexus serve the same purpose?

Likely yes, but they can be tricky to set up and often need per-ecosystem
configuration (Maven vs Docker vs apt vs Python) and `Dockerfile` changes. This project
tests the hypothesis that we can do this at a simple HTTP layer instead.

### Is enterprise support available?

Yes. Please contact [info@continusec.com](mailto:info@continusec.com) for information
and pricing for enterprise support by our Australia-based local team.
