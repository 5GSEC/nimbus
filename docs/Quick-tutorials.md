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

### Run Adapter Server
```
$ cd /nimbus-kubearmor/receiver/server
$ go run server.go
```

### Create and apply Securityintent and SecurityintentBinding file
```
$ kubectl apply -f ./test/v2/intents/network/intent-redis.yaml
```

```
$ kubectl apply -f ./test/v2/bindings/network/binding-redis.yaml
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
$ kubectl get nimbuspolicy
NAME                     AGE
net-redis-ingress-deny   38s
```
```
$ kubectl get cnp net-redis-ingress-deny -o yaml
```