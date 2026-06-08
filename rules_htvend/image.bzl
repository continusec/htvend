def _htvend_image_impl(ctx):
    output_oci_layout = ctx.actions.declare_directory(ctx.label.name + ".oci")
    script = ctx.actions.declare_file(ctx.label.name + "_offline.sh")
    blobs_dir = ctx.files.blobs[0].dirname
    ctx.actions.write(
        output = script,
        content = """#!/bin/bash
            set -euo pipefail

            tmp_context=$(mktemp -d)
            trap 'rm -rf "$tmp_context"' EXIT

            # copy all files that we need, following symlinks (else they won't work in podman)
            cp -rL "{context_dir}/." "{blobs_dir}/blobs" "$tmp_context/"

            # run podman, mounting our temp context
            PATH=/usr/local/bin:$PATH podman run --rm \\
                -v "$tmp_context:/workspace" \\
                -e BUILDAH_OPTS="--isolation=chroot" \\
                --device /dev/fuse \\
                --tmpfs /var/tmp:exec \\
                oci-archive:{oci_layout} \\
                   offline --blobs-dir=/workspace/blobs -- make -B
            cp -R $tmp_context/oci/* "{output_oci_layout}"
        """.format(
            oci_layout = ctx.file.image.path,
            context_dir = ctx.label.package,
            output_oci_layout = output_oci_layout.path,
            blobs_dir = blobs_dir,
        ),
        is_executable = True,
    )

    ctx.actions.run(
        executable = script,
        inputs = [
            ctx.file.image,
        ] + ctx.files.srcs + ctx.files.blobs,
        outputs = [output_oci_layout],
        mnemonic = "HtvendOffline",
    )

    return [DefaultInfo(files = depset([output_oci_layout]))]

def htvend_image(name, srcs = None, **kwargs):
    if srcs == None:
        srcs = native.glob(
            ["**/*"],
            exclude = ["BUILD.bazel", "BUILD"],
        )
    _htvend_image(
        name = name,
        srcs = srcs,
        **kwargs
    )

_htvend_image = rule(
    implementation = _htvend_image_impl,
    attrs = {
        "srcs": attr.label_list(allow_files = True, default = []),
        "image": attr.label(
            default = Label("//offline-img-builder:oci_tarball"),
            allow_single_file = True,
        ),
        "blobs": attr.label(
            mandatory = True,
            allow_files = True,
        ),
    },
)
