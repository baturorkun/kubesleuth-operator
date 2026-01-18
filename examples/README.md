# Examples

This directory contains examples demonstrating how to use the Kubebuilder Demo Operator.

## Directory Structure

- `featureflag/` - FeatureFlag examples
- `deploymentguard/` - DeploymentGuard examples
- `combined/` - Examples using both CRDs together

## FeatureFlag Examples

### Basic Feature Flag
Simple example of enabling/disabling a feature:
```bash
kubectl apply -f featureflag/basic-feature.yaml
```

### Multiple Feature Flags
Example with multiple features:
```bash
kubectl apply -f featureflag/multiple-features.yaml
```

## DeploymentGuard Examples

### Protecting a Single Deployment
```bash
# Create deployment first
kubectl apply -f deploymentguard/nginx-deployment.yaml
# Then create the guard
kubectl apply -f deploymentguard/nginx-guard.yaml
```

### Protecting Multiple Deployments
```bash
kubectl apply -f deploymentguard/multiple-deployments.yaml
kubectl apply -f deploymentguard/multiple-guards.yaml
```

## Testing DeploymentGuard

After applying the examples, test the guard:

```bash
# Scale deployment to 0
kubectl scale deployment nginx-deployment --replicas=0

# Watch it automatically scale back to 1
kubectl get deployment nginx-deployment -w
```

## Combined Examples

See `combined/` directory for examples using both CRDs together in a complete scenario.

## Cleanup

To remove all examples:

```bash
kubectl delete -f featureflag/
kubectl delete -f deploymentguard/
kubectl delete -f combined/
```

