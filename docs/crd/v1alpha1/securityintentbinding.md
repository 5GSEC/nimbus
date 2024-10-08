# Nimbus `SecurityIntentBinding` Specification

## Description

A `SecurityIntentBinding` object defines how a specific `SecurityIntent` is applied to resources within a namespace. It
essentially binds an intent to target resources like pods.

## Spec

```text
apiVersion: intent.security.nimbus.com/v1alpha1
kind: SecurityIntentBinding
metadata:
  name: [ securityIntentBinding name ]
  namespace: [ namespace name ]                  # Namespace where the binding applies
spec:
  intents:
    - name: [ intent-to-bind-name ]              # Name of the SecurityIntent to apply 
  selector:
    workloadSelector:
      matchLabels:
        key1: value1
       # ... (additional label selectors)
```

### Explanation of Fields

### Common Fields

- `apiVersion`: Identifies the version of the API group for this resource. This remains constant for all Nimbus
  policies.
- `kind`: Specifies the resource type, which is always `SecurityIntentBinding` in this case.
- `metadata`: Contains standard Kubernetes metadata like the resource name, which you define in the  `.metadata`
  placeholder.

```yaml
apiVersion: intent.security.nimbus.com/v1alpha1
kind: SecurityIntentBinding
metadata:
  name: securityIntentBinding-name
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
    - name: dns-manipulation
...
```

### Selector

- `spec.selector` **(Required)**: Defines the Kubernetes [workload](https://kubernetes.io/docs/concepts/workloads/) that
  will be
  subject to the bound `SecurityIntent` policies.
    - `workloadSelector` : Selects resources based on labels.
        - `matchLabels`: A key-value map where each key represents a label on the target resource and its corresponding
          value specifies the expected value for that label. Resources with matching labels will be targeted by the
          bound `SecurityIntent`.

```yaml
...
selector:
  workloadSelector:
    matchLabels:
      key1: value
...
```
