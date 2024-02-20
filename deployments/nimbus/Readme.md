# Install Nimbus

Install Nimbus operator using the official 5GSEC Helm charts.

```shell
helm repo add 5gsec https://5gsec.github.io/charts
helm repo update 5gsec
helm upgrade --install nimbus-operator 5gsec/nimbus -n nimbus --create-namespace
```

Install Nimbus using Helm charts locally (for testing)

```bash
cd deployments/nimbus/
helm upgrade --install nimbus-operator . -n nimbus --create-namespace
```

## Values

| Key              | Type   | Default      | Description                                            |
|------------------|--------|--------------|--------------------------------------------------------|
| image.repository | string | 5gsec/nimbus | Image repository from which to pull the operator image |
| image.pullPolicy | string | Always       | Operator image pull policy                             |
| image.tag        | string | latest       | Operator image tag                                     |

## Verify if all the resources are up and running

Once done, the following resources will exist in your cluster:

```shell
$  kubectl get all -n nimbus -l app.kubernetes.io/instance=nimbus-operator
NAME                                   READY   STATUS    RESTARTS   AGE
pod/nimbus-operator-57dc75bc4d-9gd5n   1/1     Running   0          20m

NAME                              READY   UP-TO-DATE   AVAILABLE   AGE
deployment.apps/nimbus-operator   1/1     1            1           20m

NAME                                         DESIRED   CURRENT   READY   AGE
replicaset.apps/nimbus-operator-57dc75bc4d   1         1         1       20m
```

## Uninstall the Operator

To uninstall, just run:

```bash
helm uninstall nimbus-operator -n nimbus
```
