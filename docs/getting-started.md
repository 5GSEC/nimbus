# Getting Started

This guide walks users through the steps to easily install and run the Nimbus operator. Each step includes the commands
needed and their descriptions to help users understand and proceed with each step.

# Prerequisites

Before you begin, set up the following:

- [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl) version 1.26 or later.
- A Kubernetes cluster running version 1.26 or later.

# Nimbus

There are various ways of installing Nimbus.

## From source

Install [go](https://go.dev/doc/install) version 1.20 or later.

Clone the repository:

```shell
git clone https://github.com/5GSEC/nimbus.git
cd nimbus
```

Install CRDs:

```shell
make install
```

Run the operator:

```shell
make run
```

## From Helm Chart

Follow [this](../deployments/nimbus/Readme.md) guide to install `nimbus` operator.

# Adapters

Just like Nimbus, there are various ways of installing Security engine adapters.

## nimbus-kubearmor

> [!Note]
> The `nimbus-kubearmor` adapter leverages the [KubeArmor](https://kubearmor.io) security engine for its functionality.
> To use this adapter, you'll need KubeArmor installed. Please
> follow [this](https://github.com/kubearmor/KubeArmor/blob/main/getting-started/deployment_guide.md) guide for
> installation.
> Creating a KubeArmorPolicy resource without KubeArmor will have no effect.

### From source

Clone the repository:

```shell
git clone https://github.com/5GSEC/nimbus.git
```

Go to nimbus-kubearmor directory:

```shell
cd nimbus/pkg/adapter/nimbus-kubearmor
```

Run `nimbus-kubearmor` adapter:

```shell
make run
```

### From Helm Chart

Follow [this](../deployments/nimbus-kubearmor/Readme.md) guide to install `nimbus-kubearmor` adapter.

## nimbus-netpol

> [!Note]
> The `nimbus-netpol` adapter leverages
> the [network plugin](https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/network-plugins/).
> To use network policies, you must be using a networking solution which supports NetworkPolicy. Creating a
> NetworkPolicy resource without a controller that implements it will have no effect.

### From source

Clone the repository:

```shell
git clone https://github.com/5GSEC/nimbus.git
```

Go to nimbus-netpol directory:

```shell
cd nimbus/pkg/adapter/nimbus-netpol
```

Run `nimbus-netpol` adapter:

```shell
make run
```

### From Helm Chart

Follow [this](../deployments/nimbus-netpol/Readme.md) guide to install `nimbus-netpol` adapter.

## nimbus-kyverno

> [!Note]
> The `nimbus-kyverno` adapter leverages the [Kyverno](https://kyverno.io/) security engine for its functionality.
> To use this adapter, you'll need Kyverno installed. Please
> follow [this](https://kyverno.io/docs/installation/methods/) guide for
> installation.
> Creating a KyvernoPolicy resource without Kyverno will have no effect.

### From source

Clone the repository:

```shell
git clone https://github.com/5GSEC/nimbus.git
```

Go to nimbus-kyverno directory:

```shell
cd nimbus/pkg/adapter/nimbus-kyverno
```

Run `nimbus-kyverno` adapter:

```shell
make run
```

### From Helm Chart

Follow [this](../deployments/nimbus-kyverno/Readme.md) guide to install `nimbus-kyverno` adapter.

