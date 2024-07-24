# Install Nimbus

Install Nimbus operator using the official 5GSEC Helm charts.

```shell
helm repo add 5gsec https://5gsec.github.io/charts
helm repo update 5gsec
helm upgrade --dependency-update --install nimbus-operator 5gsec/nimbus -n nimbus --create-namespace
```

Install Nimbus using Helm charts locally (for testing)

```bash
cd deployments/nimbus/
helm upgrade --dependency-update --install nimbus-operator . -n nimbus --create-namespace
```

## Values

| Key                  | Type   | Default      | Description                                                                                                               |
|----------------------|--------|--------------|---------------------------------------------------------------------------------------------------------------------------|
| image.repository     | string | 5gsec/nimbus | Image repository from which to pull the operator image                                                                    |
| image.pullPolicy     | string | Always       | Operator image pull policy                                                                                                |
| image.tag            | string | latest       | Operator image tag                                                                                                        |
| autoDeploy.kubearmor | bool   | true         | Auto deploy [KubeArmor](https://kubearmor.io/) adapter                                                                    |
| autoDeploy.netpol    | bool   | true         | Auto deploy [Kubernetes NetworkPolicy](https://kubernetes.io/docs/concepts/services-networking/network-policies/) adapter |
| autoDeploy.kyverno   | bool   | true         | Auto deploy [Kyverno](https://kyverno.io/) adapter                                                                        |

## Uninstall the Operator

To uninstall, just run:

```bash
helm uninstall nimbus-operator -n nimbus
```
