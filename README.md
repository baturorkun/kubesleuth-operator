# KubeSleuth Operator

[![Go Report Card](https://goreportcard.com/badge/github.com/baturorkun/kubebuilder-demo-operator)](https://goreportcard.com/report/github.com/baturorkun/kubebuilder-demo-operator)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Kubernetes](https://img.shields.io/badge/Kubernetes-1.11+-blue.svg)](https://kubernetes.io/)
[![Go Version](https://img.shields.io/badge/Go-1.24+-00ADD8.svg)](https://golang.org/)

A Kubernetes operator built with Kubebuilder that monitors and tracks non-ready pods across your cluster. PodSleuth provides real-time visibility into pod health issues with an integrated web dashboard.

## 📋 Table of Contents

- [Description](#description)
- [Documentation](#-documentation)
- [Features](#features)
- [Architecture](#architecture)
- [Getting Started](#getting-started)
  - [Prerequisites](#prerequisites)
  - [Installation](#to-deploy-on-the-cluster)
  - [Usage Examples](#create-instances-of-your-solution)
- [Development](#development)
- [How It Works](#how-it-works)
- [Troubleshooting](#troubleshooting)
- [Contributing](#contributing)
- [License](#license)

## Description

**KubeSleuth Operator** is a Kubernetes operator that automatically discovers and monitors non-ready pods across your entire cluster. It provides:

- **Cluster-wide pod monitoring**: Track pods across all namespaces using label selectors
- **Event-driven detection**: Reacts to pod changes in real-time using Kubernetes watch mechanisms
- **Periodic reconciliation**: Ensures no events are missed, even after operator restarts
- **Owner resolution**: Automatically identifies the Deployment or StatefulSet that owns each pod
- **Web dashboard**: Built-in web UI to visualize and filter non-ready pods
- **REST API**: JSON API endpoint for integration with other tools

This operator is useful for operations teams who need to quickly identify and troubleshoot pod health issues in large Kubernetes clusters.

## 📚 Documentation

- **[Quick Start Guide](QUICKSTART.md)** - Get up and running in minutes
- **[Tutorial](TUTORIAL.md)** - Learn how to build Kubernetes operators step-by-step
- **[Deployment Guide](DEPLOYMENT.md)** - Detailed deployment instructions for all environments
- **[API Reference](API.md)** - Complete API documentation for all CRDs
- **[Architecture](ARCHITECTURE.md)** - Deep dive into the operator's design
- **[Contributing](CONTRIBUTING.md)** - How to contribute to this project
- **[Examples](examples/)** - Sample manifests and use cases

## Features

### PodSleuth CRD
A cluster-scoped Custom Resource Definition that monitors pod health across your entire Kubernetes cluster:

- **Cluster-wide monitoring**: Monitor pods across all namespaces with a single resource
- **Label-based filtering**: Use pod label selectors to focus on specific workloads
- **Event-driven architecture**: Reacts to pod CREATE, UPDATE, and DELETE events in real-time
- **Periodic reconciliation**: Configurable reconciliation interval (default: 5 minutes) ensures no events are missed
- **Owner resolution**: Automatically identifies the parent Deployment or StatefulSet for each pod
- **Status tracking**: Maintains a dynamic list of non-ready pods with detailed information
- **Web dashboard**: Built-in web UI accessible at `http://localhost:8082` (when running locally)
- **REST API**: JSON API endpoint at `/api/podsleuths` for programmatic access

**Use Cases**:
- Quickly identify pods that are failing or not ready
- Monitor specific workloads using label selectors
- Track pod health issues across multiple namespaces
- Integrate with alerting systems via the REST API

## Architecture

The operator consists of:
- **Custom Resource Definition (CRD)**: `PodSleuth` - A cluster-scoped resource that defines monitoring configuration
- **Controller**: `PodSleuthReconciler` - Implements event-driven reconciliation logic with periodic safety checks
- **Web Server**: Integrated HTTP server serving the dashboard UI and REST API
- **RBAC Configuration**: Permissions to read pods, deployments, statefulsets, and replicasets across all namespaces
- **Event Watching**: Watches both PodSleuth resources and Pod resources for real-time updates

## Getting Started

### Prerequisites
- go version v1.24.6+
- docker version 17.03+.
- kubectl version v1.11.3+.
- Access to a Kubernetes v1.11.3+ cluster.

### Building and Pushing Container Image

The project includes a universal build script that works with both Docker and Podman.

**Quick method:**

```sh
# With username and version
./build-and-push.sh baturorkun v1.0.0

# With just username (uses v1.0.0 by default)
./build-and-push.sh baturorkun

# Interactive (will prompt for username)
./build-and-push.sh
```

The script automatically detects whether you have Docker or Podman installed and uses the appropriate tool.

**Manual method:**

```sh
# Login to your registry
docker login  # or: podman login docker.io

# Build and push
make docker-build docker-push IMG=your-username/kubebuilder-demo-operator:v1.0.0

# Specify container tool explicitly (optional)
make docker-build docker-push IMG=your-username/kubebuilder-demo-operator:v1.0.0 CONTAINER_TOOL=podman
```

**NOTE:** See [DOCKER.md](DOCKER.md) for detailed instructions.

### To Deploy on the cluster

**Install the CRDs into the cluster:**

```sh
make install
```

**Deploy the Manager to the cluster with your image:**

```sh
make deploy IMG=your-username/kubebuilder-demo-operator:v1.0.0
```

> **NOTE**: If you encounter RBAC errors, you may need to grant yourself cluster-admin
privileges or be logged in as admin.

**Create instances of your solution**

#### Example: PodSleuth

Create a PodSleuth resource to monitor pods:

```yaml
apiVersion: apps.ops.dev/v1alpha1
kind: PodSleuth
metadata:
  name: podsleuth-sample
spec:
  # Optional: Reconcile interval (default: 5 minutes)
  reconcileInterval: 5m
  
  # Optional: Pod label selector to filter pods across all namespaces
  # If not specified, monitors all pods in all namespaces
  podLabelSelector:
    matchLabels:
      environment: test
    # or use matchExpressions for more complex filtering
    # matchExpressions:
    #   - key: environment
    #     operator: In
    #     values:
    #     - production
    #     - staging
```

Apply it:
```sh
kubectl apply -f config/samples/infra_v1alpha1_podsleuth.yaml
```

Check the status:
```sh
kubectl get podsleuth
kubectl get podsleuth podsleuth-sample -o yaml
kubectl get podsleuth podsleuth-sample -o jsonpath='{.status.nonReadyPods[*].name}'
```

#### Access the Web Dashboard

When running locally with `make run`:
- Open your browser to `http://localhost:8082`
- The dashboard shows all non-ready pods with filtering and search capabilities

When deployed to a cluster:
```sh
# Port forward to access the dashboard
kubectl port-forward -n kubebuilder-demo-operator-system svc/controller-manager-dashboard-service 8082:8082
# Then open http://localhost:8082 in your browser
```

#### Testing with Sample Deployments

See `test/kubernetes/` directory for example deployments and a test script:
```sh
cd test/kubernetes
bash run.sh
```

This creates test deployments with non-ready pods that PodSleuth will detect.

### To Uninstall
**Delete the instances (CRs) from the cluster:**

```sh
kubectl delete -k config/samples/
```

**Delete the APIs(CRDs) from the cluster:**

```sh
make uninstall
```

**UnDeploy the controller from the cluster:**

```sh
make undeploy
```

## Project Distribution

Following the options to release and provide this solution to the users.

### By providing a bundle with all YAML files

1. Build the installer for the image built and published in the registry:

```sh
make build-installer IMG=<some-registry>/kubebuilder-demo-operator:tag
```

**NOTE:** The makefile target mentioned above generates an 'install.yaml'
file in the dist directory. This file contains all the resources built
with Kustomize, which are necessary to install this project without its
dependencies.

2. Using the installer

Users can just run 'kubectl apply -f <URL for YAML BUNDLE>' to install
the project, i.e.:

```sh
kubectl apply -f https://raw.githubusercontent.com/<org>/kubebuilder-demo-operator/<tag or branch>/dist/install.yaml
```

### By providing a Helm Chart

1. Build the chart using the optional helm plugin

```sh
kubebuilder edit --plugins=helm/v2-alpha
```

2. See that a chart was generated under 'dist/chart', and users
can obtain this solution from there.

**NOTE:** If you change the project, you need to update the Helm Chart
using the same command above to sync the latest changes. Furthermore,
if you create webhooks, you need to use the above command with
the '--force' flag and manually ensure that any custom configuration
previously added to 'dist/chart/values.yaml' or 'dist/chart/manager/manager.yaml'
is manually re-applied afterwards.

## Development

### Running Locally

You can run the operator locally for development purposes:

```sh
make install  # Install CRDs
make run      # Run the operator locally
```

When running locally, the dashboard is automatically available at:
- **Dashboard**: `http://localhost:8082`
- **API**: `http://localhost:8082/api/podsleuths`

The operator will connect to your current `kubectl` context and monitor pods in that cluster.

### Running Tests

```sh
make test  # Run unit tests
```

### Building the Operator

```sh
make build  # Build the operator binary
```

## How It Works

### PodSleuth Controller

The PodSleuth controller implements an event-driven monitoring pattern with periodic reconciliation:

1. **Event Watching**:
   - Watches `PodSleuth` resources for configuration changes
   - Watches `Pod` resources across all namespaces for pod state changes
   - Uses Kubernetes LIST+WATCH mechanism to get initial state and stream updates

2. **Reconciliation Process**:
   - When triggered (by event or periodic timer), lists all pods matching the label selector
   - Checks each pod's `Ready` condition to identify non-ready pods
   - Resolves owner references to find the parent Deployment or StatefulSet
   - Updates the PodSleuth status with the current list of non-ready pods
   - Logs non-ready pods with their owner information

3. **Owner Resolution**:
   - Traverses pod owner references to find the controller owner
   - Handles ReplicaSet → Deployment relationships
   - Directly identifies StatefulSet owners

4. **Periodic Reconciliation**:
   - Default interval: 5 minutes (configurable via `reconcileInterval`)
   - Ensures no events are missed, even after operator restarts
   - Acts as a safety net for scenarios like node restarts

**Key Implementation Details:**
- The controller uses `findObjectsForPod` function to map pod changes to affected PodSleuth resources
- RBAC permissions include read access to pods, deployments, statefulsets, and replicasets across all namespaces
- The reconciliation loop is triggered by both PodSleuth changes and Pod changes
- Cluster-scoped resource allows monitoring across all namespaces with a single resource

### Web Dashboard

The integrated web server provides:
- **Real-time updates**: Auto-refreshes every 10 seconds
- **Filtering**: Search by namespace, phase, owner, or pod name
- **Statistics**: Overview of total pods, namespaces, and deployments
- **REST API**: JSON endpoint for programmatic access

## Troubleshooting

### Operator logs
```sh
kubectl logs -n kubebuilder-demo-operator-system deployment/kubebuilder-demo-operator-controller-manager
```

### Check CRD installation
```sh
kubectl get crds | grep apps.ops.dev
kubectl get podsleuths
```

### Verify RBAC permissions
```sh
kubectl describe clusterrole kubebuilder-demo-operator-manager-role | grep podsleuth
```

### Check dashboard service
```sh
kubectl get svc -n kubebuilder-demo-operator-system | grep dashboard
```

### View PodSleuth status
```sh
kubectl get podsleuth <name> -o yaml
kubectl get podsleuth <name> -o jsonpath='{.status.nonReadyPods[*]}' | jq
```

## Contributing

Contributions are welcome! This is a demo project designed to help learn Kubebuilder and Kubernetes operators.

To contribute:
1. Fork the repository
2. Create a feature branch (`git checkout -b feature/amazing-feature`)
3. Commit your changes (`git commit -m 'Add some amazing feature'`)
4. Push to the branch (`git push origin feature/amazing-feature`)
5. Open a Pull Request

**NOTE:** Run `make help` for more information on all potential `make` targets

More information can be found via the [Kubebuilder Documentation](https://book.kubebuilder.io/introduction.html)

## License

Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.

