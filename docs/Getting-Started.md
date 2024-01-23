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

## Using helm chart

Follow [this](../deployments/nimbus/Readme.md) guide to install `nimbus` operator.

# Adapters

Just like Nimbus, there are various ways of installing Security engine adapters.

## nimbus-kubearmor

> [!Note]
> The `nimbus-kubearmor` adapter leverages the [KubeArmor](https://kubearmor.io) security engine for its functionality.
> To use this adapter, you'll need KubeArmor installed. Please
> follow [this](https://github.com/kubearmor/KubeArmor/blob/main/getting-started/deployment_guide.md) guide for
> installation.

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

### Using helm chart

Follow [this](../deployments/nimbus-kubearmor/Readme.md) guide to install `nimbus-kubearmor` adapter. 