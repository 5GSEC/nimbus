# Quick Tutorials

## Deploy test pod

busybox-pod

```
$ kubectl apply -f ./test/env/busybox-pod.yaml
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
$ kubectl apply -f ./test/env/redis-pod.yaml
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

### Create and apply Securityintent and SecurityintentBinding file
```
$ kubectl apply -f ./test/v2/intents/network/intent-redis.yaml
```
```yaml
apiVersion: intent.security.nimbus.com/v1
kind: SecurityIntent
metadata:
  name: redis-ingress-deny-traffic
  namespace: default
spec:
  intent:
    description: "Block port 6379"
    action: Block
    type: network
    resource:
      - fromCIDRSet:
          - cidr: 0.0.0.0/0
        toPorts:
          - ports:
            - port: "6379"
              protocol: tcp

```

```
$ kubectl apply -f ./test/v2/bindings/network/binding-redis.yaml
```
```yaml
apiVersion: intent.security.nimbus.com/v1
kind: SecurityIntentBinding
metadata:
  name: net-redis-ingress-deny
  namespace: default
spec:
  selector:
      any:
        - resources:
            kind: Pod
            namespace: default
            matchLabels:
              app: "redis"
  intentRequests:
    - type: network
      intentName: redis-ingress-deny-traffic
      description: "Don’t allow any outside traffic to the Redis port"
      mode: strict

```

### Verify SecurityIntent and SecurityIntentBinding
```
$ kubectl get SecurityIntent
NAME                         AGE
redis-ingress-deny-traffic   26s

```
```
$ kubectl get SecurityIntentBinding
NAME                     AGE
net-redis-ingress-deny   25s

```

### Verify Cilium Network Policy Creation
```
$ kubectl get cnp
NAME                     AGE
net-redis-ingress-deny   38s
```
```
$ kubectl get cnp net-redis-ingress-deny -o yaml
apiVersion: cilium.io/v2
kind: CiliumNetworkPolicy
metadata:
  creationTimestamp: "2023-12-12T18:11:36Z"
  generation: 1
  name: net-redis-ingress-deny
  namespace: default
  resourceVersion: "3415520"
  uid: 0b1d9071-16e8-405f-bd92-b3774c41a9df
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
$ kubectl delete -f ./test-yaml/intents/network/intent-redis.yaml
```