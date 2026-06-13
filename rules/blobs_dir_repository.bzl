"""htvend_blobs_dir_repository: expose a local directory of blobs to Bazel.

A directory-backed alternative to blobs_repository (which fetches from S3). Use
this when your blobs live on a local/shared filesystem -- e.g. the htvend cache
at ${XDG_DATA_HOME}/htvend/cache/blobs, an NFS mount, or a checked-out artifact
dir -- and you don't want S3 or credentials in the loop.

The directory may be shared by many images (it's content-addressed by sha256), so
this rule only exposes the blobs listed in `assets_json` -- the same lockfile the
matching htvend_image reads -- rather than the whole directory.

It produces the same `@<name>//:blobs` target as the S3 variant, so it's a drop-in
for the `blobs` attribute of htvend_image.
"""

def _htvend_blobs_dir_impl(ctx):
    blobs_dir = ctx.os.environ.get("HTVEND_BLOBS_DIR", ctx.attr.blobs_dir)
    if not blobs_dir:
        xdg = ctx.os.environ.get("XDG_DATA_HOME")
        if not xdg:
            xdg = ctx.os.environ.get("HOME", "") + "/.local/share"
        blobs_dir = xdg + "/htvend/cache/blobs"

    # Only expose the blobs this image's lockfile actually references, not the
    # whole (possibly shared) directory.
    assets = json.decode(ctx.read(ctx.attr.assets_json))
    for blob in assets.values():
        sha256 = blob["Sha256"]
        ctx.symlink(blobs_dir + "/" + sha256, "blobs/" + sha256)

    ctx.file("BUILD.bazel", """
load("@rules_htvend//:blobs_info.bzl", "htvend_blobs_info")

exports_files(["blobs"])

# Directory-backed: no S3 location to advertise.
htvend_blobs_info(name = "blobs_info", visibility = ["//visibility:public"])
""")

htvend_blobs_dir_repository = repository_rule(
    implementation = _htvend_blobs_dir_impl,
    attrs = {
        "assets_json": attr.label(
            mandatory = True,
            allow_single_file = True,
        ),
        # Path to the blobs directory. If empty, falls back to the HTVEND_BLOBS_DIR
        # env var, then to ${XDG_DATA_HOME:-$HOME/.local/share}/htvend/cache/blobs.
        "blobs_dir": attr.string(default = ""),
    },
    environ = ["HTVEND_BLOBS_DIR", "XDG_DATA_HOME", "HOME"],
    local = True,
)
