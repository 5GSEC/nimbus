# Quick Tutorials

## Deploy test pod

busybox-pod

```
$ kubectl apply -f ./test-yaml/busybox-pod.yaml
```

```yaml
apiVersion: v1
kind: Pod
metadata:
  name: busybox
  labels:
    app: busybox
spec:
  containers:
  - name: busybox
    image: busybox
    command: ['sh', '-c', 'echo Container is Running; while true; do sleep 3600; done']
```


redis-pod
```
$ kubectl apply -f ./test-yaml/redis-pod.yaml
```
```yaml
apiVersion: v1
kind: Pod
metadata:
  name: redis
  labels:
    app: redis
spec:
  containers:
  - name: redis
    image: redis
    ports:
    - containerPort: 6379
```




### Verify redis pods before applying policies
```
$ kubectl get pod -o wide
NAME      READY   STATUS    RESTARTS   AGE   IP           NODE        NOMINATED NODE   READINESS GATES
busybox   1/1     Running   0          47s   10.0.0.207   kubearmor   <none>           <none>
redis     1/1     Running   0          30s   10.0.0.7     kubearmor   <none>           <none>
```
```
$ kubectl exec -it busybox -- telnet <redis-pod-ip> 6379
Connected to <redis-pod-ip>
...
```

## Sample
You want to enforce a policy that blocks all external traffic from accessing port 6379 on an endpoint labeled 'app: redis' in the default namespace.

### Run Operators (Nimbus)
```
$ make run
```

### Create and apply intent file
```
$ kubectl apply -f ./test-yaml/intent-redis.yaml
```
```yaml
apiVersion: intent.security.nimbus.com/v1
kind: SecurityIntent
metadata:
  name: redis-ingress-deny-traffic
  namespace: default
spec:
  selector:
    match:
      any:
        - resources:
            names: ["redis-pod"]
            namespaces: ["default"]
            kinds: ["Pod"]
            matchLabels:
              app: "redis"
    cel:
      - "object.spec.template.spec.containers.all(container, container.ports.any(port, port.number == 6379))"
  intent: 
    action: block
    mode: strict
    type: network
    resource: 
      - key: "ingress"
        val: ["0.0.0.0/0-6379"]

```
<details>
  <summary>make run</summary>
  cclab@kubearmor:~/nimbus$ make run
make: go: Permission denied
test -s /home/cclab/nimbus/bin/controller-gen && /home/cclab/nimbus/bin/controller-gen --version | grep -q v0.13.0 || \
GOBIN=/home/cclab/nimbus/bin go install sigs.k8s.io/controller-tools/cmd/controller-gen@v0.13.0
/home/cclab/nimbus/bin/controller-gen rbac:roleName=manager-role crd webhook paths="./..." output:crd:artifacts:config=config/crd/bases
/home/cclab/nimbus/bin/controller-gen object:headerFile="hack/boilerplate.go.txt" paths="./..."
go fmt ./...
go vet ./...
go: downloading github.com/onsi/ginkgo/v2 v2.11.0
go: downloading github.com/onsi/gomega v1.27.10
go run ./main.go
2023-11-14T15:56:47Z	INFO	setup	starting manager
2023-11-14T15:56:47Z	INFO	controller-runtime.metrics	Starting metrics server
2023-11-14T15:56:47Z	INFO	starting server	{"kind": "health probe", "addr": "[::]:8081"}
2023-11-14T15:56:47Z	INFO	controller-runtime.metrics	Serving metrics server	{"bindAddress": ":8080", "secure": false}
2023-11-14T15:56:47Z	INFO	Starting EventSource	{"controller": "securityintent", "controllerGroup": "intent.security.nimbus.com", "controllerKind": "SecurityIntent", "source": "kind source: *v1.SecurityIntent"}
2023-11-14T15:56:47Z	INFO	Starting Controller	{"controller": "securityintent", "controllerGroup": "intent.security.nimbus.com", "controllerKind": "SecurityIntent"}
2023-11-14T15:56:47Z	INFO	Starting workers	{"controller": "securityintent", "controllerGroup": "intent.security.nimbus.com", "controllerKind": "SecurityIntent", "worker count": 1}
2023-11-14T16:00:00Z	INFO	SecurityIntent object fetched	{"controller": "securityintent", "controllerGroup": "intent.security.nimbus.com", "controllerKind": "SecurityIntent", "SecurityIntent": {"name":"redis-ingress-deny-traffic","namespace":"default"}, "namespace": "default", "name": "redis-ingress-deny-traffic", "reconcileID": "f1f990e8-35d2-4dfb-8107-164a916cfb2f", "intent": "redis-ingress-deny-traffic"}
2023-11-14T16:00:00Z	INFO	Applied CiliumNetworkPolicy	{"controller": "securityintent", "controllerGroup": "intent.security.nimbus.com", "controllerKind": "SecurityIntent", "SecurityIntent": {"name":"redis-ingress-deny-traffic","namespace":"default"}, "namespace": "default", "name": "redis-ingress-deny-traffic", "reconcileID": "f1f990e8-35d2-4dfb-8107-164a916cfb2f", "policy": {"namespace": "default", "name": "redis-ingress-deny-traffic"}}
2023-11-14T16:00:00Z	INFO	Successfully reconciled SecurityIntent	{"controller": "securityintent", "controllerGroup": "intent.security.nimbus.com", "controllerKind": "SecurityIntent", "SecurityIntent": {"name":"redis-ingress-deny-traffic","namespace":"default"}, "namespace": "default", "name": "redis-ingress-deny-traffic", "reconcileID": "f1f990e8-35d2-4dfb-8107-164a916cfb2f", "intent": "redis-ingress-deny-traffic"}
</details>



### Verify SecurityIntent creation
```
$ kubectl get SecurityIntent
NAME                         AGE
redis-ingress-deny-traffic   2m

```

### Verify Cilium Network Policy Creation
```
$ kubectl get CiliumNetworkPolicies
NAME                         AGE
redis-ingress-deny-traffic   2m19s
```
```
$ kubectl get CiliumNetworkPolicies redis-ingress-deny-traffic -o yaml
apiVersion: cilium.io/v2
kind: CiliumNetworkPolicy
metadata:
  creationTimestamp: "2023-11-14T16:00:00Z"
  generation: 1
  name: redis-ingress-deny-traffic
  namespace: default
  resourceVersion: "89051"
  uid: 1c3e4e7e-697f-4fbc-a3a8-3a91af6380e6
spec:
  endpointSelector:
    matchLabels:
      any:app: redis
  ingressDeny:
  - fromCIDRSet:
    - cidr: 0.0.0.0/0
    toPorts:
    - ports:
      - port: "6379"
        protocol: TCP
```

### Test for policy enforcement
```
$ kubectl exec -it busybox -- telnet <redis-pod-ip> 6379
```
You can see that the policy was applied, so access to port 6379 on the endpoint specified as 'app: redis' is blocked.


```
$ kubectl delete kubectl delete -f ./test-yaml/intent-redis.yaml
```