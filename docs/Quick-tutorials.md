# Quick Tutorials

## Deploy test pod
```
$ kubectl apply -f ./test/env/multiubuntu.yaml
```

## Sample

### Run Operators (Nimbus)
```
$ make run
```

### Run Adapter Server
```
$ cd nimbus-kubearmor/receiver/server
$ go run server.go
2024/01/02 20:35:46 Server starting on port 13000...
```

### Create and apply Securityintent and SecurityintentBinding file
```
$ kubectl apply -f ./test/v2/intents/system/intent-path-block.yaml
```

```
$ kubectl apply -f ./test/v2/bindings/system/binding-path-block.yaml
```


### Verify SecurityIntent and SecurityIntentBinding
```
$ kubectl get SecurityIntent -n multiubuntu
NAME                            AGE
group-1-proc-path-sleep-block   25s

```
```
$ kubectl get SecurityIntentBinding -n multiubuntu
NAME                        AGE
sys-proc-path-sleep-block   29s

```

### Verify Nimbus policy
```
$ kubectl get nimbuspolicy -n multiubuntu
NAME                     AGE
net-redis-ingress-deny   38s
```
```
$ kubectl get np -n multiubuntu sys-proc-path-sleep-block -o yaml
apiVersion: intent.security.nimbus.com/v1
kind: NimbusPolicy
metadata:
  creationTimestamp: "2024-01-02T20:37:33Z"
  generation: 1
  name: sys-proc-path-sleep-block
  namespace: multiubuntu
  resourceVersion: "4281015"
  uid: 00c3de93-92d4-4a88-bff6-389449751e3c
spec:
  rules:
  - description: block the execution of '/bin/sleep'
    id: sys-path-exec
    rule:
    - action: Block
      matchPaths:
      - path: /bin/sleep
  selector:
    matchLabels:
      group: group-1
```