# Supported Intents

| IntentID                                | Parameters | Description                                                                                                                                                                                                   |
|-----------------------------------------|------------|---------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------|
| `dnsManipulation`                       | NA         | An adversary can manipulate DNS requests to redirect network  traffic, exfiltrate and potentially reveal end user activity.                                                                                   |
| `swDeploymentTools`                     | NA         | Adversaries may gain access to and use third-party software suites installed within an enterprise network, such as administration, monitoring, and deployment systems, to move laterally through the network. |
| `assessTLS`                             | `schedule` | Assess the TLS configuration to ensure compliance with the security standards.                                                                                                                                |
| `unAuthorizedSaTokenAccess`             | NA         | K8s mounts the service account token by default in each pod even if there is no app using it. Attackers use these service account tokens to do lateral movements.                                             |
| `escapeToHost`                          | Todo @Ved  |                                                                                                                                                                                                               |
| `preventExecutionFromTempOrLogsFolders` | Todo @Ved  |                                                                                                                                                                                                               |
| `denyExternalNetworkAccess`             | Todo @Ved  |                                                                                                                                                                                                               |
| `cocoWorkload`                          | Todo @Ved  |                                                                                                                                                                                                               |

Here are the examples and tutorials:

- [Namespace scoped](../../examples/namespaced)
- [Cluster scoped](../../examples/clusterscoped)
- [Detailed examples](../intents)
