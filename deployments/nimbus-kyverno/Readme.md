# Install Kyverno adapter

Install `nimbus-kyverno` adapter using the official 5GSEC Helm charts.

```shell
helm repo add 5gsec https://5gsec.github.io/charts
helm repo update 5gsec
helm upgrade --dependency-update --install nimbus-kyverno 5gsec/nimbus-kyverno -n nimbus
```

Install `nimbus-kyverno` adapter using Helm charts locally (for testing)

```bash
cd deployments/nimbus-kyverno/
helm upgrade --dependency-update --install nimbus-kyverno . -n nimbus
```

## Values

| Key              | Type   | Default              | Description                                                                                                               |
|------------------|--------|----------------------|---------------------------------------------------------------------------------------------------------------------------|
| image.repository | string | 5gsec/nimbus-kyverno | Image repository from which to pull the `nimbus-kyverno` adapter's image                                                  |
| image.pullPolicy | string | Always               | `nimbus-kyverno` adapter image pull policy                                                                                |
| image.tag        | string | latest               | `nimbus-kyverno` adapter image tag                                                                                        |
| autoDeploy       | bool   | true                 | Auto deploy [Kyverno](https://kyverno.io/) in [Standalone](https://kyverno.io/docs/installation/methods/#standalone) mode |

## Uninstall the Kyverno adapter

To uninstall, just run:

```bash
helm uninstall nimbus-kyverno -n nimbus
```
