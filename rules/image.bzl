"""htvend_image_build: build an OCI image offline from a checked-in lockfile + blobs.

The rule mounts the build context and the blob set into the htvend tool image and
runs `htvend offline ... -- build-img-with-proxy`, which invokes buildah with no
network access. The output is an OCI layout directory that downstream rules (e.g.
rules_oci/rules_img push) can consume.

Most consumers use the combined `htvend_image` macro in defs.bzl, which pairs this
with htvend_lock.
"""

# Default published tool image. podman resolves this from the local image store if
# present (e.g. after `cd cli && make image IMAGE_TAG=...`), otherwise pulls it.
# Pin by digest for fully reproducible builds.
DEFAULT_HTVEND_IMAGE = "ghcr.io/continusec/htvend:2.1@sha256:ebb2c06cfc40ed6dbfa9d203127b5cd0cd535f8d07c3b4b7ec07ea35f3f184e4 "

def render_env_flags(env):
    """Render an env dict as `-e "K=V"` podman flags."""
    flags = ""
    for k, v in env.items():
        flags += ' -e "{}={}"'.format(k, v)
    return flags

def build_env_flags(env, platforms):
    """Render the env dict plus the PLATFORMS list as podman `-e` flags.

    build-img-with-proxy reads PLATFORMS (space separated os/arch) to decide which
    architectures to build into the manifest; an empty list falls back to the
    script's own default.
    """
    merged = dict(env)
    if platforms:
        merged["PLATFORMS"] = " ".join(platforms)
    return render_env_flags(merged)

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
                -e BUILDAH_OPTS="--isolation=chroot"{env_flags} \\
                --device /dev/fuse \\
                --tmpfs /var/tmp:exec \\
                {image} \\
                   offline -m {lockfile_name} --blobs-dir=/workspace/blobs -- \\
                       build-img-with-proxy -f {dockerfile} .
            cp -R $tmp_context/oci/* "{output_oci_layout}"
        """.format(
            image = ctx.attr.image,
            context_dir = ctx.label.package,
            output_oci_layout = output_oci_layout.path,
            blobs_dir = blobs_dir,
            lockfile_name = ctx.attr.lockfile_name,
            dockerfile = ctx.attr.dockerfile,
            env_flags = build_env_flags(ctx.attr.env, ctx.attr.platforms),
        ),
        is_executable = True,
    )

    ctx.actions.run(
        executable = script,
        inputs = ctx.files.srcs + ctx.files.blobs,
        outputs = [output_oci_layout],
        mnemonic = "HtvendOffline",
        execution_requirements = {
            "no-sandbox": "1",
            "local": "1",
        },
    )

    return [DefaultInfo(files = depset([output_oci_layout]))]

def htvend_image_build(name, srcs = None, lockfile_name = "assets.json", **kwargs):
    if srcs == None:
        srcs = native.glob(
            ["**/*"],
            exclude = ["BUILD.bazel", "BUILD"],
        )
    _htvend_image(
        name = name,
        srcs = srcs,
        lockfile_name = lockfile_name,
        **kwargs
    )

_htvend_image = rule(
    implementation = _htvend_image_impl,
    attrs = {
        "srcs": attr.label_list(allow_files = True, default = []),
        "image": attr.string(default = DEFAULT_HTVEND_IMAGE),
        "lockfile_name": attr.string(default = "assets.json"),
        "dockerfile": attr.string(default = "Dockerfile"),
        "env": attr.string_dict(default = {}),
        "platforms": attr.string_list(default = ["linux/amd64", "linux/arm64"]),
        "blobs": attr.label(
            mandatory = True,
            allow_files = True,
        ),
    },
)
