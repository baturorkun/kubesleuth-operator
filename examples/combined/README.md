# Complete Application Example
```
kubectl delete -f complete-app.yaml
```bash

## Cleanup

```
kubectl logs -n kubebuilder-demo-operator-system deployment/kubebuilder-demo-operator-controller-manager -f
# If deployed in cluster

# Check terminal output
# If running locally
```bash
Watch operator logs:

## Monitor

```
kubectl get featureflag new-ui-feature -o yaml
# Check status

kubectl patch featureflag new-ui-feature -p '{"spec":{"enabled":false}}' --type=merge
# Disable the new UI
```bash
Toggle a feature:

### Test FeatureFlag

```
kubectl get deployment web-frontend -w
# Watch it automatically scale back

kubectl scale deployment web-frontend --replicas=0
```bash
Try to scale down the frontend:

### Test DeploymentGuard

## Test the Setup

```
kubectl get featureflags
# Verify feature flags

kubectl get deploymentguards
# Verify guards

kubectl get deployments
# Verify deployments

kubectl apply -f complete-app.yaml
# Apply all resources
```bash

## Deploy the Complete Stack

- Feature flags for controlling application features
- Backend API deployment (protected by DeploymentGuard)
- Frontend deployment (protected by DeploymentGuard)
A web application with:

## Scenario

This example demonstrates a complete application scenario using both FeatureFlag and DeploymentGuard CRDs.


