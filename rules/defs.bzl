"""Public API for rules_htvend.

`htvend_image` is the one entry point most consumers need: a single call that creates
the pair of targets every image needs, sharing their common configuration:

    htvend_image(name = "image", blobs = "@my_blobs//:blobs")

produces:

    //pkg:image        (bazel build) -- the OCI image, built offline from the lockfile
    //pkg:image.lock   (bazel run)   -- regenerate the lockfile + populate the blobs

Typical flow:

    bazel run   //pkg:image.lock   # online, on demand: capture assets, write lockfile
    bazel build //pkg:image        # offline: build the image from the checked-in lockfile

For advanced cases where you want the two targets configured independently, load
htvend_image_build from image.bzl and htvend_lock from lock.bzl directly.
"""

load(":image.bzl", "DEFAULT_HTVEND_IMAGE", "htvend_image_build")
load(":lock.bzl", "htvend_lock")

def htvend_image(
        name,
        blobs,
        dockerfile = "Dockerfile",
        env = {},
        platforms = [],
        srcs = None,
        lockfile_name = "assets.json",
        image = None,
        exec_mode = "",
        blobs_dir = "",
        s3_bucket = "",
        s3_prefix = "",
        lock_name = None,
        **kwargs):
    """Create the offline build target `name` and the lock run target `name.lock`.

    Args:
      name: name of the offline OCI build target.
      blobs: the blob set to build from (e.g. "@my_blobs//:blobs").
      dockerfile: Dockerfile to build (relative to the package). Default "Dockerfile".
      env: dict of environment variables to pass into the build (e.g. settings paths).
      platforms: os/arch list to build into the manifest. Empty (default) builds for
        the host architecture only; set it (e.g. ["linux/amd64", "linux/arm64"]) for a
        multi-arch manifest.
      srcs: build context files. Defaults to a glob of the package.
      lockfile_name: the lockfile to read/write. Default "assets.json".
      image: override the htvend tool image (defaults to the pinned published image).
      exec_mode: build only -- "podman" (local default) or "direct" (RBE-eligible, runs
        tools from PATH). Empty -> follow the --@rules_htvend//:exec_mode flag.
      blobs_dir: lock only -- local directory to write blobs into. Empty -> taken from
        `blobs`'s own :blobs_info (i.e. its htvend_blobs_dir_repository directory), or
        the htvend cache if there isn't one.
      s3_bucket: lock only -- override the S3 bucket blobs are exported to. Empty ->
        taken from `blobs`'s own :blobs_info (i.e. its htvend_blobs_s3_repository
        config), or no S3 export for a directory-backed `blobs`.
      s3_prefix: lock only -- override the S3 key prefix. Empty -> see s3_bucket.
      lock_name: name of the lock target. Default "<name>.lock".
      **kwargs: common args (e.g. visibility) applied to both targets.
    """
    lock_name = lock_name or (name + ".lock")

    # shared image override only passed through when set, so each rule keeps its default
    image_kwargs = {"image": image} if image else {}

    # The blobs backend (htvend_blobs_s3_repository / htvend_blobs_dir_repository)
    # generates a `:blobs_info` target alongside `:blobs` advertising its own S3
    # bucket/prefix and/or local blobs_dir (if any). Pass it through so htvend_lock
    # can default blobs_dir/s3_bucket/s3_prefix from it -- one source of truth, no
    # need to repeat them here. String manipulation (not Label()) so the result is
    # resolved in the caller's repo mapping, not rules_htvend's.
    blobs_info = blobs.rsplit(":", 1)[0] + ":blobs_info"

    htvend_lock(
        name = lock_name,
        dockerfile = dockerfile,
        env = env,
        platforms = platforms,
        srcs = srcs,
        lockfile_name = lockfile_name,
        blobs_dir = blobs_dir,
        s3_bucket = s3_bucket,
        s3_prefix = s3_prefix,
        blobs_info = blobs_info,
        **dict(image_kwargs, **kwargs)
    )

    # Default the RBE worker selection to the same tool image used for podman, so the
    # image version has one source of truth (the `image` attr / DEFAULT_HTVEND_IMAGE).
    # Only takes effect under --@rules_htvend//:exec_mode=direct; harmless otherwise.
    # Consumers needing a different RBE worker can pass their own exec_properties.
    build_kwargs = dict(image_kwargs, **kwargs)
    if "exec_properties" not in build_kwargs:
        build_kwargs["exec_properties"] = {
            "container-image": "docker://" + (image or DEFAULT_HTVEND_IMAGE),
            "OSFamily": "linux",
        }

    htvend_image_build(
        name = name,
        blobs = blobs,
        dockerfile = dockerfile,
        env = env,
        platforms = platforms,
        exec_mode = exec_mode,
        srcs = srcs,
        lockfile_name = lockfile_name,
        **build_kwargs
    )
