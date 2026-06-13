"""htvend_blobs_dir_repository: expose a local directory of blobs to Bazel.

A directory-backed alternative to blobs_repository (which fetches from S3). Use
this when your blobs live on a local/shared filesystem -- e.g. the htvend cache
at ${XDG_DATA_HOME}/htvend/cache/blobs, an NFS mount, or a checked-out artifact
dir -- and you don't want S3 or credentials in the loop.

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

    ctx.symlink(blobs_dir, "blobs")
    ctx.file("BUILD.bazel", 'exports_files(["blobs"])\n')

htvend_blobs_dir_repository = repository_rule(
    implementation = _htvend_blobs_dir_impl,
    attrs = {
        # Path to the blobs directory. If empty, falls back to the HTVEND_BLOBS_DIR
        # env var, then to ${XDG_DATA_HOME:-$HOME/.local/share}/htvend/cache/blobs.
        "blobs_dir": attr.string(default = ""),
    },
    environ = ["HTVEND_BLOBS_DIR", "XDG_DATA_HOME", "HOME"],
    local = True,
)
