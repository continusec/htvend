"""htvend_blobs_repository: fetch a lockfile's blobs from S3 into Bazel.

Reads the (content-addressed) entries in a checked-in htvend lockfile and downloads
each blob from S3 by sha256, verifying the hash. Produces a `@<name>//:blobs` target
for the `blobs` attribute of htvend_image.

S3 auth is handled out-of-band by the tweag credential helper wired into the
consumer's .bazelrc (see ../docs/bazel.md); ctx.download skips blobs already present
with the matching sha256. For a credential-free local/NFS setup, see the directory
backend in blobs_dir_repository.bzl.
"""

def _htvend_blobs_impl(ctx):
    # Read the lockfile
    assets_json = ctx.read(ctx.attr.assets_json)
    assets = json.decode(assets_json)

    # Download each blob listed in assets.json
    for blob in assets.values():
        sha256 = blob["Sha256"]
        url = "https://{bucket}.s3.amazonaws.com/{prefix}{sha256}".format(
            bucket = ctx.attr.s3_bucket,
            prefix = ctx.attr.s3_prefix,
            sha256 = sha256,
        )
        ctx.download(
            url = url,
            output = "blobs/" + sha256,
            sha256 = sha256,
        )

    ctx.file("BUILD.bazel", """
load("@rules_htvend//:blobs_info.bzl", "htvend_blobs_info")

exports_files(["blobs"])

htvend_blobs_info(
    name = "blobs_info",
    s3_bucket = "{s3_bucket}",
    s3_prefix = "{s3_prefix}",
    visibility = ["//visibility:public"],
)
""".format(s3_bucket = ctx.attr.s3_bucket, s3_prefix = ctx.attr.s3_prefix))

htvend_blobs_repository = repository_rule(
    implementation = _htvend_blobs_impl,
    attrs = {
        "assets_json": attr.label(
            mandatory = True,
            allow_single_file = True,
        ),
        "s3_bucket": attr.string(mandatory = True),
        "s3_prefix": attr.string(mandatory = True),
    },
)