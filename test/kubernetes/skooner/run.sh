#!/bin/bash

echo "Installing Metrics Server..."

kubectl apply -f https://github.com/kubernetes-sigs/metrics-server/releases/latest/download/components.yaml

# Patch it to allow insecure TLS (required for Kind/Podman)
kubectl patch deployment metrics-server -n kube-system --type='json' -p='[{"op": "add", "path": "/spec/template/spec/containers/0/args/-", "value": "--kubelet-insecure-tls"}]'

echo "Creating Dashboard Admin Fix..."
kubectl create clusterrolebinding dashboard-admin-fix --clusterrole=cluster-admin --serviceaccount=kube-system:default

echo "Installing Skooner..."

kubectl apply -f https://raw.githubusercontent.com/skooner-k8s/skooner/master/kubernetes-skooner.yaml

echo "Creating Token for Default Service Account..."
TOKEN=$(kubectl -n kube-system create token default)

echo "Token: $TOKEN"

echo "Port Forwarding Skooner..."

kubectl -n kube-system port-forward svc/skooner 8080:80

echo "Skooner installed and port forwarded. Access it at http://localhost:8080"
echo "Use the token: $TOKEN to authenticate."