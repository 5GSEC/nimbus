# Install KubeArmor adapter

> [!Note]
> The `nimbus-kubearmor` adapter leverages the [KubeArmor](https://kubearmor.io) security engine for its functionality.
> To use this adapter, you'll need KubeArmor installed. Please
> follow [this](https://github.com/kubearmor/KubeArmor/blob/main/getting-started/deployment_guide.md) guide for
> installation.
> Creating a KubeArmorPolicy resource without KubeArmor will have no effect.

Install `nimbus-kubearmor` adapter using the official 5GSEC Helm charts.

```shell
helm repo add 5gsec https://5gsec.github.io/charts
helm repo update 5gsec
helm upgrade --install nimbus-kubearmor 5gsec/nimbus-kubearmor -n nimbus
```

Install `nimbus-kubearmor` adapter using Helm charts locally (for testing)

```bash
cd deployments/nimbus-kubearmor/
helm upgrade --install nimbus-kubearmor . -n nimbus
```

## Values

| Key              | Type   | Default                | Description                                                                |
|------------------|--------|------------------------|----------------------------------------------------------------------------|
| image.repository | string | 5gsec/nimbus-kubearmor | Image repository from which to pull the `nimbus-kubearmor` adapter's image |
| image.pullPolicy | string | Always                 | `nimbus-kubearmor` adapter image pull policy                               |
| image.tag        | string | latest                 | `nimbus-kubearmor` adapter image tag                                       |

## Verify if all the resources are up and running

Once done, the following resources will exist in your cluster:

```shell
$ kubectl get all -n nimbus -l app.kubernetes.io/instance=nimbus-kubearmor
NAME                                    READY   STATUS    RESTARTS   AGE
pod/nimbus-kubearmor-7f6854cf8f-gm7c8   1/1     Running   0          3m25s

NAME                               READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/nimbus-kubearmor   1/1     1            1           3m25s

NAME                                          DESIRED   CURRENT   READY   AGE
replicaset.apps/nimbus-kubearmor-7f6854cf8f   1         1         1       3m25s
```

## Uninstall the KubeArmor adapter

To uninstall, just run:

```bash
helm uninstall nimbus-kubearmor -n nimbus
```
