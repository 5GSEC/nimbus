# Install NetworkPolicy adapter

> [!Note]
> The `nimbus-netpol` adapter leverages
> the [network plugin](https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/network-plugins/).
> To use network policies, you must be using a networking solution which supports NetworkPolicy. Creating a
> NetworkPolicy resource without a controller that implements it will have no effect.

Install `nimbus-netpol` adapter using the official 5GSEC Helm charts.

```shell
helm repo add 5gsec https://5gsec.github.io/charts
helm repo update 5gsec
helm upgrade --install nimbus-netpol 5gsec/nimbus-netpol -n nimbus
```

Install `nimbus-netpol` adapter using Helm charts locally (for testing)

```bash
cd deployments/nimbus-netpol/
helm upgrade --install nimbus-netpol . -n nimbus
```

## Values

| Key              | Type   | Default             | Description                                                             |
|------------------|--------|---------------------|-------------------------------------------------------------------------|
| image.repository | string | 5gsec/nimbus-netpol | Image repository from which to pull the `nimbus-netpol` adapter's image |
| image.pullPolicy | string | Always              | `nimbus-netpol` adapter image pull policy                               |
| image.tag        | string | latest              | `nimbus-netpol` adapter image tag                                       |

## Verify if all the resources are up and running

Once done, the following resources will exist in your cluster:

```shell
$ kubectl get all -n nimbus -l app.kubernetes.io/instance=nimbus-netpol
NAME                                 READY   STATUS    RESTARTS   AGE
pod/nimbus-netpol-6ccd868c49-wb54j   1/1     Running   0          3m57s

NAME                            READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/nimbus-netpol   1/1     1            1           3m58s

NAME                                       DESIRED   CURRENT   READY   AGE
replicaset.apps/nimbus-netpol-6ccd868c49   1         1         1       3m57s
```

## Uninstall the NetworkPolicy adapter

To uninstall, just run:

```bash
helm uninstall nimbus-netpol -n nimbus
```
