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
## Quickstart

Here's a simple example:

```bash
mkdir test
cd test

htvend build -- curl https://www.google.com.au
```

Output:

```
INFO[0000] Not cached: https://www.google.com.au/       
INFO[0000] Fetching URL: https://www.google.com.au/     
... (contents) ...
```

Creates `assets.json` in your directory, with contents:

```json
{
  "https://www.google.com.au/": {
    "Sha256": "500f6cf6d3c3e33210612f92ad9fced116932293b36aedd33e836acf3b964e34",
    "Size": 17536,
    "Headers": {
      "Content-Type": "text/html; charset=ISO-8859-1"
    }
  }
}
```

If you now disconnect your internet (for example turn wifi off on your laptop), you can run:

```bash
htvend offline -- curl https://www.google.com.au
```

and the same contents are output but without making upstream connections to the internet.

```
INFO[0000] Found (manifest): https://www.google.com.au/
... (contents) ...
```

But if instead you run:

```bash
htvend offline -- curl https://www.bing.com
```

Then a 404 will be returned, and a log message printed:

```
INFO[0000] Not cached: https://www.bing.com/            
WARN[0000] missing asset for URL: https://www.bing.com/ 
```

To package up the assets (e.g. to transfer to a different environment), run:

```bash
htvend export
```

and this creates:

```
blobs/
└── 500f6cf6d3c3e33210612f92ad9fced116932293b36aedd33e836acf3b964e34
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
JKS_KEYSTORE_FILE=/tmp/htvend1586023741/cacerts.jks
...
```

When the proxy server receives a URL that is found in `assets.json`, then that content is served, along with any relevant headers in that file.

If it isn't found, then if invoked as `htvend build`, it will be fetched from upstream, or if invoked as `htvend offline`, a 404 not found response will be served.

## When is this useful?

This is useful for a number of reasons, including:

1. Some environments (such as air-gapped networks) don't have internet access. Here you can supply a directory of blobs instead.
2. Assets on the internet often change, and not always on a schedule that supports your team.
3. Assets on the internet can become unavailable due to commercial, geopolitical or other reasons (e.g. Dockerhub rate limits), or a maintainer simply deleting their repository.

Perhaps most importantly, this lets you accept changes on your schedule. If you have to make a small change to a script that lives inside of an image to address a production issue, this makes it easy to make that change without inavertently bringing in additional changes due to other upstream changes that are pulled in via an otherwise uncontrolled image build process.

## `htvend`

This is the main tool built by this repo.

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
  clean    Clean various files, see htvend clean --help for details
  export   Export referenced assets to directory
  offline  Serve assets to command, don't allow other outbound requests
  verify   Verify and fetch any missing assets in the manifest file
```

### `htvend build`

Runs the passed subprocess to create/update `assets.json` in your current directory.

After setting up a proxy server with a self-signed certificate, it will set the relevant environment variables and execute a sub-command. If none specified, an interactive shell will be made.

See `make assets.json` for an example that creates a manifest for the dependcies of this project.

```
Usage:
  htvend [OPTIONS] build [build-OPTIONS] [COMMAND] [ARG...]

Application Options:
  -C, --chdir=                         Directory to change to before running. (default: .)
  -v, --verbose                        Set for verbose output. Equivalent to setting LOG_LEVEL=debug

Help Options:
  -h, --help                           Show this help message

[build command options]
      -m, --manifest=                  File to put manifest data in (default: ./assets.json)
          --blobs-dir=                 Common directory to store downloaded blobs in (default: ${XDG_DATA_HOME}/htvend/cache/blobs)
          --cache-manifest=            Cache of all downloaded assets (default: ${XDG_DATA_HOME}/htvend/cache/assets.json)
      -l, --listen-addr=               Listen address for proxy server (:0) will allocate a dynamic open port (default: 127.0.0.1:0)
      -t, --with-temp-dir=             List of temporary directories to be creating when running this command. Env vars will be be pointing to these for the
                                       sub-process.
          --set-env-var-ssl-cert-file= List of environment variables that will be set pointing to the temporary CA certificates file in PEM format. (default:
                                       SSL_CERT_FILE)
          --set-env-var-jks-keystore=  List of environment variables that will be set pointing to the temporary CA certificates file in JKS format. (default:
                                       JKS_KEYSTORE_FILE)
          --set-env-var-http-proxy=    List of environment variables that will be set pointing to the proxy host:port. (default: HTTP_PROXY, HTTPS_PROXY,
                                       http_proxy, https_proxy)
          --set-env-var-no-proxy=      List of environment variables that will be set blank. (default: NO_PROXY, no_proxy)
          --no-cache-response=         Regex list of URLs to never store in cache. Useful for token endpoints. (default: ^http.*/v2/$, /token\?)
          --cache-header=              List of headers for which we will cache the first value. (default: Content-Type, Content-Encoding, X-Checksum-Sha1)
          --force-refresh              If set, always fetch from upstream (and save to both local and global cache).
          --clean                      If set, reset local blob list to empty before running.

[build command arguments]
  COMMAND:                             Sub-process to run. If not specified an interactive-shell is opened
  ARG:                                 Arguments to pass to the sub-process
```

### `htvend offline`

This runs the specified sub-porcess with a proxy which only serves the contents referenced in `assets.json`. Anything else will return a 404 not found error.

`make offline` does this for this repository.

If you have `unshare` installed, then a good way to *really* verify that you are offline can be as follows:

```bash
unshare -r -n -- \
  bash -c "ip link set lo up && make offline"
```

The `unshare -r -n` runs the sub-command in a new namespace with no networks. The `ip link set lo up` creates a loopback interface in that empty namespace so that `htvend` can create a server that it's sub-command can then hit.

By default all blobs are saved to and retrieved from `${XDG_DATA_HOME}/htvend/cache/blobs` (`XDG_DATA_HOME` defaults to `~/.local/share`).

A cache `assets.json` is also saved at `${XDG_DATA_HOME}/htvend/cache/assets.json`, and this is useful during rebuilds of `assets.json` to avoid needing to connect to upstream servers more than neccessary.

```
Usage:
  htvend [OPTIONS] offline [offline-OPTIONS] [COMMAND] [ARG...]

Application Options:
  -C, --chdir=                         Directory to change to before running. (default: .)
  -v, --verbose                        Set for verbose output. Equivalent to setting LOG_LEVEL=debug

Help Options:
  -h, --help                           Show this help message

[offline command options]
          --blobs-dir=                 Common directory to store downloaded blobs in (default: ${XDG_DATA_HOME}/htvend/cache/blobs)
          --cache-manifest=            Cache of all downloaded assets (default: ${XDG_DATA_HOME}/htvend/cache/assets.json)
      -m, --manifest=                  File to put manifest data in (default: ./assets.json)
      -l, --listen-addr=               Listen address for proxy server (:0) will allocate a dynamic open port (default: 127.0.0.1:0)
      -t, --with-temp-dir=             List of temporary directories to be creating when running this command. Env vars will be be pointing to these for the
                                       sub-process.
          --set-env-var-ssl-cert-file= List of environment variables that will be set pointing to the temporary CA certificates file in PEM format. (default:
                                       SSL_CERT_FILE)
          --set-env-var-jks-keystore=  List of environment variables that will be set pointing to the temporary CA certificates file in JKS format. (default:
                                       JKS_KEYSTORE_FILE)
          --set-env-var-http-proxy=    List of environment variables that will be set pointing to the proxy host:port. (default: HTTP_PROXY, HTTPS_PROXY,
                                       http_proxy, https_proxy)
          --set-env-var-no-proxy=      List of environment variables that will be set blank. (default: NO_PROXY, no_proxy)
          --dummy-ok-response=         Regex list of URLs that we return a dummy 200 OK reply to. Useful for some Docker clients. (default: ^http.*/v2/$)

[offline command arguments]
  COMMAND:                             Sub-process to run. If not specified an interactive-shell is opened
  ARG:                                 Arguments to pass to the sub-process
```

### `htvend export`

This copies all cached blobs referred to by `assets.json` to a directory of your choosing. This is useful when packaging your assets to send to another environment (which may not have internet access).

`make blobs` runs this for the `assets.json` file in this repo and creates the `blobs` directory.

```
Usage:
  htvend [OPTIONS] export [export-OPTIONS]

Application Options:
  -C, --chdir=                Directory to change to before running. (default: .)
  -v, --verbose               Set for verbose output. Equivalent to setting LOG_LEVEL=debug

Help Options:
  -h, --help                  Show this help message

[export command options]
          --blobs-dir=        Common directory to store downloaded blobs in (default: ${XDG_DATA_HOME}/htvend/cache/blobs)
          --cache-manifest=   Cache of all downloaded assets (default: ${XDG_DATA_HOME}/htvend/cache/assets.json)
      -m, --manifest=         File to put manifest data in (default: ./assets.json)
      -o, --output-directory= Directory to export blobs to. (default: ./blobs)
```


### `htvend verify`

Iterates through all referenced and confirm they exist locally and with the correct SHA256.

If `--fetch` is set, it tries to fetch anything missing.

If `--repair` is set, then the local manifest is updated if the content has changed since.

`make blobs` runs this for the `assets.json` file in this repo, it also runs `htvend export`.

```
Usage:
  htvend [OPTIONS] verify [verify-OPTIONS]

Application Options:
  -C, --chdir=                 Directory to change to before running. (default: .)
  -v, --verbose                Set for verbose output. Equivalent to setting LOG_LEVEL=debug

Help Options:
  -h, --help                   Show this help message

[verify command options]
          --blobs-dir=         Common directory to store downloaded blobs in (default: ${XDG_DATA_HOME}/htvend/cache/blobs)
          --cache-manifest=    Cache of all downloaded assets (default: ${XDG_DATA_HOME}/htvend/cache/assets.json)
      -m, --manifest=          File to put manifest data in (default: ./assets.json)
          --no-cache-response= Regex list of URLs to never store in cache. Useful for token endpoints. (default: ^http.*/v2/$, /token\?)
          --cache-header=      List of headers for which we will cache the first value. (default: Content-Type, Content-Encoding, X-Checksum-Sha1)
          --fetch              If set, fetch missing assets
          --repair             If set, replace any missing assets with new versions currently found (implies fetch). May still require a rebuild afterwards (e.g.
                               if they trigger other new calls).
```

### `htvend clean`

Removes any dangling blobs (ie not referred to by global `assets.json` cache) from global cache blobs directory.

Pass `--all` to remove entire global cache.

```
Usage:
  htvend [OPTIONS] clean [clean-OPTIONS]

Application Options:
  -C, --chdir=              Directory to change to before running. (default: .)
  -v, --verbose             Set for verbose output. Equivalent to setting LOG_LEVEL=debug

Help Options:
  -h, --help                Show this help message

[clean command options]
          --blobs-dir=      Common directory to store downloaded blobs in (default: ${XDG_DATA_HOME}/htvend/cache/blobs)
          --cache-manifest= Cache of all downloaded assets (default: ${XDG_DATA_HOME}/htvend/cache/assets.json)
          --all             If set, remove entire shared global cache.
```

## Frequently asked questions

### Can this work with building Docker / OCI images?

Yes. Packaging software into OCI Images is a very useful way to distribute software.

Further using a `Dockerfile` to populate `assets.json` is an excellent way to ensure that a build is done from scratch (that is, it pulls through all needed assets) and thus is a great way of producing a canonical `assets.json` file.

However at time of writing none of the image building tools evaluated make full and effective use of `HTTP_PROXY` and `SSL_CERT_FILE` values.

We have a (temporary, until PRs are accepted) fork of the `buildah` tool that has a number of small patches that enable it to work in this manner, and we have that packaged up here:
<https://github.com/aeijdenberg/buildah>

See [oci-image-building.md](./oci-image-building.md) for details on how to use this.

### Isn't `go mod vendor` a better solution for Go code?

Yes it is. We use the `assets.json` in this repo as an example only - not all languages are as good as Go.

### Why is this needed, can't we just ship built images around?

Shipping built images around might work well for your use-case.

This tool recognises that many projects end up being a combination of public upstream images / packages / assets and private application source code.

The intent is to help make it easier to make changes to the private application part without pulling in any other changes from the internet.

### Can specialised pull through caches like Artifactory and Nexus serve the same purpose?

Yes, they likely can. However they can be tricky to setup and may require specialist configuration for each package type (e.g. Maven vs Docker vs apt vs Python) and modification of each `Dockerfile` to use.

This project tests the hypothesis that we can do this at a simple HTTP layer.

### Is enterprise support available?

Yes. Please contact [info@continusec.com](mailto:info@continusec.com) for information and pricing for enterprise support by our Australia-based local team.
