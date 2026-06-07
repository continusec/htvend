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
                oci-archive:{oci_layout} \\
                   build --blobs-dir=/blobs -- make -B

            # export the blobs to s3
            podman run --rm -ti \\
                -v "$tmp_context:/workspace" \\
                -v "$HOME/.aws:/root/.aws" \\
                -v "${{XDG_DATA_HOME:-$HOME/.local/share}}/htvend/cache/blobs":/blobs \\
                oci-archive:{oci_layout} \\
                   export \\
                    --blobs-dir=/blobs \\
                    --dest.blobs-backend=s3 \\
                    --dest.blobs-bucket={s3_bucket} \\
                    --dest.blobs-prefix={s3_prefix}

            # save assets.json back to our source tree
            cp "$tmp_context/assets.json" "{package_dir}"
        """.format(
            oci_layout = ctx.file.image.short_path,
            context_dir = ctx.label.package,
            package_dir = "$BUILD_WORKSPACE_DIRECTORY/" + ctx.label.package,
            s3_bucket = ctx.attr.s3_bucket,
            s3_prefix = ctx.attr.s3_prefix,
        ),
        is_executable = True,
    )

    return [DefaultInfo(
        executable = script,
        runfiles = ctx.runfiles(
            files = [ctx.file.image] + ctx.files.srcs,
        ),
    )]

def htvend_lock(name, srcs = None, **kwargs):
    if srcs == None:
        srcs = native.glob(
            ["**/*"],
            exclude = ["BUILD.bazel", "BUILD", "assets.json"],
        )
    _htvend_lock(
        name = name,
        srcs = srcs,
        **kwargs
    )

_htvend_lock = rule(
    implementation = _htvend_lock_impl,
    executable = True,
    attrs = {
        "srcs": attr.label_list(allow_files = True, default = []),
        "image": attr.label(
            default = Label("//offline-img-builder:oci_tarball"),
            allow_single_file = True,
        ),
        "s3_bucket": attr.string(mandatory = True),
        "s3_prefix": attr.string(mandatory = True),
    },
)
