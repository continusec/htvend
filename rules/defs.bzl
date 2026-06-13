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

load(":image.bzl", "htvend_image_build")
load(":lock.bzl", "htvend_lock")

def htvend_image(
        name,
        blobs,
        dockerfile = "Dockerfile",
        env = {},
        srcs = None,
        lockfile_name = "assets.json",
        image = None,
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
      srcs: build context files. Defaults to a glob of the package.
      lockfile_name: the lockfile to read/write. Default "assets.json".
      image: override the htvend tool image (defaults to the pinned published image).
      blobs_dir: lock only -- local directory to write blobs into. Empty -> htvend cache.
      s3_bucket: lock only -- if set, also export blobs to this S3 bucket.
      s3_prefix: lock only -- S3 key prefix.
      lock_name: name of the lock target. Default "<name>.lock".
      **kwargs: common args (e.g. visibility) applied to both targets.
    """
    lock_name = lock_name or (name + ".lock")

    # shared image override only passed through when set, so each rule keeps its default
    image_kwargs = {"image": image} if image else {}

    htvend_lock(
        name = lock_name,
        dockerfile = dockerfile,
        env = env,
        srcs = srcs,
        lockfile_name = lockfile_name,
        blobs_dir = blobs_dir,
        s3_bucket = s3_bucket,
        s3_prefix = s3_prefix,
        **dict(image_kwargs, **kwargs)
    )

    htvend_image_build(
        name = name,
        blobs = blobs,
        dockerfile = dockerfile,
        env = env,
        srcs = srcs,
        lockfile_name = lockfile_name,
        **dict(image_kwargs, **kwargs)
    )
