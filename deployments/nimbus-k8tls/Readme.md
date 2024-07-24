# Install [k8tls](https://github.com/kubearmor/k8tls) adapter

Install `nimbus-k8tls` adapter using the official 5GSEC Helm charts.

```shell
helm repo add 5gsec https://5gsec.github.io/charts
helm repo update 5gsec
helm upgrade --install nimbus-k8tls 5gsec/nimbus-k8tls -n nimbus
```

Install `nimbus-k8tls` adapter using Helm charts locally (for testing)

```bash
cd deployments/nimbus-k8tls/
helm upgrade --install nimbus-k8tls . -n nimbus
```

## Values

| Key              | Type   | Default            | Description                                                            |
|------------------|--------|--------------------|------------------------------------------------------------------------|
| image.repository | string | 5gsec/nimbus-k8tls | Image repository from which to pull the `nimbus-k8tls` adapter's image |
| image.pullPolicy | string | Always             | `nimbus-k8tls` adapter image pull policy                               |
| image.tag        | string | latest             | `nimbus-k8tls` adapter image tag                                       |

## Verify if all the resources are up and running

Once done, the following resources will exist in your cluster:

```shell
$ kubectl get all -n nimbus -l app.kubernetes.io/instance=nimbus-k8tls
NAME                     READY   STATUS    RESTARTS   AGE
pod/nimbus-k8tls-q2tt7   1/1     Running   0          3m8s

NAME                          DESIRED   CURRENT   READY   UP-TO-DATE   AVAILABLE   NODE SELECTOR   AGE
daemonset.apps/nimbus-k8tls   1         1         1       1            1           <none>          3m8s
```

## Uninstall the k8tls adapter

To uninstall:

```bash
helm uninstall nimbus-k8tls -n nimbus
```
