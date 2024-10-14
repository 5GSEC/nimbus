## Objective

The coco-workload intent likely aims to enhance security by ensuring that sensitive workloads are executed in environments that provide confidentiality and isolation. This could involve leveraging technologies like Confidential VMs, which are designed to protect data in use, thereby reducing the risk of data exposure or leakage.


**Note** : For the escapeToHost intent one needs to have either [nimbus-kyverno](../../deployments/nimbus-kyverno/Readme.md) adapter running in their cluster. To install the complete suite with all the adapters pls follow the steps mentioned [here](../getting-started.md#nimbus)

## Policy Creation

### Kyverno Policy

#### Prereq

- K8s cluster nodes need to have nested virtualization enabled for the confidential containers intent. Additionally kvm needs to be installed ([ubuntu-kvm](https://help.ubuntu.com/community/KVM/Installation)). 

- One should have [ConfidentialContainers](../getting-started.md#confidential-containers) runner installed in their cluster.

#### Policy Description

- The policy is designed to operate during the admission phase (admission: true), meaning it will enforce rules when workloads (like Deployments) are created. The background: true setting indicates that the policy can also apply to existing resources in the background, ensuring compliance over time. Apply on existing resource means that the policy will can generate policy reports for the resources which are ommitting the compliance defined by the policy.

- The key action in this policy is to mutate the workload by adding a runtimeClassName: kata-clh to the Deployment's spec. This is crucial because kata-clh likely refers to a runtime class configured to use Confidential VMs. By ensuring that the workload runs under this runtime, the policy enforces that the deployment is secured within a Confidential VM. User can apply any runtimeClassName by specifying it as a intent param.


   ```
    params:
      runtimeClass: ["kata-qemu"]
   ```


