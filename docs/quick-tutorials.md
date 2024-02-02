# Quick Tutorials

## Install Nimbus Operator

Follow [this](../deployments/nimbus/Readme.md) guide to install `nimbus` operator.

## Install Adapters

### KubeArmor

Follow [this](../deployments/nimbus-kubearmor/Readme.md) guide to install `nimbus-kubearmor` adapter.

### Network Policy

Follow [this](../deployments/nimbus-netpol/Readme.md) guide to install `nimbus-netpol` adapter.

## Create a sample deployment

```shell
kubectl apply -f ./examples/env/nginx-deploy.yaml
```

## Create SecurityIntent and SecurityIntentBinding

### [DNS Manipulation](https://fight.mitre.org/techniques/FGT5006)

Create SecurityIntent and SecurityIntentBinding to prevent DNS Manipulation.

```shell
$ kubectl apply -f ./examples/namespaced/dns-manipulation-si-sib.yaml
securityintent.intent.security.nimbus.com/dns-manipulation created
securityintentbinding.intent.security.nimbus.com/dns-manipulation-binding created
```

## Verify SecurityIntent and SecurityIntentBinding

* Verify SecurityIntent

```shell
$ kubectl get securityintent
NAME               STATUS
dns-manipulation   Created
```

* Verify SecurityIntentBinding

```shell
$ kubectl get securityintentbinding
NAME                       STATUS
dns-manipulation-binding   Created
```

## Verify the Security Engines policies

### KubeArmorPolicy

Review the policies that were successfully generated as part of `DNSManipulation` SecurityIntent and
SecurityIntentBinding:

```shell
$ kubectl get kubearmorpolicy
NAME                                       AGE
dns-manipulation-binding-dnsmanipulation   2m44s
```

Inspect the policy for detailed info:

```shell
$ kubectl get kubearmorpolicy dns-manipulation-binding-dnsmanipulation -o yaml
```

```yaml
apiVersion: security.kubearmor.com/v1
kind: KubeArmorPolicy
metadata:
  annotations:
    app.kubernetes.io/managed-by: nimbus-kubearmor
  creationTimestamp: "2024-02-02T08:27:03Z"
  generation: 1
  name: dns-manipulation-binding-dnsmanipulation
  namespace: default
  ownerReferences:
    - apiVersion: intent.security.nimbus.com/v1
      blockOwnerDeletion: true
      controller: true
      kind: NimbusPolicy
      name: dns-manipulation-binding
      uid: c2571f5b-8299-4e0f-9594-b6804a5a4d8f
  resourceVersion: "610470"
  uid: 7f23a7f3-3012-449d-92ee-1ea2a741b7ec
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

Review the network policies that were successfully generated as part of `DNSManipulation` SecurityIntent and
SecurityIntentBinding:

```shell
$ kubectl get networkpolicy
NAME                                       POD-SELECTOR   AGE
dns-manipulation-binding-dnsmanipulation   app=nginx      5m54s
```

Inspect policy for detailed info:

```shell
$ kubectl get networkpolicy multiple-sis-nsscoped-binding-dnsmanipulation -o yaml
```

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  annotations:
    app.kubernetes.io/managed-by: nimbus-netpol
  creationTimestamp: "2024-02-02T08:27:03Z"
  generation: 1
  name: dns-manipulation-binding-dnsmanipulation
  namespace: default
  ownerReferences:
    - apiVersion: intent.security.nimbus.com/v1
      blockOwnerDeletion: true
      controller: true
      kind: NimbusPolicy
      name: dns-manipulation-binding
      uid: c2571f5b-8299-4e0f-9594-b6804a5a4d8f
  resourceVersion: "610469"
  uid: 7cbf50e3-8c47-443e-8851-01b0ca167bd3
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

From the `DNSManipulation` SecurityIntent two security policies were generated:

- KubeArmor Policy: This policy prevents modification of the `/etc/resolv.conf` file, ensuring the integrity of DNS
  configuration and preventing potential DNS hijacking.


- Kubernetes Network Policy: This policy allows outbound traffic on UDP and TCP ports 53 only to the
  `kube-dns` pods within the `kube-system` namespace. This restricts access to the DNS server, enhancing security while
  enabling pods to resolve DNS names.

## Cleanup

* The SecurityIntent and SecurityIntentBinding created earlier are no longer needed and can be deleted:

```shell
$ kubectl delete -f ./examples/namespaced/dns-manipulation-si-sib.yaml
securityintent.intent.security.nimbus.com "dns-manipulation" deleted
securityintentbinding.intent.security.nimbus.com "dns-manipulation-binding" deleted
```

* Delete deployment

```shell
$ kubectl delete -f ./examples/env/nginx-deploy.yaml
deployment.apps "nginx" deleted
```

* Confirm all resources have been deleted (Optional)

```shell
$ kubectl get securityintent,securityintentbinding,kubearmorpolicy,netpol -A
No resources found
```

## Next steps

- Try out other sample [SecurityIntents](../examples/namespaced) and review the policy generation.
- Checkout [Security Intents](https://github.com/5GSEC/security-intents).
