# Quick Tutorials

## Create a sample deployment

```shell
kubectl apply -f ./test/env/nginx-deploy.yaml
deployment.apps/nginx created
```

## Install Nimbus Operator

Follow [this](../deployments/nimbus/Readme.md) guide to install `nimbus` operator.

## Run Adapters

### KubeArmor

> [!Note]
> The `nimbus-kubearmor` adapter leverages the [KubeArmor](https://kubearmor.io) security engine for its functionality.
> To use this adapter, you'll need KubeArmor installed. Please
> follow [this](https://github.com/kubearmor/KubeArmor/blob/main/getting-started/deployment_guide.md) guide for
> installation.
> Creating a KubeArmorPolicy resource without KubeArmor will have no effect.

Follow [this](../deployments/nimbus-kubearmor/Readme.md) guide to install `nimbus-kubearmor` adapter.

Open a new terminal and execute following command to check logs:

```shell
$ kubectl -n nimbus logs -f deploy/nimbus-kubearmor
{"level":"info","ts":"2024-01-31T14:55:11+05:30","msg":"KubeArmor adapter started"}
{"level":"info","ts":"2024-01-31T14:55:11+05:30","msg":"ClusterNimbusPolicy watcher started"}
{"level":"info","ts":"2024-01-31T14:55:11+05:30","msg":"NimbusPolicy watcher started"}
```

### Network Policy

> [!Note]
> The `nimbus-netpol` adapter leverages
> the [network plugin](https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/network-plugins/).
> To use network policies, you must be using a networking solution which supports NetworkPolicy. Creating a
> NetworkPolicy resource without a controller that implements it will have no effect.


Follow [this](../deployments/nimbus-netpol/Readme.md) guide to install `nimbus-netpol` adapter.

Open a new terminal and execute following command to check logs:

```shell
$ kubectl -n nimbus logs -f deploy/nimbus-netpol
{"level":"info","ts":"2024-01-31T14:53:36+05:30","msg":"NimbusPolicy watcher started"}
{"level":"info","ts":"2024-01-31T14:53:36+05:30","msg":"ClusterNimbusPolicy watcher started"}
{"level":"info","ts":"2024-01-31T14:53:36+05:30","msg":"Network Policy adapter started"}
```

## Create SecurityIntent and SecurityIntentBinding

```shell
$ kubectl apply -f ./test/v2/namespaced/multiple-si-sib-namespaced.yaml
securityintent.intent.security.nimbus.com/pkg-mgr-exec-multiple-nsscoped created
securityintent.intent.security.nimbus.com/unauthorized-sa-token-access-multiple-nsscoped created
securityintent.intent.security.nimbus.com/dns-manipulation-multiple-nsscoped created
securityintentbinding.intent.security.nimbus.com/multiple-sis-nsscoped-binding created
```

## Verify SecurityIntent and SecurityIntentBinding

* Verify SecurityIntent

```shell
$ kubectl get securityintent
NAME                                             STATUS
pkg-mgr-exec-multiple-nsscoped                   Created
unauthorized-sa-token-access-multiple-nsscoped   Created
dns-manipulation-multiple-nsscoped               Created
```

* Verify SecurityIntentBinding

```shell
$ kubectl get securityintentbinding
NAME                            STATUS
multiple-sis-nsscoped-binding   Created
```

## Verify the Security Engines policies

### KubeArmorPolicy

KubeArmor adapter logs that detected NimbusPolicy is shown below:

```shell
...
...
{"level":"info","ts":"2024-01-31T14:55:18+05:30","msg":"NimbusPolicy found","NimbusPolicy.Name":"multiple-sis-nsscoped-binding","NimbusPolicy.Namespace":"default"}
{"level":"info","ts":"2024-01-31T14:55:19+05:30","msg":"KubeArmorPolicy Created","KubeArmorPolicy.Name":"multiple-sis-nsscoped-binding-swdeploymenttools","KubeArmorPolicy.Namespace":"default"}
{"level":"info","ts":"2024-01-31T14:55:19+05:30","msg":"KubeArmorPolicy Created","KubeArmorPolicy.Name":"multiple-sis-nsscoped-binding-unauthorizedsatokenaccess","KubeArmorPolicy.Namespace":"default"}
{"level":"info","ts":"2024-01-31T14:55:19+05:30","msg":"KubeArmorPolicy Created","KubeArmorPolicy.Name":"multiple-sis-nsscoped-binding-dnsmanipulation","KubeArmorPolicy.Namespace":"default"}
```

You can also review the policies that were successfully generated:

```shell
$ kubectl get kubearmorpolicy
NAME                                                      AGE
multiple-sis-nsscoped-binding-swdeploymenttools           2m
multiple-sis-nsscoped-binding-unauthorizedsatokenaccess   2m
multiple-sis-nsscoped-binding-dnsmanipulation             2m
```

Or, inspect each individual policy for detailed info:

```shell
$ kubectl get kubearmorpolicy multiple-sis-nsscoped-binding-swdeploymenttools -o yaml
```

```yaml
apiVersion: security.kubearmor.com/v1
kind: KubeArmorPolicy
metadata:
  annotations:
    app.kubernetes.io/managed-by: nimbus-kubearmor
  creationTimestamp: "2024-01-31T09:25:19Z"
  generation: 1
  name: multiple-sis-nsscoped-binding-swdeploymenttools
  namespace: default
  ownerReferences:
    - apiVersion: intent.security.nimbus.com/v1
      blockOwnerDeletion: true
      controller: true
      kind: NimbusPolicy
      name: multiple-sis-nsscoped-binding
      uid: d2176ea3-3e0b-4671-8f58-dbff376d87b0
  resourceVersion: "594438"
  uid: 363d5191-20b9-471e-80c2-a142f8396e13
spec:
  action: Block
  capabilities: { }
  file: { }
  message: Do not allow the execution of package managers inside the containers
  network: { }
  process:
    matchPaths:
      - path: /usr/bin/apt
      - path: /usr/bin/apt-get
      - path: /bin/apt-get
      - path: /bin/apt
      - path: /usr/bin/dpkg
      - path: /bin/dpkg
      - path: /usr/bin/gdebi
      - path: /bin/gdebi
      - path: /usr/bin/make
      - path: /bin/make
      - path: /usr/bin/yum
      - path: /bin/yum
      - path: /usr/bin/rpm
      - path: /bin/rpm
      - path: /usr/bin/dnf
      - path: /bin/dnf
      - path: /usr/bin/pacman
      - path: /usr/sbin/pacman
      - path: /bin/pacman
      - path: /sbin/pacman
      - path: /usr/bin/makepkg
      - path: /usr/sbin/makepkg
      - path: /bin/makepkg
      - path: /sbin/makepkg
      - path: /usr/bin/yaourt
      - path: /usr/sbin/yaourt
      - path: /bin/yaourt
      - path: /sbin/yaourt
      - path: /usr/bin/zypper
      - path: /bin/zypper
    severity: 5
  selector:
    matchLabels:
      app: nginx
  syscalls: { }
  tags:
    - NIST
    - CM-7(5)
    - SI-4
    - Package Manager
```

```shell
$ kubectl get kubearmorpolicy multiple-sis-nsscoped-binding-unauthorizedsatokenaccess -o yaml
```

```yaml
apiVersion: security.kubearmor.com/v1
kind: KubeArmorPolicy
metadata:
  annotations:
    app.kubernetes.io/managed-by: nimbus-kubearmor
  creationTimestamp: "2024-01-31T09:25:19Z"
  generation: 1
  name: multiple-sis-nsscoped-binding-unauthorizedsatokenaccess
  namespace: default
  ownerReferences:
    - apiVersion: intent.security.nimbus.com/v1
      blockOwnerDeletion: true
      controller: true
      kind: NimbusPolicy
      name: multiple-sis-nsscoped-binding
      uid: d2176ea3-3e0b-4671-8f58-dbff376d87b0
  resourceVersion: "594439"
  uid: 166b1193-751c-4b6b-acbd-a68ed1dd26e8
spec:
  action: Block
  capabilities: { }
  file:
    matchDirectories:
      - dir: /run/secrets/kubernetes.io/serviceaccount/
        recursive: true
  network: { }
  process: { }
  selector:
    matchLabels:
      app: nginx
  syscalls: { }
```

```shell
$ kubectl get kubearmorpolicy multiple-sis-nsscoped-binding-dnsmanipulation -o yaml
```

```yaml
apiVersion: security.kubearmor.com/v1
kind: KubeArmorPolicy
metadata:
  annotations:
    app.kubernetes.io/managed-by: nimbus-kubearmor
  creationTimestamp: "2024-01-31T09:25:19Z"
  generation: 1
  name: multiple-sis-nsscoped-binding-dnsmanipulation
  namespace: default
  ownerReferences:
    - apiVersion: intent.security.nimbus.com/v1
      blockOwnerDeletion: true
      controller: true
      kind: NimbusPolicy
      name: multiple-sis-nsscoped-binding
      uid: d2176ea3-3e0b-4671-8f58-dbff376d87b0
  resourceVersion: "594440"
  uid: cbce8ea8-988d-4033-9d9d-c597acbe496a
spec:
  action: Block
  capabilities: { }
  file:
    matchPaths:
      - path: /etc/resolv.conf
        readOnly: true
  network: { }
  process: { }
  selector:
    matchLabels:
      app: nginx
  syscalls: { }
```

### NetworkPolicy

Network Policy adapter logs that detected NimbusPolicy is shown below:

```shell
...
...
{"level":"info","ts":"2024-01-31T14:55:18+05:30","msg":"NimbusPolicy found","NimbusPolicy.Name":"multiple-sis-nsscoped-binding","NimbusPolicy.Namespace":"default"}
{"level":"info","ts":"2024-01-31T14:55:18+05:30","msg":"Network Policy adapter does not support this ID","ID":"swDeploymentTools","NimbusPolicy.Name":"multiple-sis-nsscoped-binding","NimbusPolicy.Namespace":"default"}
{"level":"info","ts":"2024-01-31T14:55:18+05:30","msg":"Network Policy adapter does not support this ID","ID":"unAuthorizedSaTokenAccess","NimbusPolicy.Name":"multiple-sis-nsscoped-binding","NimbusPolicy.Namespace":"default"}
{"level":"info","ts":"2024-01-31T14:55:18+05:30","msg":"NetworkPolicy created","NetworkPolicy.Name":"multiple-sis-nsscoped-binding-dnsmanipulation","NetworkPolicy.Namespace":"default"}
```

You can also review the network policies that were successfully generated:

```shell
$ kubectl get networkpolicy
NAME                                            POD-SELECTOR   AGE
multiple-sis-nsscoped-binding-dnsmanipulation   app=nginx      5m6s
```

Or, inspect policy for detailed info:

```shell
$ kubectl get networkpolicy multiple-sis-nsscoped-binding-dnsmanipulation -o yaml
```

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  annotations:
    app.kubernetes.io/managed-by: nimbus-netpol
  creationTimestamp: "2024-01-31T09:25:18Z"
  generation: 1
  name: multiple-sis-nsscoped-binding-dnsmanipulation
  namespace: default
  ownerReferences:
    - apiVersion: intent.security.nimbus.com/v1
      blockOwnerDeletion: true
      controller: true
      kind: NimbusPolicy
      name: multiple-sis-nsscoped-binding
      uid: d2176ea3-3e0b-4671-8f58-dbff376d87b0
  resourceVersion: "594436"
  uid: 5d7743e6-7dfd-4d3e-b503-6c43bea4473d
spec:
  egress:
    - ports:
        - port: 53
          protocol: UDP
        - port: 53
          protocol: TCP
      to:
        - namespaceSelector:
            matchLabels:
              kubernetes.io/metadata.name: kube-system
          podSelector:
            matchLabels:
              k8s-app: kube-dns
  podSelector:
    matchLabels:
      app: nginx
  policyTypes:
    - Egress
```

## Cleanup

* The SecurityIntent and SecurityIntentBinding created earlier are no longer needed and can be deleted:

```shell
$ kubectl delete -f ./test/v2/namespaced/multiple-si-sib-namespaced.yaml
securityintent.intent.security.nimbus.com "pkg-mgr-exec-multiple-nsscoped" deleted
securityintent.intent.security.nimbus.com "unauthorized-sa-token-access-multiple-nsscoped" deleted
securityintent.intent.security.nimbus.com "dns-manipulation-multiple-nsscoped" deleted
securityintentbinding.intent.security.nimbus.com "multiple-sis-nsscoped-binding" deleted
```

* Check KubeArmor Security Engine adapter logs:

```shell
...
...
{"level":"info","ts":"2024-01-31T15:01:09+05:30","msg":"NimbusPolicy deleted","NimbusPolicy.Name":"multiple-sis-nsscoped-binding","NimbusPolicy.Namespace":"default"}
{"level":"info","ts":"2024-01-31T15:01:10+05:30","msg":"KubeArmorPolicy already deleted, no action needed","KubeArmorPolicy.Name":"multiple-sis-nsscoped-binding-swdeploymenttools","KubeArmorPolicy.Namespace":"default"}
{"level":"info","ts":"2024-01-31T15:01:10+05:30","msg":"KubeArmorPolicy already deleted, no action needed","KubeArmorPolicy.Name":"multiple-sis-nsscoped-binding-unauthorizedsatokenaccess","KubeArmorPolicy.Namespace":"default"}
{"level":"info","ts":"2024-01-31T15:01:10+05:30","msg":"KubeArmorPolicy already deleted, no action needed","KubeArmorPolicy.Name":"multiple-sis-nsscoped-binding-dnsmanipulation","KubeArmorPolicy.Namespace":"default"}
```

* Check Network Policy adapter logs:

```shell
...
...
{"level":"info","ts":"2024-01-31T15:01:09+05:30","msg":"NimbusPolicy deleted","NimbusPolicy.Name":"multiple-sis-nsscoped-binding","NimbusPolicy.Namespace":"default"}
{"level":"info","ts":"2024-01-31T15:01:09+05:30","msg":"Network Policy adapter does not support this ID","ID":"swDeploymentTools","NimbusPolicy.Name":"multiple-sis-nsscoped-binding","NimbusPolicy.Namespace":"default"}
{"level":"info","ts":"2024-01-31T15:01:09+05:30","msg":"Network Policy adapter does not support this ID","ID":"unAuthorizedSaTokenAccess","NimbusPolicy.Name":"multiple-sis-nsscoped-binding","NimbusPolicy.Namespace":"default"}
{"level":"info","ts":"2024-01-31T15:01:09+05:30","msg":"NetworkPolicy already deleted, no action needed","NetworkPolicy.Name":"multiple-sis-nsscoped-binding-dnsmanipulation","NetworkPolicy.Namespace":"default"}
```

* Delete deployment

```shell
$ kubectl delete -f ./test/env/nginx-deploy.yaml
deployment.apps "nginx" deleted
```

* Confirm all resources have been deleted (Optional)

```shell
$ kubectl get securityintent,securityintentbinding,kubearmorpolicy,netpol -A
No resources found
```