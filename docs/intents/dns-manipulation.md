## Objective

**Attack vectors**: Adversaries can manipulate DNS requests to steal data (ex-filtration), redirect network traffic, or
expose user activity. This can be achieved by tampering with a system's DNS configuration.

**Mitigation**: The `DNS Manipulation` `SecurityIntent` helps prevent these attacks by:

- **Enforcing Integrity**: It prevents unauthorized changes to the `/etc/resolv.conf` file, which configures DNS
  settings.
  This ensures the system uses the intended DNS server.
- **Restricting DNS Access**: It allows pods to only resolve DNS queries through the designated DNS service within the
  cluster, typically located in the `kube-system` namespace.

> [!Note]
> Enforcement is handled by the relevant security engines. In this case, [KubeArmor](https://kubearmor.io/) and a CNI
> capable of
> enforcing [Kubernetes NetworkPolicy](https://kubernetes.io/docs/concepts/services-networking/network-policies/).

## Policies

From the `DNSManipulation` SecurityIntent two security policies will be generated to satisfy the SecurityIntent:

- **KubeArmor Policy**: This policy prevents changes to the `/etc/resolv.conf` file, ensuring DNS configuration
  integrity.
- **Kubernetes Network Policy**: This policy allows DNS requests only to `kube-dns` pods within the `kube-system`
  namespace.
  This restricts access to the designated DNS server for secure name resolution.
