# Getting Started

This guide walks users through the steps to easily install and run the Nimbus operator. Each step includes the commands needed and their descriptions to help users understand and proceed with each step.

## Preparation

Before you begin, you'll need to set up the following preferences :

- containerd + K8S + Cilium + Kubearmor (3-node/standalone)
- Environment for using KubeBuilder
    - go version v1.19.0+
    - docker version 17.03+.
    - kubectl version v1.11.3+.
    - Access to a Kubernetes v1.11.3+ cluster.
	- make

## Installation
### 1. Clone Nimbus source code:
```
$ git clone https://github.com/5GSEC/nimbus.git
```

### 2. Install Kubearmor:
Install Kubearmor and related tools:<br>

```
$ curl -sfL http://get.kubearmor.io/ | sudo sh -s -- -b /usr/local/bin && karmor install
```

Install the Discovery Engine:<br>
```
$ curl -o discovery-engine.yaml https://raw.githubusercontent.com/kubearmor/discovery-engine/dev/deployments/k8s/deployment.yaml
$ kubectl apply -f discovery-engine.yaml
```

### 3. Install KubeBuilder
```
$ curl -L -o kubebuilder https://go.kubebuilder.io/dl/latest/$(go env GOOS)/$(go env GOARCH)
$ chmod +x kubebuilder
$ sudo mv ./kubebuilder /usr/local/bin/
$ kubebuilder version
```
    
### 4. Install Kustomize

```
$ curl -s "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"  | bash
$ chmod +x kustomize
$ sudo mv ./kustomize /usr/local/bin/
$ kustomize version
```

## Running Nimbus

Commands to run Nimbus operators:

### 1. Apply API group resources 
Apply API group resources 
    
```
$ make generate
```


### 2. Install CRD
Install Custom Resource Definitions in a Kubernetes Cluster
    
```
$ make install
```

📌  Steps 1 and 2 are required if you have a completely clean environment, as they allow the server to find the requested resources.

### 3. Run Operators
Run the operator in your local environment
```
$ make build
$ make run
```
<br><br>
📌 After completing these steps, the Nimbus operator is successfully installed and running.
