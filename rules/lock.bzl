"""htvend_lock: (re)generate a lockfile and populate the blob store.

This is a `bazel run` target -- it needs the network. It builds the image online
inside the htvend tool image (recording every fetched asset into the lockfile),
pushes the blobs to S3, then copies the updated lockfile back into the source tree
so it can be checked in. The companion htvend_image rule then builds offline from
that lockfile + blobs.
"""

load(":image.bzl", "DEFAULT_HTVEND_IMAGE")

def _htvend_lock_impl(ctx):
    script = ctx.actions.declare_file(ctx.label.name + "_lock.sh")
    ctx.actions.write(
        output = script,
        content = """#!/bin/bash
            set -euo pipefail

            tmp_context=$(mktemp -d)
            trap 'rm -rf "$tmp_context"' EXIT

            # copy all files that we need, following symlinks (else they won't work in podman)
            cp -rL "{context_dir}/." "$tmp_context/"

            # run podman, mounting our temp context
            # we also pass through our local blob cache to speed it up
            podman run --rm -ti \\
                -v "$tmp_context:/workspace" \\
                -v "${{XDG_DATA_HOME:-$HOME/.local/share}}/htvend/cache/blobs":/blobs \\
                -e BUILDAH_OPTS="--isolation=chroot" \\
                --device /dev/fuse \\
                --tmpfs /var/tmp:exec \\
                {image} \\
                   build -m {lockfile_name} --blobs-dir=/blobs -- make -B

            # export the blobs to s3
            podman run --rm -ti \\
                -v "$tmp_context:/workspace" \\
                -v "$HOME/.aws:/root/.aws" \\
                -v "${{XDG_DATA_HOME:-$HOME/.local/share}}/htvend/cache/blobs":/blobs \\
                {image} \\
                   export \\
                    -m {lockfile_name} \\
                    --blobs-dir=/blobs \\
                    --dest.blobs-backend=s3 \\
                    --dest.blobs-bucket={s3_bucket} \\
                    --dest.blobs-prefix={s3_prefix}

            # save the lockfile back to our source tree
            cp "$tmp_context/{lockfile_name}" "{package_dir}"
        """.format(
            image = ctx.attr.image,
            context_dir = ctx.label.package,
            package_dir = "$BUILD_WORKSPACE_DIRECTORY/" + ctx.label.package,
            s3_bucket = ctx.attr.s3_bucket,
            s3_prefix = ctx.attr.s3_prefix,
            lockfile_name = ctx.attr.lockfile_name,
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
        "s3_bucket": attr.string(mandatory = True),
        "s3_prefix": attr.string(mandatory = True),
    },
)
