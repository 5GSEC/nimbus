# Quick Tutorials

## Create a sample deployment

```shell
kubectl apply -f ./test/env/nginx-deploy.yaml
deployment.apps/nginx created
```

## Run Nimbus Operator

```shell
$ make run
test -s /Users/anurag/workspace/nimbus/bin/controller-gen && /Users/anurag/workspace/nimbus/bin/controller-gen --version | grep -q v0.13.0 || \
        GOBIN=/Users/anurag/workspace/nimbus/bin go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.13.0
/Users/anurag/workspace/nimbus/bin/controller-gen rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases
/Users/anurag/workspace/nimbus/bin/controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./api/..."
go fmt ./...
go vet ./...
go run cmd/main.go
2024-01-13T22:12:20+05:30       INFO    setup   Starting manager
2024-01-13T22:12:20+05:30       INFO    starting server {"kind": "health probe", "addr": "[::]:8081"}
2024-01-13T22:12:20+05:30       INFO    controller-runtime.metrics      Starting metrics server
2024-01-13T22:12:20+05:30       INFO    controller-runtime.metrics      Serving metrics server  {"bindAddress": ":8080", "secure": false}
2024-01-13T22:12:20+05:30       INFO    Starting EventSource    {"controller": "clustersecurityintentbinding", "controllerGroup": "intent.security.nimbus.com", "controllerKind": "ClusterSecurityIntentBinding", "source": "kind source: *v1.ClusterSecurityIntentBinding"}
2024-01-13T22:12:20+05:30       INFO    Starting EventSource    {"controller": "securityintentbinding", "controllerGroup": "intent.security.nimbus.com", "controllerKind": "SecurityIntentBinding", "source": "kind source: *v1.SecurityIntentBinding"}
2024-01-13T22:12:20+05:30       INFO    Starting EventSource    {"controller": "securityintentbinding", "controllerGroup": "intent.security.nimbus.com", "controllerKind": "SecurityIntentBinding", "source": "kind source: *v1.NimbusPolicy"}
2024-01-13T22:12:20+05:30       INFO    Starting Controller     {"controller": "securityintentbinding", "controllerGroup": "intent.security.nimbus.com", "controllerKind": "SecurityIntentBinding"}
2024-01-13T22:12:20+05:30       INFO    Starting EventSource    {"controller": "clustersecurityintentbinding", "controllerGroup": "intent.security.nimbus.com", "controllerKind": "ClusterSecurityIntentBinding", "source": "kind source: *v1.ClusterNimbusPolicy"}
2024-01-13T22:12:20+05:30       INFO    Starting Controller     {"controller": "clustersecurityintentbinding", "controllerGroup": "intent.security.nimbus.com", "controllerKind": "ClusterSecurityIntentBinding"}
2024-01-13T22:12:20+05:30       INFO    Starting EventSource    {"controller": "securityintent", "controllerGroup": "intent.security.nimbus.com", "controllerKind": "SecurityIntent", "source": "kind source: *v1.SecurityIntent"}
2024-01-13T22:12:20+05:30       INFO    Starting Controller     {"controller": "securityintent", "controllerGroup": "intent.security.nimbus.com", "controllerKind": "SecurityIntent"}
2024-01-13T22:12:20+05:30       INFO    Starting workers        {"controller": "securityintent", "controllerGroup": "intent.security.nimbus.com", "controllerKind": "SecurityIntent", "worker count": 1}
2024-01-13T22:12:20+05:30       INFO    Starting workers        {"controller": "clustersecurityintentbinding", "controllerGroup": "intent.security.nimbus.com", "controllerKind": "ClusterSecurityIntentBinding", "worker count": 1}
2024-01-13T22:12:20+05:30       INFO    Starting workers        {"controller": "securityintentbinding", "controllerGroup": "intent.security.nimbus.com", "controllerKind": "SecurityIntentBinding", "worker count": 1}
```

## Run Adapters

### KubeArmor

> [!Note]
> The `nimbus-kubearmor` adapter leverages the [KubeArmor](https://kubearmor.io) security engine for its functionality.
> To use this adapter, you'll need KubeArmor installed. Please
> follow [this](https://github.com/kubearmor/KubeArmor/blob/main/getting-started/deployment_guide.md) guide for
> installation.

```shell
$ cd pkg/adapter/nimbus-kubearmor
$ make run
{"level":"info","ts":"2024-01-13T22:13:25+05:30","msg":"KubeArmor Adapter started"}
{"level":"info","ts":"2024-01-13T22:13:25+05:30","msg":"NimbusPolicy watcher started"}
```

### Network Policy

> [!Note]
> The `nimbus-netpol` adapter leverages
> the [network plugin](https://kubernetes.io/docs/concepts/extend-kubernetes/compute-storage-net/network-plugins/).
> To use network policies, you must be using a networking solution which supports NetworkPolicy.

```shell
$ cd pkg/adapter/nimbus-netpol
$ make run
{"level":"info","ts":"2024-01-23T17:20:46+05:30","msg":"Network Policy adapter started"}
{"level":"info","ts":"2024-01-23T17:20:46+05:30","msg":"NimbusPolicy watcher started"}
{"level":"info","ts":"2024-01-23T17:20:46+05:30","msg":"ClusterNimbusPolicy watcher started"}
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
{"level":"info","ts":"2024-01-13T22:13:57+05:30","msg":"KubeArmor does not support this ID","ID":"dnsManipulation","NimbusPolicy":"multiple-sis-nsscoped-binding","NimbusPolicy.Namespace":"default"}
{"level":"info","ts":"2024-01-13T22:13:57+05:30","msg":"KubeArmorPolicy Created","KubeArmorPolicy.Name":"multiple-sis-nsscoped-binding-swdeploymenttools","KubeArmorPolicy.Namespace":"default"}
{"level":"info","ts":"2024-01-13T22:13:57+05:30","msg":"KubeArmorPolicy Created","KubeArmorPolicy.Name":"multiple-sis-nsscoped-binding-unauthorizedsatokenaccess","KubeArmorPolicy.Namespace":"default"}
```

You can also review the policies that were successfully generated:

```shell
$ kubectl get kubearmorpolicy
NAME                                                      AGE
multiple-sis-nsscoped-binding-swdeploymenttools           2m8s
multiple-sis-nsscoped-binding-unauthorizedsatokenaccess   2m8s
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
  creationTimestamp: "2024-01-23T12:05:54Z"
  generation: 1
  name: multiple-sis-nsscoped-binding-swdeploymenttools
  namespace: default
  ownerReferences:
    - apiVersion: intent.security.nimbus.com/v1
      blockOwnerDeletion: true
      controller: true
      kind: NimbusPolicy
      name: multiple-sis-nsscoped-binding
      uid: 2e634795-0e4d-4172-9d1d-bf783e6bc1c6
  resourceVersion: "550197"
  uid: 22f38fe4-3e71-437d-93e8-8eb517a12ad1
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
  creationTimestamp: "2024-01-23T12:05:54Z"
  generation: 1
  name: multiple-sis-nsscoped-binding-unauthorizedsatokenaccess
  namespace: default
  ownerReferences:
    - apiVersion: intent.security.nimbus.com/v1
      blockOwnerDeletion: true
      controller: true
      kind: NimbusPolicy
      name: multiple-sis-nsscoped-binding
      uid: 2e634795-0e4d-4172-9d1d-bf783e6bc1c6
  resourceVersion: "550198"
  uid: 8ac4bf6f-d543-4dad-9c9d-c2dc96f53925
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

### NetworkPolicy

Network Policy adapter logs that detected NimbusPolicy is shown below:

```shell
...
...
{"level":"info","ts":"2024-01-23T17:26:24+05:30","msg":"Network Policy adapter does not support this ID","ID":"swDeploymentTools","NimbusPolicy.Name":"multiple-sis-nsscoped-binding","NimbusPolicy.Namespace":"default"}
{"level":"info","ts":"2024-01-23T17:26:24+05:30","msg":"Network Policy adapter does not support this ID","ID":"unAuthorizedSaTokenAccess","NimbusPolicy.Name":"multiple-sis-nsscoped-binding","NimbusPolicy.Namespace":"default"}
{"level":"info","ts":"2024-01-23T17:26:24+05:30","msg":"NetworkPolicy created","NetworkPolicy.Name":"multiple-sis-nsscoped-binding-dnsmanipulation","NetworkPolicy.Namespace":"default"}
```

You can also review the network policies that were successfully generated:

```shell
$ kubectl get networkpolicy
NAME                                            POD-SELECTOR   AGE
multiple-sis-nsscoped-binding-dnsmanipulation   app=nginx      3m44s
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
  creationTimestamp: "2024-01-23T11:56:24Z"
  generation: 1
  name: multiple-sis-nsscoped-binding-dnsmanipulation
  namespace: default
  ownerReferences:
    - apiVersion: intent.security.nimbus.com/v1
      blockOwnerDeletion: true
      controller: true
      kind: NimbusPolicy
      name: multiple-sis-nsscoped-binding
      uid: a151ee11-539f-4dad-92ae-9a813a681790
  resourceVersion: "549724"
  uid: 8018a181-d317-418f-a700-d41369235701
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
{"level":"info","ts":"2024-01-23T17:40:51+05:30","msg":"KubeArmorPolicy already deleted, no action needed","KubeArmorPolicy.Name":"multiple-sis-nsscoped-binding-swdeploymenttools","KubeArmorPolicy.Namespace":"default"}
{"level":"info","ts":"2024-01-23T17:40:51+05:30","msg":"KubeArmorPolicy already deleted, no action needed","KubeArmorPolicy.Name":"multiple-sis-nsscoped-binding-unauthorizedsatokenaccess","KubeArmorPolicy.Namespace":"default"}
{"level":"info","ts":"2024-01-23T17:40:51+05:30","msg":"KubeArmorPolicy already deleted, no action needed","KubeArmorPolicy.Name":"multiple-sis-nsscoped-binding-dnsmanipulation","KubeArmorPolicy.Namespace":"default"}
```

* Check Network Policy adapter logs:

```shell
...
...
{"level":"info","ts":"2024-01-23T17:33:28+05:30","msg":"Network Policy adapter does not support this ID","ID":"swDeploymentTools","NimbusPolicy.Name":"multiple-sis-nsscoped-binding","NimbusPolicy.Namespace":"default"}
{"level":"info","ts":"2024-01-23T17:33:28+05:30","msg":"Network Policy adapter does not support this ID","ID":"unAuthorizedSaTokenAccess","NimbusPolicy.Name":"multiple-sis-nsscoped-binding","NimbusPolicy.Namespace":"default"}
{"level":"info","ts":"2024-01-23T17:33:28+05:30","msg":"NetworkPolicy already deleted, no action needed","NetworkPolicy.Name":"multiple-sis-nsscoped-binding-dnsmanipulation","NetworkPolicy.Namespace":"default"}
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