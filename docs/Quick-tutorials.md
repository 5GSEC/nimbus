# Quick Tutorials

## Deploy test pod
```
$ kubectl apply -f ./test/env/multiubuntu.yaml
```

## Sample

### Run Operators (Nimbus)
```
$ make run
test -s /home/cclab/nimbus_accuknox/bin/controller-gen && /home/cclab/nimbus_accuknox/bin/controller-gen --version | grep -q v0.13.0 || \
GOBIN=/home/cclab/nimbus_accuknox/bin go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.13.0
/home/cclab/nimbus_accuknox/bin/controller-gen rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases
/home/cclab/nimbus_accuknox/bin/controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./pkg/..."
go fmt ./...
go vet ./...
go run cmd/main.go
2024-01-09T13:36:57Z    INFO    setup   Starting manager
2024-01-09T13:36:57Z    INFO    controller-runtime.metrics      Starting metrics server
2024-01-09T13:36:57Z    INFO    starting server {"kind": "health probe", "addr": "[::]:8081"}
2024-01-09T13:36:57Z    INFO    Starting EventSource    {"controller": "securityintent", "controllerGroup": "intent.security.nimbus.com", "controllerKind": "SecurityIntent", "source": "kind source: *v1.SecurityIntent"}
2024-01-09T13:36:57Z    INFO    Starting EventSource    {"controller": "nimbuspolicy", "controllerGroup": "intent.security.nimbus.com", "controllerKind": "NimbusPolicy", "source": "kind source: *v1.NimbusPolicy"}
...
```

### Run Adapter
```
$ cd /pk/nimbus-kubearmor
$ make build
$ make run
./nimbus-kubearmor
2024/01/09 13:36:18 Starting Kubernetes client configuration
2024/01/09 13:36:18 Starting NimbusPolicyWatcher
2024/01/09 13:36:18 Starting policy processing loop
```

### Create and apply Securityintent and SecurityintentBinding file
```
$ kubectl apply -f intents/system/intent-path-block.yaml
securityintent.intent.security.nimbus.com/group-1-proc-path-sleep-block created
```

```
$ kubectl apply -f bindings/system/binding-path-block.yaml
securityintentbinding.intent.security.nimbus.com/sys-proc-path-sleep-block created
```


### Verify SecurityIntent and SecurityIntentBinding
You can also check the operator's logs to see the detection and the process of creating the Nimbus Policy. 

```
$ make run
...
2024-01-09T13:37:06Z    INFO    SecurityIntent resource found   {"controller": "securityintent", "controllerGroup": "intent.security.nimbus.com", "controllerKind": "SecurityIntent", "SecurityIntent": {"name":"group-1-proc-path-sleep-block","namespace":"multiubuntu"}, "namespace": "multiubuntu", "name": "group-1-proc-path-sleep-block", "reconcileID": "5f7f67ea-33af-46b9-942a-af99a792c621", "Name": "group-1-proc-path-sleep-block", "Namespace": "multiubuntu"}
2024-01-09T13:37:19Z    INFO    SecurityIntentBinding resource found    {"controller": "securityintentbinding", "controllerGroup": "intent.security.nimbus.com", "controllerKind": "SecurityIntentBinding", "SecurityIntentBinding": {"name":"sys-proc-path-sleep-block","namespace":"multiubuntu"}, "namespace": "multiubuntu", "name": "sys-proc-path-sleep-block", "reconcileID": "6425366f-c6ca-4a73-87e1-1191d7984166", "Name": "sys-proc-path-sleep-block", "Namespace": "multiubuntu"}
2024-01-09T13:37:19Z    INFO    Starting intent and binding matching    {"controller": "securityintentbinding", "controllerGroup": "intent.security.nimbus.com", "controllerKind": "SecurityIntentBinding", "SecurityIntentBinding": {"name":"sys-proc-path-sleep-block","namespace":"multiubuntu"}, "namespace": "multiubuntu", "name": "sys-proc-path-sleep-block", "reconcileID": "6425366f-c6ca-4a73-87e1-1191d7984166"}
2024-01-09T13:37:19Z    INFO    Matching completed      {"controller": "securityintentbinding", "controllerGroup": "intent.security.nimbus.com", "controllerKind": "SecurityIntentBinding", "SecurityIntentBinding": {"name":"sys-proc-path-sleep-block","namespace":"multiubuntu"}, "namespace": "multiubuntu", "name": "sys-proc-path-sleep-block", "reconcileID": "6425366f-c6ca-4a73-87e1-1191d7984166", "Matched Intent Names": ["group-1-proc-path-sleep-block"], "Matched Binding Names": ["sys-proc-path-sleep-block"]}
2024-01-09T13:37:19Z    INFO    Starting NimbusPolicy building  {"controller": "securityintentbinding", "controllerGroup": "intent.security.nimbus.com", "controllerKind": "SecurityIntentBinding", "SecurityIntentBinding": {"name":"sys-proc-path-sleep-block","namespace":"multiubuntu"}, "namespace": "multiubuntu", "name": "sys-proc-path-sleep-block", "reconcileID": "6425366f-c6ca-4a73-87e1-1191d7984166"}
2024-01-09T13:37:19Z    INFO    NimbusPolicy built successfully {"controller": "securityintentbinding", "controllerGroup": "intent.security.nimbus.com", "controllerKind": "SecurityIntentBinding", "SecurityIntentBinding": {"name":"sys-proc-path-sleep-block","namespace":"multiubuntu"}, "namespace": "multiubuntu", "name": "sys-proc-path-sleep-block", "reconcileID": "6425366f-c6ca-4a73-87e1-1191d7984166", "Policy": {"namespace": "multiubuntu", "name": "sys-proc-path-sleep-block"}}
2024-01-09T13:37:19Z    INFO    Found: NimbusPolicy     {"controller": "nimbuspolicy", "controllerGroup": "intent.security.nimbus.com", "controllerKind": "NimbusPolicy", "NimbusPolicy": {"name":"sys-proc-path-sleep-block","namespace":"multiubuntu"}, "namespace": "multiubuntu", "name": "sys-proc-path-sleep-block", "reconcileID": "46b8482e-bd09-44d4-9cdc-6b9b8c17febf", "Name": "sys-proc-path-sleep-block", "Namespace": "multiubuntu"}
...
```

To verify that it was actually created, you can check the following. 
* Verify SecurityIntent
```
$ kubectl get SecurityIntent -n multiubuntu
NAME                            AGE
group-1-proc-path-sleep-block   28s
```
* Verify SecurityIntentBinding
```
$ kubectl get SecurityIntentBinding -n multiubuntu
NAME                        AGE
sys-proc-path-sleep-block   29s
```
* Verify Nimbus policy
```
$ kubectl get nimbuspolicy -n multiubuntu
NAME                        AGE
sys-proc-path-sleep-block   39s
```
```
$ kubectl get np -n multiubuntu sys-proc-path-sleep-block -o yaml
apiVersion: intent.security.nimbus.com/v1
kind: NimbusPolicy
metadata:
  creationTimestamp: "2024-01-09T13:37:19Z"
  generation: 1
  name: sys-proc-path-sleep-block
  namespace: multiubuntu
  resourceVersion: "5753517"
  uid: 5d2ae075-98b8-4958-850e-8114cb6dec19
spec:
  rules:
  - description: block the execution of '/bin/sleep'
    id: sys-proc-paths
    rule:
    - action: Block
      matchPaths:
      - path: /bin/sleep
  selector:
    matchLabels:
      group: group-1
```

### Verify the adapter 
The log for the adapter that detected nimbuspolicy is shown below. 
```
$ make run
./nimbus-kubearmor
2024/01/09 13:36:18 Starting Kubernetes client configuration
2024/01/09 13:36:18 Starting NimbusPolicyWatcher
2024/01/09 13:36:18 Starting policy processing loop
2024/01/09 13:37:28 NimbusPolicy: Detected policy: Name: multiubuntu, Namespace: sys-proc-path-sleep-block, ID: [sys-proc-paths]
{TypeMeta:{Kind:NimbusPolicy APIVersion:intent.security.nimbus.com/v1} ObjectMeta:{Name:sys-proc-path-sleep-block GenerateName: Namespace:multiubuntu SelfLink: UID:5d2ae075-98b8-4958-850e-8114cb6dec19 ResourceVersion:5753517 Generation:1 CreationTimestamp:2024-01-09 13:37:19 +0000 UTC DeletionTimestamp:<nil> DeletionGracePeriodSeconds:<nil> Labels:map[] Annotations:map[] OwnerReferences:[] Finalizers:[] ManagedFields:[{Manager:main Operation:Update APIVersion:intent.security.nimbus.com/v1 Time:2024-01-09 13:37:19 +0000 UTC FieldsType:FieldsV1 FieldsV1:{"f:spec":{".":{},"f:rules":{},"f:selector":{".":{},"f:matchLabels":{".":{},"f:group":{}}}}} Subresource:}]} Spec:{Selector:{MatchLabels:map[group:group-1]} NimbusRules:[{Id:sys-proc-paths Type: Description:block the execution of '/bin/sleep' Rule:[{RuleAction:Block MatchProtocols:[] MatchPaths:[{Path:/bin/sleep}] MatchDirectories:[] MatchPatterns:[] MatchCapabilities:[] MatchSyscalls:[] MatchSyscallPaths:[] FromCIDRSet:[] ToPorts:[]}]}]} Status:{PolicyStatus:}}
2024/01/09 13:37:28 Exporting and Applying NimbusPolicy to KubeArmorPolicy
2024-01-09T13:37:28Z    INFO    Start Converting a NimbusPolicy {"PolicyName": "sys-proc-path-sleep-block"}
2024-01-09T13:37:28Z    INFO    Apply a new KubeArmorPolicy     {"PolicyName": "sys-proc-path-sleep-block", "Policy": {"metadata":{"name":"sys-proc-path-sleep-block","namespace":"multiubuntu","creationTimestamp":null},"spec":{"selector":{"matchLabels":{"group":"group-1"}},"process":{"matchPaths":[{"path":"/bin/sleep"}]},"file":{},"network":{"matchProtocols":[{"protocol":"raw"}]},"capabilities":{"matchCapabilities":[{"capability":"lease"}]},"syscalls":{},"action":"Block"},"status":{}}}
2024/01/09 13:37:28 Successfully exported NimbusPolicy to KubeArmorPolicy
```
<br>
You can also see the policies that were actually created. 

``` 
$ kubectl get ksp -n multiubuntu
NAME                        AGE
sys-proc-path-sleep-block   3m24s
```
```
$  kubectl get ksp -n multiubuntu sys-proc-path-sleep-block -o yaml
apiVersion: security.kubearmor.com/v1
kind: KubeArmorPolicy
metadata:
  creationTimestamp: "2024-01-09T13:37:28Z"
  generation: 1
  name: sys-proc-path-sleep-block
  namespace: multiubuntu
  resourceVersion: "5753537"
  uid: 16cb107b-e442-442f-90fe-dbb139658d5e
spec:
  action: Block
  capabilities:
    matchCapabilities:
    - capability: lease
  file: {}
  network:
    matchProtocols:
    - protocol: raw
  process:
    matchPaths:
    - path: /bin/sleep
  selector:
    matchLabels:
      group: group-1
  syscalls: {}
```