"""HtvendBlobsInfo: lets a blobs backend advertise its own location to htvend_lock.

blobs_s3_repository.bzl and blobs_dir_repository.bzl each generate a `:blobs_info`
target alongside `:blobs`, carrying this provider. htvend_lock reads it (via the
`blobs` label passed to htvend_image) so that:

- the S3 bucket/prefix used to export blobs after a lock run matches the
  bucket/prefix the corresponding htvend_blobs_s3_repository downloads them from, and
- the local directory a lock run writes blobs into matches the directory the
  corresponding htvend_blobs_dir_repository reads them from,

one source of truth, no need to repeat `s3_bucket`/`s3_prefix`/`blobs_dir` on
htvend_image itself.
"""

HtvendBlobsInfo = provider(
    doc = "Location backing a blobs repository.",
    fields = {
        "s3_bucket": "S3 bucket blobs are stored in, or \"\" for a directory-backed repo.",
        "s3_prefix": "S3 key prefix blobs are stored under.",
        "blobs_dir": "Local directory blobs are read from, or \"\" for an S3-backed repo.",
    },
)

def _htvend_blobs_info_impl(ctx):
    return [HtvendBlobsInfo(
        s3_bucket = ctx.attr.s3_bucket,
        s3_prefix = ctx.attr.s3_prefix,
        blobs_dir = ctx.attr.blobs_dir,
    )]

htvend_blobs_info = rule(
    implementation = _htvend_blobs_info_impl,
    attrs = {
        "s3_bucket": attr.string(default = ""),
        "s3_prefix": attr.string(default = ""),
        "blobs_dir": attr.string(default = ""),
    },
)

def read_assets_json(ctx, assets_json):
    """Read a lockfile label as a dict, or {} if it doesn't exist yet.

    Repository rules run at fetch time, before the first `bazel run //pkg:image.lock`
    has had a chance to create `assets.json`. Treating a missing lockfile as empty lets
    that first lock run succeed without a manually checked-in placeholder file.

    ctx.watch() registers a dependency on the lockfile *even when it's absent*, so when
    the lock run later creates (or updates) it, Bazel refetches this repository and the
    per-blob symlinks below get (re)generated. Without it the repo would be fetched once
    with an empty/missing lockfile and never refreshed -- the offline build would then
    fail with the blobs "missing" because no symlink was ever created for them.
    """
    path = ctx.path(assets_json)
    ctx.watch(path)
    if not path.exists:
        return {}
    return json.decode(ctx.read(assets_json))
