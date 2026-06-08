def _htvend_blobs_impl(ctx):
    # Read the assets.json lockfile
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
exports_files(["blobs"])
""")

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