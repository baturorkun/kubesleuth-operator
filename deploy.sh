#!/usr/bin/env bash

export KIND_EXPERIMENTAL_PROVIDER=podman

# Choose container engine (docker preferred, fall back to podman)
if command -v docker >/dev/null 2>&1; then
  CONTAINER_ENGINE=docker
elif command -v podman >/dev/null 2>&1; then
  CONTAINER_ENGINE=podman
else
  echo "Error: neither 'docker' nor 'podman' found in PATH. Install one or adjust your environment." >&2
  exit 1
fi
export CONTAINER_ENGINE

IMG="${IMG:-my-controller:latest}"
echo "Using $CONTAINER_ENGINE to build image $IMG"

# Build image with the selected engine
$CONTAINER_ENGINE build -t "$IMG" .

# 3) load image to cluster
# Use different strategy when using podman (kind's docker-image loader may not see podman images)
if [ "$CONTAINER_ENGINE" = "podman" ]; then
  tmpimg="$(mktemp /tmp/image-XXXXXX.tar)"
  echo "Saving podman image to $tmpimg"
  podman save -o "$tmpimg" "$IMG"
  if [ -n "${KIND_CLUSTER:-}" ]; then
    echo "Loading image into kind cluster ${KIND_CLUSTER} via image-archive"
    kind load image-archive --name "${KIND_CLUSTER}" "$tmpimg"
  else
    echo "Loading image into default kind cluster via image-archive"
    kind load image-archive "$tmpimg"
  fi
  rm -f "$tmpimg"
else
  kind load docker-image "$IMG"
fi

# 4) Deploy
make deploy IMG=$IMG
