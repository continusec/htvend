#!/usr/bin/env bash
#
# End-to-end test for the Bazel rules, driving the examples/alpine-img image: lock it,
# build it, then clean + rebuild and assert the image digest is unchanged (i.e. the
# offline build is reproducible). This is the executable version of the walkthrough in
# docs/bazel.md.
#
# Runs against a throwaway blob store (HTVEND_BLOBS_DIR) and removes the generated
# lockfile afterwards, so it leaves the working tree as it found it.
#
# Needs bazel + podman, so run it inside the build VM, e.g.:
#   limactl shell u2404 /path/to/httpvendor/e2e-bazel-test.sh
set -euo pipefail

# This script lives at the repo root; the example workspace is in ./examples.
SCRIPT_DIR=$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)
EXAMPLES_DIR="$SCRIPT_DIR/examples"

# Throwaway blob store: the directory backend reads from HTVEND_BLOBS_DIR, and the lock
# run exports the captured blobs back into the same place. Using a temp dir keeps the
# test hermetic (no dependency on / pollution of the shared htvend cache).
export HTVEND_BLOBS_DIR=$(mktemp -d)

# Clean up the blob store and the lockfile the lock run writes into the source tree.
trap 'rm -rf "$HTVEND_BLOBS_DIR" "$EXAMPLES_DIR/alpine-img/assets.json"' EXIT

cd "$EXAMPLES_DIR"

echo "==> blob store: $HTVEND_BLOBS_DIR"

# The image's manifest digest, read from the OCI layout index. Stable iff the offline
# build is reproducible.
manifest_digest() {
    jq -r '.manifests[0].digest' bazel-bin/alpine-img/image.oci/index.json
}

echo "==> bazel run //alpine-img:image.lock (online: capture assets + populate blob store)"
bazel run //alpine-img:image.lock

echo "==> bazel build //alpine-img:image (first build)"
bazel build //alpine-img:image
DIGEST1=$(manifest_digest)
echo "    digest: $DIGEST1"

echo "==> bazel clean"
bazel clean

echo "==> bazel build //alpine-img:image (rebuild after clean)"
bazel build //alpine-img:image
DIGEST2=$(manifest_digest)
echo "    digest: $DIGEST2"

echo
if [ "$DIGEST1" = "$DIGEST2" ]; then
    echo "PASS: image digest unchanged across rebuilds ($DIGEST1)"
else
    echo "FAIL: image digest changed"
    echo "  first:  $DIGEST1"
    echo "  second: $DIGEST2"
    exit 1
fi
