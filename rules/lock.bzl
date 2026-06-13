"""htvend_lock: (re)generate a lockfile and populate the blob store.

This is a `bazel run` target -- it needs the network. It builds the image online
inside the htvend tool image (recording every fetched asset into the lockfile), then
copies the updated lockfile back into the source tree so it can be checked in. The
companion htvend_image_build rule then builds offline from that lockfile + blobs.

Blobs are always written to a local directory (`blobs_dir`, defaulting to the htvend
cache). If `s3_bucket` is set, they are additionally exported to S3 so other machines
can fetch them via htvend_blobs_repository. With no `s3_bucket`, the flow is fully
local and credential-free -- pair it with htvend_blobs_dir_repository.

Most consumers use the combined `htvend_image` macro in defs.bzl, which pairs this
with htvend_image_build.
"""

load(":image.bzl", "DEFAULT_HTVEND_IMAGE", "render_env_flags")

# default local blob directory: the shared htvend cache (shell-expanded at runtime)
_DEFAULT_BLOBS_DIR = "${XDG_DATA_HOME:-$HOME/.local/share}/htvend/cache/blobs"

def _htvend_lock_impl(ctx):
    blobs_dir = ctx.attr.blobs_dir or _DEFAULT_BLOBS_DIR
    env_flags = render_env_flags(ctx.attr.env)

    # optional: also push the blobs up to S3
    s3_block = ""
    if ctx.attr.s3_bucket:
        s3_block = """
            # export the blobs to s3
            podman run --rm \\
                -v "$tmp_context:/workspace" \\
                -v "$HOME/.aws:/root/.aws" \\
                -v "{blobs_dir}":/blobs \\
                {image} \\
                   export \\
                    -m {lockfile_name} \\
                    --blobs-dir=/blobs \\
                    --dest.blobs-backend=s3 \\
                    --dest.blobs-bucket={s3_bucket} \\
                    --dest.blobs-prefix={s3_prefix}
""".format(
            image = ctx.attr.image,
            blobs_dir = blobs_dir,
            lockfile_name = ctx.attr.lockfile_name,
            s3_bucket = ctx.attr.s3_bucket,
            s3_prefix = ctx.attr.s3_prefix,
        )

    script = ctx.actions.declare_file(ctx.label.name + "_lock.sh")
    ctx.actions.write(
        output = script,
        content = """#!/bin/bash
            set -euo pipefail

            tmp_context=$(mktemp -d)
            trap 'rm -rf "$tmp_context"' EXIT

            # copy all files that we need, following symlinks (else they won't work in podman)
            cp -rL "{context_dir}/." "$tmp_context/"

            # ensure the local blobs directory exists before we mount it
            mkdir -p "{blobs_dir}"

            # build online inside the tool image, recording every asset and storing
            # blobs into our local blobs directory
            podman run --rm \\
                -v "$tmp_context:/workspace" \\
                -v "{blobs_dir}":/blobs \\
                -e BUILDAH_OPTS="--isolation=chroot"{env_flags} \\
                --device /dev/fuse \\
                --tmpfs /var/tmp:exec \\
                {image} \\
                   build -m {lockfile_name} --blobs-dir=/blobs -- \\
                       build-img-with-proxy -f {dockerfile} .
{s3_block}
            # save the lockfile back to our source tree
            cp "$tmp_context/{lockfile_name}" "{package_dir}"
        """.format(
            image = ctx.attr.image,
            context_dir = ctx.label.package,
            package_dir = "$BUILD_WORKSPACE_DIRECTORY/" + ctx.label.package,
            blobs_dir = blobs_dir,
            lockfile_name = ctx.attr.lockfile_name,
            dockerfile = ctx.attr.dockerfile,
            env_flags = env_flags,
            s3_block = s3_block,
        ),
        is_executable = True,
    )

    return [DefaultInfo(
        executable = script,
        runfiles = ctx.runfiles(
            files = ctx.files.srcs,
        ),
    )]

def htvend_lock(name, srcs = None, lockfile_name = "assets.json", **kwargs):
    if srcs == None:
        srcs = native.glob(
            ["**/*"],
            exclude = ["BUILD.bazel", "BUILD", lockfile_name],
        )
    _htvend_lock(
        name = name,
        srcs = srcs,
        lockfile_name = lockfile_name,
        **kwargs
    )

_htvend_lock = rule(
    implementation = _htvend_lock_impl,
    executable = True,
    attrs = {
        "srcs": attr.label_list(allow_files = True, default = []),
        "image": attr.string(default = DEFAULT_HTVEND_IMAGE),
        "lockfile_name": attr.string(default = "assets.json"),
        "dockerfile": attr.string(default = "Dockerfile"),
        "env": attr.string_dict(default = {}),
        # local directory to store blobs in. Empty -> the shared htvend cache.
        # Should match the directory the matching htvend_blobs_dir_repository reads.
        "blobs_dir": attr.string(default = ""),
        # if set, also export blobs to this S3 bucket (for htvend_blobs_repository).
        "s3_bucket": attr.string(default = ""),
        "s3_prefix": attr.string(default = ""),
    },
)
