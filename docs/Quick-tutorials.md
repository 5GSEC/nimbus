# Quick Tutorials

## Create a sample deployment

```shell
$ kubectl apply -f ./test/env/nginx-deploy.yaml
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

## Run Adapter (in this example, KubeArmor)

```shell
$ cd pkg/adapter/nimbus-kubearmor
$ make run
{"level":"info","ts":"2024-01-13T22:13:25+05:30","msg":"KubeArmor Adapter started"}
{"level":"info","ts":"2024-01-13T22:13:25+05:30","msg":"NimbusPolicy watcher started"}
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

* Verify Nimbus policy

```shell
$ kubectl get nimbuspolicy
NAME                            STATUS
multiple-sis-nsscoped-binding   Created
```

or inspect nimbuspolicy for detailed info:

```shell
$ kubectl get nimbuspolicy multiple-sis-nsscoped-binding -o yaml
```

```yaml
apiVersion: intent.security.nimbus.com/v1
kind: NimbusPolicy
metadata:
  creationTimestamp: "2024-01-13T16:43:56Z"
  generation: 1
  name: multiple-sis-nsscoped-binding
  namespace: default
  ownerReferences:
    - apiVersion: intent.security.nimbus.com/v1
      blockOwnerDeletion: true
      controller: true
      kind: SecurityIntentBinding
      name: multiple-sis-nsscoped-binding
      uid: b047d013-b402-4126-9798-529d96d2cc85
  resourceVersion: "406627"
  uid: 6ef05c5b-660f-4ba0-baa3-bbf87e501cca
spec:
  rules:
    - description: Do not allow the execution of package managers inside the containers
      id: swDeploymentTools
      rule:
        action: Block
        mode: Strict
    - id: unAuthorizedSaTokenAccess
      rule:
        action: Block
        mode: strict
    - id: dnsManipulation
      rule:
        action: Block
        mode: best-effort
  selector:
    matchLabels:
      app: nginx
status:
  status: Created
```

## Verify the Security Engine policy (in this example, KubeArmorPolicy)

KubeArmor adapter logs can that that detected NimbusPolicy is shown below:

```shell
{"level":"info","ts":"2024-01-13T22:13:25+05:30","msg":"KubeArmor Adapter started"}
{"level":"info","ts":"2024-01-13T22:13:25+05:30","msg":"NimbusPolicy watcher started"}
{"level":"info","ts":"2024-01-13T22:13:56+05:30","msg":"NimbusPolicy found","NimbusPolicy.Name":"multiple-sis-nsscoped-binding","NimbusPolicy.Namespace":"default"}
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
  creationTimestamp: "2024-01-13T16:43:57Z"
  generation: 1
  name: multiple-sis-nsscoped-binding-swdeploymenttools
  namespace: default
  resourceVersion: "406628"
  uid: b665ed3c-89de-40c4-bf24-1ac7e8ca63eb
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
  creationTimestamp: "2024-01-13T16:43:57Z"
  generation: 1
  name: multiple-sis-nsscoped-binding-unauthorizedsatokenaccess
  namespace: default
  resourceVersion: "406629"
  uid: 6644f0a9-46a2-4bde-9b5a-b01947da3311
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

## Cleanup

* The SecurityIntent and SecurityIntentBinding created earlier are no longer needed and can be deleted:

```shell
$ kubectl delete -f ./test/v2/namespaced/multiple-si-sib-namespaced.yaml
securityintent.intent.security.nimbus.com "pkg-mgr-exec-multiple-nsscoped" deleted
securityintent.intent.security.nimbus.com "unauthorized-sa-token-access-multiple-nsscoped" deleted
securityintent.intent.security.nimbus.com "dns-manipulation-multiple-nsscoped" deleted
securityintentbinding.intent.security.nimbus.com "multiple-sis-nsscoped-binding" deleted
```

* Check Security Engine adapter logs:

```shell
{"level":"info","ts":"2024-01-13T22:13:25+05:30","msg":"KubeArmor Adapter started"}
{"level":"info","ts":"2024-01-13T22:13:25+05:30","msg":"NimbusPolicy watcher started"}
{"level":"info","ts":"2024-01-13T22:13:56+05:30","msg":"NimbusPolicy found","NimbusPolicy.Name":"multiple-sis-nsscoped-binding","NimbusPolicy.Namespace":"default"}
{"level":"info","ts":"2024-01-13T22:13:57+05:30","msg":"KubeArmor does not support this ID","ID":"dnsManipulation","NimbusPolicy":"multiple-sis-nsscoped-binding","NimbusPolicy.Namespace":"default"}
{"level":"info","ts":"2024-01-13T22:13:57+05:30","msg":"KubeArmorPolicy Created","KubeArmorPolicy.Name":"multiple-sis-nsscoped-binding-swdeploymenttools","KubeArmorPolicy.Namespace":"default"}
{"level":"info","ts":"2024-01-13T22:13:57+05:30","msg":"KubeArmorPolicy Created","KubeArmorPolicy.Name":"multiple-sis-nsscoped-binding-unauthorizedsatokenaccess","KubeArmorPolicy.Namespace":"default"}
{"level":"info","ts":"2024-01-13T22:17:48+05:30","msg":"NimbusPolicy deleted","NimbusPolicy.Name":"multiple-sis-nsscoped-binding","NimbusPolicy.Namespace":"default"}
{"level":"info","ts":"2024-01-13T22:17:49+05:30","msg":"KubeArmor does not support this ID","ID":"dnsManipulation","NimbusPolicy":"multiple-sis-nsscoped-binding","NimbusPolicy.Namespace":"default"}
{"level":"info","ts":"2024-01-13T22:17:49+05:30","msg":"KubeArmorPolicy deleted due to NimbusPolicy deletion","KubeArmorPolicy.Name":"multiple-sis-nsscoped-binding-swdeploymenttools","KubeArmorPolicy.Namespace":"default","NimbusPolicy.Name":"multiple-sis-nsscoped-binding","NimbusPolicy.Namespace":"default"}
{"level":"info","ts":"2024-01-13T22:17:49+05:30","msg":"KubeArmorPolicy deleted due to NimbusPolicy deletion","KubeArmorPolicy.Name":"multiple-sis-nsscoped-binding-unauthorizedsatokenaccess","KubeArmorPolicy.Namespace":"default","NimbusPolicy.Name":"multiple-sis-nsscoped-binding","NimbusPolicy.Namespace":"default"}
```

* Delete deployment

```shell
$ kubectl delete -f ./test/env/nginx-deploy.yaml
deployment.apps "nginx" deleted
```

* Confirm all resources have been deleted (Optional)

```shell
$ kubectl get securityintent,securityintentbinding,nimbuspolicy,kubearmorpolicy -A
No resources found
```