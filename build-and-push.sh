#!/bin/bash

# Container Build and Push Script
# Supports both Docker and Podman - automatically detects which one you have
#
# Usage:
#   ./build-and-push.sh [USERNAME] [VERSION]
#   ./build-and-push.sh baturorkun v1.0.0
#   ./build-and-push.sh baturorkun
#   ./build-and-push.sh

set -e

echo "🐳 Kubebuilder Demo Operator - Container Build & Push"
echo "======================================================"
echo ""

# Detect container runtime (Podman or Docker)
if command -v podman &> /dev/null; then
    export CONTAINER_TOOL="podman"
    export CONTAINER_ENGINE="podman"
    echo "✅ Detected: Podman"
elif command -v docker &> /dev/null; then
    export CONTAINER_TOOL="docker"
    export CONTAINER_ENGINE="docker"
    echo "✅ Detected: Docker"
else
    echo "❌ Error: Neither Docker nor Podman is installed."
    echo "   Please install one of them:"
    echo "   - Docker: https://docs.docker.com/get-docker/"
    echo "   - Podman: https://podman.io/getting-started/installation"
    exit 1
fi

# Check if container runtime is running
if ! $CONTAINER_TOOL info > /dev/null 2>&1; then
    if [ "$CONTAINER_TOOL" = "docker" ]; then
        echo "❌ Error: Docker is not running. Please start Docker Desktop."
    else
        echo "❌ Error: Podman is not accessible."
    fi
    exit 1
fi

echo ""

# Get username from parameter or prompt
REGISTRY_USERNAME="${1}"
if [ -z "$REGISTRY_USERNAME" ]; then
    echo "📝 Enter your registry username (e.g., baturorkun):"
    read -r REGISTRY_USERNAME
fi

if [ -z "$REGISTRY_USERNAME" ]; then
    echo "❌ Error: Username cannot be empty"
    exit 1
fi

# Get version from parameter or use default
IMAGE_NAME="kubebuilder-demo-operator"
IMAGE_TAG="${2:-v1.0.0}"
FULL_IMAGE="${REGISTRY_USERNAME}/${IMAGE_NAME}:${IMAGE_TAG}"
LATEST_IMAGE="${REGISTRY_USERNAME}/${IMAGE_NAME}:latest"

echo ""
echo "📦 Image details:"
echo "   Tool: ${CONTAINER_TOOL}"
echo "   Registry: ${REGISTRY_USERNAME}"
echo "   Image: ${IMAGE_NAME}"
echo "   Tag: ${IMAGE_TAG}"
echo "   Full: ${FULL_IMAGE}"
echo ""

# Check registry login
echo "🔐 Checking registry login..."
if [ "$CONTAINER_TOOL" = "podman" ]; then
    # Extract registry from username
    if [[ "$REGISTRY_USERNAME" == *"/"* ]]; then
        REGISTRY=$(echo "$REGISTRY_USERNAME" | cut -d'/' -f1)
    else
        REGISTRY="docker.io"
    fi
    echo "   Registry: $REGISTRY"
    $CONTAINER_TOOL login $REGISTRY
else
    # Docker
    $CONTAINER_TOOL login
fi

echo ""
echo "🔨 Building image..."
make docker-build IMG="${FULL_IMAGE}" CONTAINER_TOOL="${CONTAINER_TOOL}"

if [ $? -ne 0 ]; then
    echo "❌ Error: Build failed"
    exit 1
fi

echo ""
echo "✅ Build successful!"
echo ""

# Tag as latest
echo "🏷️  Tagging as latest..."
$CONTAINER_TOOL tag "${FULL_IMAGE}" "${LATEST_IMAGE}"

echo ""
echo "🚀 Pushing images..."
echo "   → ${FULL_IMAGE}"
$CONTAINER_TOOL push "${FULL_IMAGE}"

if [ $? -ne 0 ]; then
    echo "❌ Error: Push failed"
    exit 1
fi

echo "   → ${LATEST_IMAGE}"
$CONTAINER_TOOL push "${LATEST_IMAGE}"

echo ""
echo "✅ Successfully pushed!"
echo ""
echo "📋 Summary:"
echo "   Tool: ${CONTAINER_TOOL}"
echo "   Images:"
echo "   - ${FULL_IMAGE}"
echo "   - ${LATEST_IMAGE}"
echo ""
echo "🎯 Deploy with:"
echo "   make deploy IMG=${FULL_IMAGE}"
echo ""

