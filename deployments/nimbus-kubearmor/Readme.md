# Install KubeArmor adapter

Install `nimbus-kubearmor` adapter using the official 5GSEC Helm charts.

```shell
helm repo add 5gsec https://5gsec.github.io/charts
helm repo update 5gsec
helm upgrade --dependency-update --install nimbus-kubearmor 5gsec/nimbus-kubearmor -n nimbus
```

Install `nimbus-kubearmor` adapter using Helm charts locally (for testing)

```bash
cd deployments/nimbus-kubearmor/
helm upgrade --dependency-update --install nimbus-kubearmor . -n nimbus
```

## Values

| Key              | Type   | Default                | Description                                                                |
|------------------|--------|------------------------|----------------------------------------------------------------------------|
| image.repository | string | 5gsec/nimbus-kubearmor | Image repository from which to pull the `nimbus-kubearmor` adapter's image |
| image.pullPolicy | string | Always                 | `nimbus-kubearmor` adapter image pull policy                               |
| image.tag        | string | latest                 | `nimbus-kubearmor` adapter image tag                                       |
| autoDeploy       | bool   | true                   | Auto deploy [KubeArmor]() with default configurations                      |

## Uninstall the KubeArmor adapter

To uninstall, just run:

```bash
helm uninstall nimbus-kubearmor -n nimbus
```
