## Objective

- The `pkg-mgr-execution` intent likely aims to prevent unauthorized or potentially harmful package management operations. This is critical in a Kubernetes environment, where package managers can be exploited by adversaries to install malicious software or manipulate existing applications.

**Note** : For the exploit-pfa intent one needs to have  [nimbus-kubearmor](../../deployments/nimbus-kubearmor/Readme.md) adapter running in their cluster. To install the complete suite with all the adapters pls follow the steps mentioned [here](../getting-started.md#nimbus)

## Policy Creation

The exploit-pfa intent results in `KubeArmorPolicy`. Below is the behaviour of intent in terms of policy: 

### KubeArmorPolicy

#### Prereq

- For the `KubeArmorPolicy` to work, one should have a [BPF-LSM](https://github.com/kubearmor/KubeArmor/blob/main/getting-started/FAQ.md#checking-and-enabling-support-for-bpf-lsm) enabled for each node in their cluster.

#### Policy Description

- The KubeArmorPolicy created here specifies that any attempt to execute certain package management commands will be blocked. This is a proactive security measure to prevent unauthorized changes to the system.

- By blocking execution of these critical pkg-mgmt tools, the policy significantly reduces the attack surface for the application. This prevents attackers from executing potentially malicious scripts or binaries that could lead to data breaches or further compromises.