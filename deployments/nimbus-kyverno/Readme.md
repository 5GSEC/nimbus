# Install Kyverno adapter

> [!Note]
> The `nimbus-kyverno` adapter leverages the [kyverno](https://kyverno.io/) security engine for its functionality.
> To use this adapter, you'll need kyverno installed. Please
> follow [this](https://kyverno.io/docs/installation/methods/) guide for
> installation.
> Creating a Policy and ClusterPolicy resource without Kyverno will have no effect.

Install `nimbus-kyverno` adapter using the official 5GSEC Helm charts.

```shell
helm repo add 5gsec https://5gsec.github.io/charts
helm repo update 5gsec
helm upgrade --install nimbus-kyverno 5gsec/nimbus-kyverno -n nimbus
```

Install `nimbus-kyverno` adapter using Helm charts locally (for testing)

```bash
cd deployments/nimbus-kyverno/
helm upgrade --install nimbus-kyverno . -n nimbus
```

## Values

| Key              | Type   | Default                | Description                                                                |
|------------------|--------|------------------------|----------------------------------------------------------------------------|
| image.repository | string | 5gsec/nimbus-kyverno | Image repository from which to pull the `nimbus-kyverno` adapter's image |
| image.pullPolicy | string | Always                 | `nimbus-kyverno` adapter image pull policy                               |
| image.tag        | string | latest                 | `nimbus-kyverno` adapter image tag                                       |

## Verify if all the resources are up and running

Once done, the following resources will exist in your cluster:

```shell
$ kubectl get all -n nimbus -l app.kubernetes.io/instance=nimbus-kyverno
NAME                                    READY   STATUS    RESTARTS   AGE
pod/nimbus-kyverno-7f6854cf8f-gm7c8   1/1     Running   0          3m25s

NAME                               READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/nimbus-kyverno   1/1     1            1           3m25s

NAME                                          DESIRED   CURRENT   READY   AGE
replicaset.apps/nimbus-kyverno-7f6854cf8f   1         1         1       3m25s
```

## Uninstall the Kyverno adapter

To uninstall, just run:

```bash
helm uninstall nimbus-kyverno -n nimbus
```
