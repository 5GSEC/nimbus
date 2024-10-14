# Getting Started

This guide walks users through the steps to easily install and run the Nimbus operator. Each step includes the commands
needed and their descriptions to help users understand and proceed with each step.

# Prerequisites

Before you begin, set up the following:

- [kubectl](https://kubernetes.io/docs/tasks/tools/#kubectl) version 1.26 or later.
- A Kubernetes cluster running version 1.26 or later.
- Make sure that the bpf-lsm module needs is installed ([bpf-lsm](https://docs.kubearmor.io/kubearmor/documentation/faq#how-to-enable-kubearmorhostpolicy-for-k8s-cluster)).
- The Kubernetes clusters should be configured with a CNI that supports network policy.
  - For kind clusters, this reference ([calico-kind](https://docs.tigera.io/calico/latest/getting-started/kubernetes/kind)) has the details.
  - For on-prem clusters, this reference ([calico-onprem](https://docs.tigera.io/calico/latest/getting-started/kubernetes/self-managed-onprem/onpremises)) has the details.
  - For AWS EKS clusters, the VPC CNI supports kubernetes network policies ([vpc-cni-policy](https://aws.amazon.com/blogs/containers/amazon-vpc-cni-now-supports-kubernetes-network-policies/)).
- K8s cluster nodes need to have nested virtualization enabled for the confidential containers intent. Additionally kvm needs to be installed ([ubuntu-kvm](https://help.ubuntu.com/community/KVM/Installation)). 
  - For GCP VMs, nested virtualization can be enabled at create time using below command. The machine types which support nested virtualization are listed here ([cpu-virt](https://cloud.google.com/compute/docs/machine-resource#machine_type_comparison)).
```
export VM_MACHINE=n2-standard-16
export VM_IMAGE=ubuntu-2204-jammy-v20240614
export VM_IM_PROJ=ubuntu-os-cloud
gcloud compute instances create $VM_NAME --zone=$VM_ZONE --machine-type=$VM_MACHINE --image=$VM_IMAGE --image-project=$VM_IM_PROJ --boot-disk-size="200GB" --enable-nested-virtualization
```
  - For AWS, bare metal instances should be used as worker nodes in EKS, as nested virtualization cannot be enabled on standard EC2 instances ([aws-kata](https://aws.amazon.com/blogs/containers/enhancing-kubernetes-workload-isolation-and-security-using-kata-containers/)).


# Nimbus
## From Helm Chart

Follow [this](../deployments/nimbus/Readme.md) guide to install `nimbus` operator. By default the install of the `nimbus` operator installs the adapters also, and all the security engines - except confidential containers - too.

# Confidential Containers

You need to enable Confidential Containers in the Kubernetes cluster using the Confidential Containers Operator.

- Set a label on atleast one node 
```
$ kubectl get nodes
NAME                              STATUS   ROLES           AGE   VERSION
regional-md-0-t5clc-kh5ks-v429s   Ready    <none>          10d   v1.26.3
regional-tq6q6-6l6mv              Ready    control-plane   10d   v1.26.3

$ kubectl label node regional-md-0-t5clc-kh5ks-v429s node.kubernetes.io/worker=
```

- Deploy the operator. 
```
export RELEASE_VERSION="v0.8.0"
kubectl apply -k "github.com/confidential-containers/operator/config/release?ref=${RELEASE_VERSION}"
```

- Wait until each pod has status of running
```
kubectl get pods -n confidential-containers-system --watch
```

- Check that the crd is created.
```
$ kubectl get crd | grep ccruntime
ccruntimes.confidentialcontainers.org                      2024-06-28T08:04:02Z
```

- Create a custom resource
```
kubectl apply -k github.com/confidential-containers/operator/config/samples/ccruntime/default?ref=${RELEASE_VERSION}
kubectl get pods -n confidential-containers-system --watch
NAME                                              READY   STATUS    RESTARTS   AGE
cc-operator-controller-manager-857f844f7d-9pfbp   2/2     Running   0          10d
cc-operator-daemon-install-qd9ll                  1/1     Running   0          10d
cc-operator-pre-install-daemon-ccngm              1/1     Running   0          10d
```

- Check that the runtimeclass is created
```
$ kubectl get runtimeclass

NAME            HANDLER         AGE
kata            kata-qemu       10d
kata-clh        kata-clh        10d
kata-clh-tdx    kata-clh-tdx    10d
kata-qemu       kata-qemu       10d
kata-qemu-sev   kata-qemu-sev   10d
kata-qemu-snp   kata-qemu-snp   10d
kata-qemu-tdx   kata-qemu-tdx   10d
```
