# Nimbus `SecurityIntent` Specification

## Description

A `SecurityIntent` resource defines the desired security state for your Kubernetes cluster at a high level. It describes
**_what security outcome you want_**, not how to achieve it. This resource is cluster-scoped resource.

## Spec

```text
apiVersion: intent.security.nimbus.com/v1alpha1
kind: SecurityIntent
metadata:
  name: [SecurityIntent name]
spec:
  intent:
    id: [supported intent ID]                    # ID from the predefined pool
    action: [Audit|Block]                        # Block by default.
    params:                                      # Optional. Parameters allows fine-tuning of intents to specific requirements.
     key: ["value1", "value2"]
```

### Explanation of Fields

### Common Fields

- `apiVersion`: Identifies the version of the API group for this resource. This remains constant for all Nimbus
  policies.
- `kind`: Specifies the resource type, which is always `SecurityIntent` in this case.
- `metadata`: Contains standard Kubernetes metadata like the resource name, which you define in the  `.metadata.name`
  placeholder.

```yaml
apiVersion: intent.security.nimbus.com/v1alpha1
kind: SecurityIntent
metadata:
  name: securityIntent-name
```

### Intent Fields

The `.spec.intent` field defines the specific security behavior you want:

- `id` **(Required)**: This refers to a predefined intent ID from the [pool]( ../../intents/supportedIntents).
  Security engines use this ID to generate corresponding security policies.
- `action` **(Required)**: This defines how the generated policy will be enforced. Supported actions are `Audit` (logs
  the violation) and `Block` (prevents the violation). By default, the action is set to `Block`.
- `params` **(Optional)**: Parameters are key-value pairs that allow you to customize the chosen intent for your
  specific needs. Refer to the [supported intents]( ../../intents/supportedIntents) for details on available
  parameters and their valid values.

```yaml
...
spec:
  intent:
    id: assessTLS
    action: Audit
    params:
      schedule: [ "* * * * *" ]
...
```
