# Experimental

These artifacts are kept for reference and are **not part of the active build or
CI**. They may be out of date and are not guaranteed to work as-is.

## `Dockerfile.githubaction`

A Debian-based image that builds `htvend`, `build-img-with-proxy`, and a buildah
binary, and was previously published to `ghcr.io/continusec/htvend` and consumed by
the offline-dependencies workflow below.

The canonical tool image is now built from [`../cli/Dockerfile`](../cli/Dockerfile)
(a small wolfi base with upstream buildah ≥ 1.44) via `cd cli && make image-push`.
This Dockerfile additionally built buildah from source — that is no longer needed
now that the `--mount` patches are upstream in buildah ≥ 1.44.

## `Dockerfile.offlinebuild`

Minimal image (`FROM golang`, `make all`) used by the offline-dependencies workflow
as the thing-to-build under htvend, to exercise htvend on its own dependencies.

## `offline-dependencies.yml`

A GitHub Actions workflow that used the published image to rebuild `assets.json`,
sync blobs to/from S3, and commit lockfile updates automatically. It has been moved
out of `.github/workflows/` so it no longer runs. To revive it, move it back and fix
the now-relative paths (`Dockerfile.offlinebuild`, `assets.json`) to point at the
`cli/` directory, and refresh the `htvend_img` tag.

## Retired: the patched-buildah fork

Earlier versions built a fork of buildah (`aeijdenberg/buildah`) to add
`--mount`/secret-env support to `RUN` instructions. Those patches are now merged
upstream and released in [buildah v1.44.0](https://github.com/podman-container-tools/buildah/releases/tag/v1.44.0),
so the fork and the old `make install-patched-buildah` flow have been removed.
`build-img-with-proxy` now calls upstream `buildah` directly.
