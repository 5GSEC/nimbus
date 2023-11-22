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

## Installation
### 1. Clone Nimbus source code:
```
$ git clone https://git.cclab-inu.com/b0m313/nimbus.git
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

### 3. Install Make
```
$ sudo apt-get update
$ sudo apt-get install -y make
```

### 4. Install Golang
```
$ wget https://golang.org/dl/go1.21.3.linux-amd64.tar.gz
$ sudo tar -C /usr/bin -xzf go1.21.3.linux-amd64.tar.gz
$ export PATH=$PATH:/usr/bin/go/bin
$ source ~/.profile
$ go version
```

### 5. Install KubeBuilder
```
$ curl -L -o kubebuilder https://go.kubebuilder.io/dl/latest/$(go env GOOS)/$(go env GOARCH)
$ chmod +x kubebuilder
$ sudo mv ./kubebuilder /usr/local/bin/
$ kubebuilder version
```
    
### 6. Install Kustomize

```
$ curl -s "https://raw.githubusercontent.com/kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"  | bash
$ chmod +x kustomize
$ sudo mv ./kustomize /usr/local/bin/
$ kustomize version
```

## Running Nimbus

Commands to run Nimbus operators:

### 1. Generate code
```
$ cd nimbus
$ ~/nimbus$ pwd
/home/cclab/nimbus
$ ~/nimbus$ ls
api  bin  config  Dockerfile  Getting-Started.md  go.mod  go.sum  hack  internal  main.go  Makefile  PROJECT  Quick-tutorials.md  README.md  test-yaml
```
Generate the necessary code based on the API definition
    
```
$ make generate
```

### 2. Install CRD
Install Custom Resource Definitions in a Kubernetes Cluster
    
```
$ make install
```

### 3. Run Operators
Run the operator in your local environment
```
$ make run
```
<br><br>
📌 After completing these steps, the Nimbus operator is successfully installed and running.
