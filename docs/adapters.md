# Adapters

Clone your forked repository onto your local machine.

```shell
git clone git@github.com:<your-username>/nimbus.git
```

## nimbus-kubearmor

### From source

**Requires installing corresponding security engine**:
Follow [this](https://docs.kubearmor.io/kubearmor/quick-links/deployment_guide) guide to install
KubeArmor.

Navigate to `nimbus-kubearmor` directory:

```shell
cd nimbus/pkg/adapter/nimbus-kubearmor
```

Run adapter:

```shell
make run
```

### From Helm chart

Follow [this](../../deployments/nimbus-kubearmor/Readme.md) guide to install using a helm cahrt.

## nimbus-netpol

> [!Note]
> The `nimbus-netpol` adapter leverages
> the [network plugin](https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/network-plugins/).
> To use network policies, you must be using a networking solution which supports NetworkPolicy. Creating a
> NetworkPolicy resource without a controller that implements it will have no effect.

### From source

Navigate to `nimbus-netpol` directory:

```shell
cd nimbus/pkg/adapter/nimbus-netpol
```

Run adapter:

```shell
make run
```

### From Helm chart

Follow [this](../../deployments/nimbus-netpol/Readme.md) to install using a helm chart.

## nimbus-kyverno

**Requires installing corresponding security engine**:
Follow [this](https://kyverno.io/docs/installation/) guide to install
Kyverno.

### From source

Navigate to `nimbus-kyverno` directory:

```shell
cd nimbus/pkg/adapter/nimbus-kyverno
```

Run adapter:

```shell
make run
```

### From Helm chart

Follow [this](../../deployments/nimbus-kyverno/Readme.md) to install using a helm chart.

## nimbus-k8tls

### From source

Navigate to `nimbus-k8tls` directory:

```shell
cd nimbus/pkg/adapter/nimbus-kyverno
```

Run adapter:

```shell
make run
```

### From Helm chart

Follow [this](../../deployments/nimbus-k8tls/Readme.md) to install using a helm chart.
