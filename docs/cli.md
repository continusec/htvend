# `htvend` command reference

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
  verify   Verify and fetch any missing assets in the manifest file
  export   Export referenced assets to directory
  offline  Serve assets to command, don't allow other outbound requests
```

## `htvend build`

Runs the passed subprocess to create/update `assets.json` in your current directory.

After setting up a proxy server with a self-signed certificate, it sets the relevant
environment variables and executes a sub-command. If none is specified, an
interactive shell is opened.

```
Usage:
  htvend [OPTIONS] build [build-OPTIONS] [COMMAND] [ARG...]

[build command options]
          --blobs-backend=[filesystem|registry|s3] Type of blob store (default: filesystem)
          --blobs-registry=                     URL for registry to store / fetch blobs from
          --blobs-dir=                          Common directory to store downloaded blobs in (default: ${XDG_DATA_HOME}/htvend/cache/blobs)
          --cache-manifest=                     Cache of all downloaded assets (default: ${XDG_DATA_HOME}/htvend/cache/assets.json)
      -m, --manifest=                           File to put manifest data in (default: ./assets.json)
      -l, --listen-addr=                        Listen address for proxy server (:0) will allocate a dynamic open port (default: 127.0.0.1:0)
      -c, --ca-out=                             Cert file out location - defaults to a temp file
      -d, --daemon                              Run as a daemon until terminated
      -s, --single-thread                       Don't service HTTP request until previous one is complete.
      -t, --with-temp-dir=                      List of temporary directories to be created when running this command. Env vars will be pointing to these for the sub-process.
          --set-env-var-ssl-cert-file=          List of environment variables that will be set pointing to the temporary CA certificates file in PEM format. (default: SSL_CERT_FILE)
          --set-env-var-jks-keystore=           List of environment variables that will be set pointing to the temporary CA certificates file in JKS format. (default: JKS_KEYSTORE_FILE)
          --set-env-var-http-proxy=             List of environment variables that will be set pointing to the proxy host:port. (default: HTTP_PROXY, HTTPS_PROXY, http_proxy, https_proxy)
          --set-env-var-no-proxy=               List of environment variables that will be set blank. (default: NO_PROXY, no_proxy)
          --no-cache-response=                  Regex list of URLs to never store in cache. Useful for token endpoints. (default: ^http.*/v2/$, /token\?)
          --cache-header=                       List of headers for which we will cache the first value. (default: Content-Length, Docker-Content-Digest, Content-Type, Content-Encoding, X-Checksum-Sha1)
          --force-refresh                       If set, always fetch from upstream (and save to both local and global cache).
          --clean                               If set, reset local blob list to empty before running.

[build command arguments]
  COMMAND:                                      Sub-process to run. If not specified an interactive-shell is opened
  ARG:                                          Arguments to pass to the sub-process
```

## `htvend offline`

Runs the specified sub-process with a proxy which only serves the contents
referenced in `assets.json`. Anything else returns a 404 not found error.

If you have `unshare` installed, a good way to *really* verify that you are offline:

```bash
unshare -r -n -- \
  bash -c "ip link set lo up && htvend offline -- <your build command>"
```

`unshare -r -n` runs the sub-command in a new namespace with no networks; the
`ip link set lo up` creates a loopback interface in that empty namespace so that
`htvend` can serve to its sub-command.

By default all blobs are saved to and retrieved from
`${XDG_DATA_HOME}/htvend/cache/blobs` (`XDG_DATA_HOME` defaults to `~/.local/share`).
A cache `assets.json` is also saved at `${XDG_DATA_HOME}/htvend/cache/assets.json`,
useful during rebuilds to avoid connecting to upstream servers more than necessary.

```
Usage:
  htvend [OPTIONS] offline [offline-OPTIONS] [COMMAND] [ARG...]

[offline command options]
          --blobs-backend=[filesystem|registry|s3] Type of blob store (default: filesystem)
          --blobs-registry=                     URL for registry to store / fetch blobs from
          --blobs-dir=                          Common directory to store downloaded blobs in (default: ${XDG_DATA_HOME}/htvend/cache/blobs)
          --cache-manifest=                     Cache of all downloaded assets (default: ${XDG_DATA_HOME}/htvend/cache/assets.json)
      -m, --manifest=                           File to put manifest data in (default: ./assets.json)
      -l, --listen-addr=                        Listen address for proxy server (:0) will allocate a dynamic open port (default: 127.0.0.1:0)
      -c, --ca-out=                             Cert file out location - defaults to a temp file
      -d, --daemon                              Run as a daemon until terminated
      -s, --single-thread                       Don't service HTTP request until previous one is complete.
      -t, --with-temp-dir=                      List of temporary directories to be created when running this command. Env vars will be pointing to these for the sub-process.
          --set-env-var-ssl-cert-file=          List of environment variables that will be set pointing to the temporary CA certificates file in PEM format. (default: SSL_CERT_FILE)
          --set-env-var-jks-keystore=           List of environment variables that will be set pointing to the temporary CA certificates file in JKS format. (default: JKS_KEYSTORE_FILE)
          --set-env-var-http-proxy=             List of environment variables that will be set pointing to the proxy host:port. (default: HTTP_PROXY, HTTPS_PROXY, http_proxy, https_proxy)
          --set-env-var-no-proxy=               List of environment variables that will be set blank. (default: NO_PROXY, no_proxy)
          --dummy-ok-response=                  Regex list of URLs that we return a dummy 200 OK reply to. Useful for some Docker clients. (default: ^http.*/v2/$)

[offline command arguments]
  COMMAND:                                      Sub-process to run. If not specified an interactive-shell is opened
  ARG:                                          Arguments to pass to the sub-process
```

## `htvend export`

Copies all cached blobs referred to by `assets.json` to a destination of your
choosing (a directory, or another backend such as S3). Useful when packaging assets
to send to another environment, or to populate a shared blob store.

For example, `htvend export --output-directory=blobs` writes every blob referenced by
`assets.json` into a local `blobs/` directory.

```
Usage:
  htvend [OPTIONS] export [export-OPTIONS]

[export command options]
          --blobs-backend=[filesystem|registry|s3] Type of blob store (default: filesystem)
          --blobs-registry=                     URL for registry to store / fetch blobs from
          --blobs-dir=                          Common directory to store downloaded blobs in (default: ${XDG_DATA_HOME}/htvend/cache/blobs)
          --cache-manifest=                     Cache of all downloaded assets (default: ${XDG_DATA_HOME}/htvend/cache/assets.json)
      -m, --manifest=                           File to put manifest data in (default: ./assets.json)
      -o, --output-directory=                   Directory to export blobs to. (default: ./blobs)
```

The `export` command also accepts `--dest.blobs-backend`, `--dest.blobs-bucket`, and
`--dest.blobs-prefix` to push to an S3 destination (used by the Bazel `htvend_lock`
rule). See `htvend export --help` for the full set.

## `htvend verify`

Iterates through all referenced assets and confirms they exist locally with the
correct SHA256.

- `--fetch` tries to fetch anything missing.
- `--repair` updates the local manifest if the content has changed since (implies
  `--fetch`; may still require a rebuild afterwards).

```
Usage:
  htvend [OPTIONS] verify [verify-OPTIONS]

[verify command options]
          --blobs-backend=[filesystem|registry|s3] Type of blob store (default: filesystem)
          --blobs-registry=                     URL for registry to store / fetch blobs from
          --blobs-dir=                          Common directory to store downloaded blobs in (default: ${XDG_DATA_HOME}/htvend/cache/blobs)
          --cache-manifest=                     Cache of all downloaded assets (default: ${XDG_DATA_HOME}/htvend/cache/assets.json)
      -m, --manifest=                           File to put manifest data in (default: ./assets.json)
          --no-cache-response=                  Regex list of URLs to never store in cache. Useful for token endpoints. (default: ^http.*/v2/$, /token\?)
          --cache-header=                       List of headers for which we will cache the first value. (default: Content-Length, Docker-Content-Digest, Content-Type, Content-Encoding, X-Checksum-Sha1)
          --fetch                               If set, fetch missing assets
          --repair                              If set, replace any missing assets with new versions currently found (implies fetch).
```
