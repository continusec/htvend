"""htvend_lock: (re)generate a lockfile and populate the blob store.

This is a `bazel run` target -- it needs the network. It builds the image online
inside the htvend tool image (recording every fetched asset into the lockfile,
captured into a scratch directory), then copies the updated lockfile back into the
source tree so it can be checked in, and runs `htvend export` to copy the captured
blobs from the scratch directory to their final destination.

The destination is S3 if `s3_bucket` is set (paired with htvend_blobs_s3_repository),
otherwise a local directory `blobs_dir` (paired with htvend_blobs_dir_repository,
defaulting to its directory, or the htvend cache if there isn't one). Either way the
companion htvend_image_build rule then builds offline from the checked-in lockfile +
that blob store.

Most consumers use the combined `htvend_image` macro in defs.bzl, which pairs this
with htvend_image_build.
"""

load(":blobs_info.bzl", "HtvendBlobsInfo")
load(":image.bzl", "DEFAULT_HTVEND_IMAGE", "HOST_PLATFORM_SH", "build_env_flags")

# default local blob directory: the shared htvend cache (shell-expanded at runtime)
_DEFAULT_BLOBS_DIR = "${XDG_DATA_HOME:-$HOME/.local/share}/htvend/cache/blobs"

def _build_run(image, lockfile_name, dockerfile, env_flags):
    """Render the podman invocation that builds the image online, capturing every
    fetched asset into the lockfile and every fetched blob into /blobs (the
    ephemeral scratch directory mounted by the caller)."""
    return """podman run --rm \\
                -v "$tmp_context:/workspace" \\
                -v "$tmp_blobs:/blobs" \\
                -e BUILDAH_OPTS="--isolation=chroot"{env_flags} \\
                --device /dev/fuse \\
                --tmpfs /var/tmp:exec \\
                {image} \\
                   build -m {lockfile_name} --blobs-dir=/blobs -- \\
                       build-img-with-proxy -f {dockerfile} .""".format(
        image = image,
        lockfile_name = lockfile_name,
        dockerfile = dockerfile,
        env_flags = env_flags,
    )

def _export_run(image, lockfile_name, s3_bucket, s3_prefix, blobs_dir):
    """Render the podman invocation that copies the captured blobs from the
    ephemeral scratch directory (/blobs) to their final destination: S3 if
    s3_bucket is set, otherwise the local blobs_dir."""
    if s3_bucket:
        return """podman run --rm \\
                -v "$tmp_context:/workspace" \\
                -v "$HOME/.aws:/root/.aws" \\
                -v "$tmp_blobs:/blobs" \\
                {image} \\
                   export -m {lockfile_name} --blobs-dir=/blobs \\
                    --dest.blobs-backend=s3 \\
                    --dest.blobs-bucket={s3_bucket} \\
                    --dest.blobs-prefix={s3_prefix}""".format(
            image = image,
            lockfile_name = lockfile_name,
            s3_bucket = s3_bucket,
            s3_prefix = s3_prefix,
        )

    return """mkdir -p "{blobs_dir}"
            podman run --rm \\
                -v "$tmp_context:/workspace" \\
                -v "$tmp_blobs:/blobs" \\
                -v "{blobs_dir}:/dest" \\
                {image} \\
                   export -m {lockfile_name} --blobs-dir=/blobs \\
                    --dest.blobs-backend=filesystem \\
                    --dest.blobs-dir=/dest""".format(
        image = image,
        lockfile_name = lockfile_name,
        blobs_dir = blobs_dir,
    )

def _htvend_lock_impl(ctx):
    env_flags = build_env_flags(ctx.attr.env, ctx.attr.platforms)

    # blobs_dir / S3 bucket/prefix: explicit attrs win, otherwise take them from the
    # blobs backend's own :blobs_info (one source of truth with
    # htvend_blobs_dir_repository / htvend_blobs_s3_repository).
    blobs_dir = ctx.attr.blobs_dir
    s3_bucket = ctx.attr.s3_bucket
    s3_prefix = ctx.attr.s3_prefix
    if ctx.attr.blobs_info:
        info = ctx.attr.blobs_info[HtvendBlobsInfo]
        if not blobs_dir:
            blobs_dir = info.blobs_dir
        if not s3_bucket:
            s3_bucket = info.s3_bucket
            s3_prefix = info.s3_prefix
    blobs_dir = blobs_dir or _DEFAULT_BLOBS_DIR

    script = ctx.actions.declare_file(ctx.label.name + "_lock.sh")
    ctx.actions.write(
        output = script,
        content = """#!/bin/bash
            set -euo pipefail

            {host_platform}

            tmp_context=$(mktemp -d)
            tmp_blobs=$(mktemp -d)
            trap 'rm -rf "$tmp_context" "$tmp_blobs"' EXIT

            # copy all files that we need, following symlinks (else they won't work in podman)
            cp -rL "{context_dir}/." "$tmp_context/"

            # build online inside the tool image, recording every asset and capturing
            # blobs into a scratch directory
            {build_run}

            # copy the captured blobs to their final destination
            {export_run}

            # save the lockfile back to our source tree
            cp "$tmp_context/{lockfile_name}" "{package_dir}"
        """.format(
            host_platform = HOST_PLATFORM_SH,
            context_dir = ctx.label.package,
            package_dir = "$BUILD_WORKSPACE_DIRECTORY/" + ctx.label.package,
            lockfile_name = ctx.attr.lockfile_name,
            build_run = _build_run(ctx.attr.image, ctx.attr.lockfile_name, ctx.attr.dockerfile, env_flags),
            export_run = _export_run(ctx.attr.image, ctx.attr.lockfile_name, s3_bucket, s3_prefix, blobs_dir),
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
        # os/arch list captured into the lockfile. Empty (default) -> just the host
        # architecture (see HOST_PLATFORM_SH); set it to capture a multi-arch manifest.
        "platforms": attr.string_list(default = []),
        # local directory blobs are exported to when s3_bucket isn't set. Empty ->
        # taken from blobs_info (if set), i.e. the matching htvend_blobs_dir_repository's
        # own directory; otherwise the shared htvend cache. Set to override either of
        # those.
        "blobs_dir": attr.string(default = ""),
        # S3 bucket/prefix to export blobs to. Empty -> taken from blobs_info (if set),
        # i.e. the matching htvend_blobs_s3_repository's own s3_bucket/s3_prefix. Set
        # these to override that, or to export to S3 with a directory-backed blobs_info.
        "s3_bucket": attr.string(default = ""),
        "s3_prefix": attr.string(default = ""),
        # the `:blobs_info` target generated alongside the blobs backend's `:blobs`
        # (see blobs_info.bzl) -- supplies the default blobs_dir / s3_bucket/s3_prefix
        # above.
        "blobs_info": attr.label(providers = [HtvendBlobsInfo], default = None),
    },
)
