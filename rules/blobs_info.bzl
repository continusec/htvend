"""HtvendBlobsInfo: lets a blobs backend advertise its own S3 location (if any).

blobs_repository.bzl and blobs_dir_repository.bzl each generate a `:blobs_info`
target alongside `:blobs`, carrying this provider. htvend_lock reads it (via the
`blobs` label passed to htvend_image) so the S3 bucket/prefix used to export blobs
after a lock run matches the bucket/prefix the corresponding htvend_blobs_repository
downloads them from -- one source of truth, no need to repeat `s3_bucket`/`s3_prefix`
on htvend_image itself.
"""

HtvendBlobsInfo = provider(
    doc = "S3 location backing a blobs repository, if any.",
    fields = {
        "s3_bucket": "S3 bucket blobs are stored in, or \"\" for a directory-backed repo.",
        "s3_prefix": "S3 key prefix blobs are stored under.",
    },
)

def _htvend_blobs_info_impl(ctx):
    return [HtvendBlobsInfo(s3_bucket = ctx.attr.s3_bucket, s3_prefix = ctx.attr.s3_prefix)]

htvend_blobs_info = rule(
    implementation = _htvend_blobs_info_impl,
    attrs = {
        "s3_bucket": attr.string(default = ""),
        "s3_prefix": attr.string(default = ""),
    },
)
