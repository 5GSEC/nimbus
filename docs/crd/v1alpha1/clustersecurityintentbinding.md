# Nimbus ClusterSecurityIntentBinding Specification

## Description

A `ClusterSecurityIntentBinding` binds specific `SecurityIntent` resources to targeted resources within a cluster.
Unlike its namespaced counterpart (`SecurityIntentBinding`), it operates at the cluster level, enabling intent
application across multiple namespaces.

```text
apiVersion: intent.security.nimbus.com/v1alpha1
kind: ClusterSecurityIntentBinding
metadata:
  name: [ ClusterSecurityIntentBinding name ]
spec:
  intents:
    - name: [ intent-to-bind-name ]
  selector:
    workloadSelector:                              # --> optional
      matchLabels:
        [ key1 ]: [ value1 ]
          [ keyN ]: [ valueN ]
    nsSelector:                                   # --> optional
      excludeNames:                               # --> optional
        - [ namespace-to-exclude ]
      matchNames:                                 # --> optional
        - [ namespace-to-include ]
```

### Explanation of Fields

### Common Fields

- `apiVersion`: Identifies the version of the API group for this resource. This remains constant for all Nimbus
  policies.
- `kind`: Specifies the resource type, which is always `ClusterSecurityIntentBinding` in this case.
- `metadata`: Contains standard Kubernetes metadata like the resource name, which you define in the  `.metadata`
  placeholder.

```yaml
apiVersion: intent.security.nimbus.com/v1alpha1
kind: ClusterSecurityIntentBinding
metadata:
  name: cluster-security-intent-binding-name
```

### Intents

- `.spec.intents` **(Required)**: An array containing one or more objects specifying the names of `SecurityIntent`
  resources to be
  bound. Each object has a single field:
    - `name` **(Required)**: The name of the `SecurityIntent` that should be applied to resources selected by this
      binding.

```yaml
...
spec:
  intents:
    - name: assess-tls-scheduled
...
```

### Selector

`ClusterSecurityIntentBinding` has different selector to bind intent(s) to resources across namespaces.

- `.spec.selector` **(Required)**: Defines resources targeted by the bound `SecurityIntent` policies.
    - `workloadSelector` **(Optional)**: Same selector as `SecurityIntentBinding`.
    - `nsSelector` **(Optional)**: Namespace selection criteria.
        - `excludeNames` **(Optional)**: Exclude namespaces from the binding.
        - `matchNames` **(Optional)**: Include namespaces in the binding.
          Note: At least one of `matchNames` or `excludeNames` must be specified in `nsSelector`.

Here are some examples:

- [Apply to all namespaces](../../../examples/clusterscoped/csib-1-all-ns-selector.yaml)
- [Apply to specific namespaces](../../../examples/clusterscoped/csib-2-match-names.yaml)
- [Apply to all namespaces excluding specific namespaces](../../../examples/clusterscoped/csib-3-exclude-names.yaml)
