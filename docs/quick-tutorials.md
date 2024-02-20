# Quick Tutorials

## Prerequisites

- **Nimbus operator**: Follow [this](../deployments/nimbus/Readme.md) guide to install `nimbus` operator.
- Nimbus adapters: To generate multiple security engines policies
    - `nimbus-kubearmor`: Follow [this](../deployments/nimbus-kubearmor/Readme.md) guide to install `nimbus-kubearmor`
      adapter.
    - `nimbus-netpol`: Follow [this](../deployments/nimbus-netpol/Readme.md) guide to install `nimbus-netpol` adapter.

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

## Verify Resources

* SecurityIntent

```shell
$ kubectl get securityintent
NAME               STATUS    AGE
dns-manipulation   Created   9s
```

Output in `-o wide` for detailed info:
```shell
$ kubectl get securityintent dns-manipulation -o wide
NAME               STATUS    AGE   ID                ACTION
dns-manipulation   Created   17s   dnsManipulation   Block
```

* SecurityIntentBinding

```shell
$ kubectl get securityintentbinding
NAME                       STATUS    AGE   INTENTS   NIMBUSPOLICY
dns-manipulation-binding   Created   69s   1         dns-manipulation-binding
```

* NimbusPolicy

```shell
$ kubectl get nimbuspolicy
NAME                       STATUS    AGE    POLICIES
dns-manipulation-binding   Created   2m9s   2
```

Describe the nimbuspolicy to check which policies are created:

```shell
$ kubectl describe nimbuspolicy dns-manipulation-binding
Name:         dns-manipulation-binding
Namespace:    default
Labels:       <none>
Annotations:  <none>
API Version:  intent.security.nimbus.com/v1
Kind:         NimbusPolicy
Metadata:
  Creation Timestamp:  2024-02-20T06:04:32Z
  Generation:          1
  Owner References:
    API Version:           intent.security.nimbus.com/v1
    Block Owner Deletion:  true
    Controller:            true
    Kind:                  SecurityIntentBinding
    Name:                  dns-manipulation-binding
    UID:                   c3b7046f-26c7-4edb-ad82-de243e9ee378
  Resource Version:        56960
  UID:                     109a7b54-8643-487e-9454-6a79c5f4cacc
Spec:
  Rules:
    Description:  An adversary can manipulate DNS requests to redirect network traffic and potentially reveal end user activity.
    Id:           dnsManipulation
    Rule:
      Action:  Block
  Selector:
    Match Labels:
      App:  nginx
Status:
  Adapter Policies:
    KubeArmorPolicy/dns-manipulation-binding-dnsmanipulation
    NetworkPolicy/dns-manipulation-binding-dnsmanipulation
  Last Updated:                2024-02-20T06:04:32Z
  Number Of Adapter Policies:  2
  Status:                      Created
Events:                        <none>
```

## Verify the Security Engines policies
Review the policies that are successfully generated as part of `DNSManipulation` SecurityIntent and
SecurityIntentBinding:

### KubeArmorPolicy

```shell
$ kubectl get kubearmorpolicy
NAME                                       AGE
dns-manipulation-binding-dnsmanipulation   5m45s
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
  creationTimestamp: "2024-02-20T06:04:32Z"
  generation: 1
  name: dns-manipulation-binding-dnsmanipulation
  namespace: default
  ownerReferences:
    - apiVersion: intent.security.nimbus.com/v1
      blockOwnerDeletion: true
      controller: true
      kind: NimbusPolicy
      name: dns-manipulation-binding
      uid: 109a7b54-8643-487e-9454-6a79c5f4cacc
  resourceVersion: "56955"
  uid: 03afa2ec-ea86-4248-9f63-243493aa1db9
spec:
  action: Block
  capabilities: { }
  file:
    matchPaths:
      - path: /etc/resolv.conf
        readOnly: true
  message: An adversary can manipulate DNS requests to redirect network traffic and
    potentially reveal end user activity.
  network: { }
  process: { }
  selector:
    matchLabels:
      app: nginx
  syscalls: { }
```

### NetworkPolicy

```shell
$  kubectl get networkpolicy
NAME                                       POD-SELECTOR   AGE
dns-manipulation-binding-dnsmanipulation   app=nginx      6m43s
```

Inspect policy for detailed info:

```shell
$ kubectl get networkpolicy dns-manipulation-binding-dnsmanipulation -o yaml
```

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  annotations:
    app.kubernetes.io/managed-by: nimbus-netpol
  creationTimestamp: "2024-02-20T06:04:32Z"
  generation: 1
  name: dns-manipulation-binding-dnsmanipulation
  namespace: default
  ownerReferences:
    - apiVersion: intent.security.nimbus.com/v1
      blockOwnerDeletion: true
      controller: true
      kind: NimbusPolicy
      name: dns-manipulation-binding
      uid: 109a7b54-8643-487e-9454-6a79c5f4cacc
  resourceVersion: "56956"
  uid: 473c293e-3006-4843-9eb3-2a21f142d6e3
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
$ kubectl get securityintent,securityintentbinding,nimbuspolicy,kubearmorpolicy,netpol -A
No resources found
```

## Next steps

- Try out other sample [SecurityIntents](../examples/namespaced) and review the policy generation.
- Checkout [Security Intents](https://github.com/5GSEC/security-intents).
