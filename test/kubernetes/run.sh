#!/bin/bash

# Script to deploy test deployments for PodSleuth testing
# This creates 5 deployments across 3 namespaces with varied error scenarios

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
YAML_FILE="${SCRIPT_DIR}/deployments.yaml"

echo "========================================="
echo "Deploying test deployments for PodSleuth"
echo "========================================="
echo ""

# Check if kubectl is available
if ! command -v kubectl &> /dev/null; then
    echo "Error: kubectl is not installed or not in PATH"
    exit 1
fi

# Check if we can connect to cluster
if ! kubectl cluster-info &> /dev/null; then
    echo "Error: Cannot connect to Kubernetes cluster"
    exit 1
fi

echo "Step 1: Creating namespaces and applying deployments..."
kubectl apply -f "${YAML_FILE}"

echo ""
echo "Step 2: Waiting a moment for pods to start creating..."
sleep 5

echo ""
echo "Step 3: Checking deployment status across all test namespaces..."
for ns in podsleuth-test1 podsleuth-test2 podsleuth-test3; do
    echo ""
    echo "--- Namespace: $ns ---"
    kubectl get deployments -n "$ns" -l environment=test 2>/dev/null || echo "No deployments found in $ns"
done

echo ""
echo "Step 4: Checking pod status across all test namespaces..."
for ns in podsleuth-test1 podsleuth-test2 podsleuth-test3; do
    echo ""
    echo "--- Namespace: $ns ---"
    kubectl get pods -n "$ns" -l environment=test 2>/dev/null || echo "No pods found in $ns"
done

echo ""
echo "Step 5: Showing pods that are not ready across all namespaces..."
for ns in podsleuth-test1 podsleuth-test2 podsleuth-test3; do
    echo ""
    echo "--- Non-ready pods in $ns ---"
    kubectl get pods -n "$ns" -l environment=test --field-selector=status.phase!=Succeeded --no-headers 2>/dev/null | while read line; do
        if [ -n "$line" ]; then
            pod_name=$(echo "$line" | awk '{print $1}')
            ready_status=$(echo "$line" | awk '{print $2}')
            if [[ "$ready_status" != *"/"* ]] || [[ "$ready_status" == "0/"* ]]; then
                echo "Pod not ready: $pod_name (Status: $ready_status)"
                kubectl describe pod -n "$ns" "$pod_name" 2>/dev/null | grep -A 5 "Conditions:" || true
            fi
        fi
    done
done

echo ""
echo "Step 6: Creating PodSleuth CRD to monitor test pods..."
PODSLEUTH_FILE="${SCRIPT_DIR}/podsleuth-example.yaml"
if [ -f "${PODSLEUTH_FILE}" ]; then
    kubectl apply -f "${PODSLEUTH_FILE}"
    echo "PodSleuth created. Check status with:"
    echo "  kubectl get podsleuth podsleuth-test -o yaml"
    echo "  kubectl get podsleuth podsleuth-test -o jsonpath='{.status.nonReadyPods[*].name}'"
else
    echo "Warning: PodSleuth example file not found at ${PODSLEUTH_FILE}"
fi

echo ""
echo "========================================="
echo "Deployment complete!"
echo "========================================="
echo ""
echo "To check all non-ready pods across all test namespaces:"
echo "  kubectl get pods --all-namespaces -l environment=test --field-selector=status.phase!=Succeeded"
echo ""
echo "To check pods in specific namespace:"
echo "  kubectl get pods -n podsleuth-test1 -l environment=test"
echo "  kubectl get pods -n podsleuth-test2 -l environment=test"
echo "  kubectl get pods -n podsleuth-test3 -l environment=test"
echo ""
echo "To check PodSleuth status:"
echo "  kubectl get podsleuth podsleuth-test"
echo "  kubectl get podsleuth podsleuth-test -o yaml"
echo ""
echo "To delete all test deployments and PodSleuth:"
echo "  kubectl delete -f ${YAML_FILE}"
echo "  kubectl delete -f ${PODSLEUTH_FILE}"
echo ""
echo "To watch pods in real-time:"
echo "  kubectl get pods --all-namespaces -l environment=test -w"
echo ""
