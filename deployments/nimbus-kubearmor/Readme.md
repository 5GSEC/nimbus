# Install KubeArmor adapter

> [!NOTE]
> `nimbus-kubearmor` adapter depends on [KubeArmor](https://kubearmor.io) security engine for its functionalities.
> Follow [this](https://github.com/kubearmor/KubeArmor/blob/main/getting-started/deployment_guide.md) guide to install
> KubeArmor.


Install `nimbus-kubearmor` adapter using Helm charts locally (for testing)

```bash
cd deployments/nimbus-kubearmor/
helm upgrade --install nimbus-kubearmor . -n nimbus
```

## Values

| Key              | Type   | Default                | Description                                                                |
|------------------|--------|------------------------|----------------------------------------------------------------------------|
| image.repository | string | 5gsec/nimbus-kubearmor | Image repository from which to pull the `nimbus-kubearmor` adapter's image |
| image.pullPolicy | string | IfNotPresent           | `nimbus-kubearmor` adapter image pull policy                               |
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

## Uninstall the Operator

To uninstall, just run:

```bash
helm uninstall nimbus-kubearmor -n nimbus
```
