## Objective

- The primary goal of the dns-manipulation intent is to enforce policies that manage and restrict how pods interact with DNS services, particularly regarding outbound DNS queries. This can be crucial for security, compliance, and operational integrity.  

- The dns-manipulation intent presumably defines the requirement for the application to interact with DNS services, potentially to manipulate DNS requests or responses based on the application’s needs.

- The dnsManipulation intent emphasizes the importance of safeguarding DNS resolution mechanisms. This is crucial because adversaries might exploit vulnerabilities to alter DNS requests, redirect traffic, or expose sensitive user activity.

**Note** : For the dns-manipulation intent one needs to have either [nimbus-netpol](../../deployments/nimbus-netpol/Readme.md) adapter or [nimbus-kubearmor](../../deployments/nimbus-kubearmor/Readme.md) adapter or both adapters running in their cluster

## Policy Creation

The dns-manipulation intent results in two policies `NetworkPolicy` and a `KubeArmorPolicy`. Below are the behaviours of intent in terms of policy:

### KubeArmor Policy

#### Prereq

- For the `KubeArmorPolicy` to work, one should have a [BPF-LSM](https://github.com/kubearmor/KubeArmor/blob/main/getting-started/FAQ.md#checking-and-enabling-support-for-bpf-lsm) enabled for each node in their cluster.


#### Policy Description

- The `KubeArmorPolicy` is configured to block any unauthorized access or modifications to critical files, particularly `/etc/resolv`.conf, which is essential for DNS resolution in Linux-based systems.

- This file is where the system looks for DNS servers and configuration details. Ensuring it is read-only protects against malicious changes that could redirect DNS requests.

- The policy defines a Block action, indicating that any attempts to modify or write to the specified file `/etc/resolv.conf` will be denied.

- This approach minimizes the attack surface of the pods by limiting egress traffic strictly to defined endpoints, which helps in maintaining a secure network posture.

- By securing `/etc/resolv.conf`, the policy effectively mitigates the risk of DNS spoofing or hijacking, which can lead to compromised network traffic and potential data leakage.

----

### Network Policy

#### Prereq

- For the `NetworkPolicy` to work, one should have a [Calico-CNI](https://docs.tigera.io/calico/latest/getting-started/kubernetes/self-managed-onprem/onpremises)  installed in their cluster.


#### Policy Description

- The `NetworkPolicy` created as a result of this intent includes specific egress rules that align with the intent’s goals. It reflects the desire to secure and control DNS traffic.

- By specifying the egress rules to allow traffic to the kube-dns service, the policy ensures that pods  can resolve DNS queries through the designated DNS service within the cluster.

- By allowing access to kube-dns, the intent ensures that the pods can perform DNS lookups necessary for their operation without exposing them to arbitrary external IPs.

- This approach minimizes the attack surface of the pods by limiting egress traffic strictly to defined endpoints, which helps in maintaining a secure network posture.

- It ensures compliance with security policies that require minimal exposure of services to the outside world while allowing necessary functionality.


----

Together with the NetworkPolicy, this KubeArmorPolicy creates a layered defense strategy:

- The NetworkPolicy restricts egress traffic to legitimate DNS services and IP addresses.

- The KubeArmorPolicy protects the integrity of the DNS configuration itself, ensuring that no unauthorized processes can alter it.