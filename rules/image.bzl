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
DEFAULT_HTVEND_IMAGE = "ghcr.io/continusec/htvend:2.2@sha256:c8a817e67e119693c1f583b6f867e2c3a1a9019760425e1821c49ec077f4f611"

# Shell snippet (prepended to the generated build/lock scripts) that computes the
# host os/arch in buildah's `os/arch` form, e.g. linux/arm64. Used as the default
# single-platform target when `platforms` is unset -- so a fresh project builds for
# the machine it's run on, with no multi-arch / qemu setup required.
HOST_PLATFORM_SH = 'host_platform="linux/$(uname -m | sed -e s/x86_64/amd64/ -e s/aarch64/arm64/)"'

def _platforms_value(platforms):
    """The value for the PLATFORMS env var: the explicit os/arch list if given,
    otherwise a reference to the `$host_platform` shell var defined by
    HOST_PLATFORM_SH (i.e. build for the current architecture)."""
    if platforms:
        return " ".join(platforms)
    return "$host_platform"

def render_env_flags(env):
    """Render an env dict as `-e "K=V"` podman flags."""
    flags = ""
    for k, v in env.items():
        flags += ' -e "{}={}"'.format(k, v)
    return flags

def build_env_flags(env, platforms):
    """Render the env dict plus the PLATFORMS list as podman `-e` flags.

    build-img-with-proxy reads PLATFORMS (space separated os/arch) to decide which
    architectures to build into the manifest. An empty list defaults to the host
    architecture via the `$host_platform` shell var (see HOST_PLATFORM_SH).
    """
    merged = dict(env)
    merged["PLATFORMS"] = _platforms_value(platforms)
    return render_env_flags(merged)

def build_env_exports(env, platforms):
    """Render the env dict plus PLATFORMS as `K="V"` pairs for the `env` command.

    The direct (non-podman) execution path runs the tools straight from PATH, so the
    same variables podman would inject via `-e` are passed through `env K=V ...`.
    """
    merged = dict(env)
    merged["PLATFORMS"] = _platforms_value(platforms)
    parts = []
    for k, v in merged.items():
        parts.append('{}="{}"'.format(k, v))
    return " ".join(parts)

# Build setting flag: `--@rules_htvend//:exec_mode={podman,direct}` selects how the
# offline build runs. "podman" (default) runs the tool image via podman and is
# local-only. "direct" runs htvend/buildah/build-img-with-proxy straight from PATH
# (e.g. an RBE worker whose image IS the tool image), with all inputs declared, so
# the action is sandbox- and remote-eligible.
ExecModeInfo = provider(
    doc = "Selected htvend offline-build execution mode.",
    fields = {"value": "string: 'podman' or 'direct'"},
)

def _exec_mode_flag_impl(ctx):
    return ExecModeInfo(value = ctx.build_setting_value)

exec_mode_flag = rule(
    implementation = _exec_mode_flag_impl,
    build_setting = config.string(flag = True),
)

def _htvend_image_impl(ctx):
    mode = ctx.attr.exec_mode or ctx.attr._exec_mode_flag[ExecModeInfo].value

    output_oci_layout = ctx.actions.declare_directory(ctx.label.name + ".oci")
    script = ctx.actions.declare_file(ctx.label.name + "_offline.sh")
    blobs_dir = ctx.files.blobs[0].dirname

    if mode == "podman":
        # Deliver the tooling via the published tool image and run it under podman.
        # podman needs the real host (devices, $HOME storage, its own namespaces), so
        # this path stays local + unsandboxed. --network=none gives the container no
        # external network at all (buildah's inner --network=host still shares this
        # netns, so loopback to the htvend proxy keeps working) -- so a plain
        # `bazel build :image` is itself the offline/hermeticity test, no Bazel
        # sandbox or host buildah install required.
        #
        # storage_driver is normally left empty here: the tool image runs buildah with
        # overlay + the --device /dev/fuse passed below, which is the efficient local
        # path. Honour an explicit override if the consumer sets one.
        podman_env = dict(ctx.attr.env)
        if ctx.attr.storage_driver:
            podman_env["STORAGE_DRIVER"] = ctx.attr.storage_driver

        run_block = """# run podman, mounting our temp context
            PATH=/usr/local/bin:$PATH podman run --rm \\
                -v "$tmp_context:/workspace" \\
                -e BUILDAH_OPTS="--isolation=chroot"{env_flags} \\
                --device /dev/fuse \\
                --tmpfs /var/tmp:exec \\
                --network=none \\
                {image} \\
                   offline -m {lockfile_name} --blobs-dir=/workspace/blobs -- \\
                       build-img-with-proxy -f {dockerfile} .""".format(
            image = ctx.attr.image,
            lockfile_name = ctx.attr.lockfile_name,
            dockerfile = ctx.attr.dockerfile,
            env_flags = build_env_flags(podman_env, ctx.attr.platforms),
        )
        execution_requirements = {
            "no-sandbox": "1",
            "local": "1",
        }
    else:
        # Run htvend/buildah/build-img-with-proxy straight from PATH (the exec
        # environment provides them -- e.g. an RBE worker whose image is the tool
        # image). All inputs are declared and there is no network, so this path is
        # sandbox- and remote-eligible. Worker selection is left to the consumer's
        # exec_properties + platform.
        #
        # Default the storage driver to vfs: unlike overlay it needs no /dev/fuse, so the
        # action runs on an RBE worker with no fuse device. The output image is identical
        # (vfs just copies where overlay would mount). Override with the storage_driver attr.
        direct_env = dict(ctx.attr.env)
        direct_env["STORAGE_DRIVER"] = ctx.attr.storage_driver or "vfs"

        run_block = """# run the tools directly from PATH (no podman); subshell keeps the
            # cd local so the final cp below still resolves the relative output path
            ( cd "$tmp_context" && \\
              env BUILDAH_ISOLATION=chroot BUILDAH_OPTS="--isolation=chroot" {env_exports} \\
                  htvend offline -m {lockfile_name} --blobs-dir="$tmp_context/blobs" -- \\
                      build-img-with-proxy -f {dockerfile} . )""".format(
            lockfile_name = ctx.attr.lockfile_name,
            dockerfile = ctx.attr.dockerfile,
            env_exports = build_env_exports(direct_env, ctx.attr.platforms),
        )
        execution_requirements = {}

    ctx.actions.write(
        output = script,
        content = """#!/bin/bash
            set -euo pipefail

            {host_platform}

            tmp_context=$(mktemp -d)
            trap 'rm -rf "$tmp_context"' EXIT

            # copy all files that we need, following symlinks (else they won't work in
            # podman, and we need a writable tree for the build output)
            cp -rL "{context_dir}/." "{blobs_dir}/blobs" "$tmp_context/"

            {run_block}

            cp -R $tmp_context/oci/* "{output_oci_layout}"
        """.format(
            host_platform = HOST_PLATFORM_SH,
            context_dir = ctx.label.package,
            blobs_dir = blobs_dir,
            run_block = run_block,
            output_oci_layout = output_oci_layout.path,
        ),
        is_executable = True,
    )

    ctx.actions.run(
        executable = script,
        inputs = ctx.files.srcs + ctx.files.blobs,
        outputs = [output_oci_layout],
        mnemonic = "HtvendOffline",
        execution_requirements = execution_requirements,
        # Inherit a real (exported) PATH so the tools resolve: in direct mode htvend
        # execs build-img-with-proxy via Go's LookPath against its process env, and in
        # podman mode the script needs `podman` on PATH. Without this the action env has
        # no exported PATH and child processes can't find anything. Tune with
        # --action_env=PATH=... ; on RBE the worker (= tool image) supplies the default.
        use_default_shell_env = True,
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
        # buildah storage driver for the offline build (sets STORAGE_DRIVER). Empty ->
        # per-mode default: direct/RBE mode uses "vfs" (needs no /dev/fuse on the worker);
        # podman mode leaves it to the tool image (overlay + the --device /dev/fuse it
        # passes). Set to "overlay" or "vfs" to override either mode.
        "storage_driver": attr.string(default = ""),
        # os/arch list built into the manifest. Empty (default) -> just the host
        # architecture (see HOST_PLATFORM_SH); set it to build a multi-arch manifest.
        "platforms": attr.string_list(default = []),
        # "" -> follow the --@rules_htvend//:exec_mode flag; otherwise override per target.
        "exec_mode": attr.string(default = "", values = ["", "podman", "direct"]),
        "_exec_mode_flag": attr.label(default = Label("//:exec_mode")),
        "blobs": attr.label(
            mandatory = True,
            allow_files = True,
        ),
    },
)
