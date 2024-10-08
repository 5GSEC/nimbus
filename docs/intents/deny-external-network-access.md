## Objective

- The denyExternalNetworkAccess intent focuses on enhancing security by restricting external network access for specific applications, such as those labeled with selectors. This intent aims to ensure that these applications can only communicate with trusted internal resources while preventing unauthorized access from external networks.

- The goal of the denyExternalNetworkAccess intent is to create a secure environment for the application by limiting both ingress and egress traffic. This is critical for minimizing the attack surface and protecting sensitive data from external threats.

**Note** : For the denyExternalNetworkAccess intent one needs to have  [nimbus-netpol](../../deployments/nimbus-netpol/Readme.md) adapter running in their cluster. To install the complete suite with all the adapters pls follow the steps mentioned [here](../getting-started.md#nimbus)

## Policy Creation

The denyExternalNetworkAccess intent results in `NetworkPolicy`. Below is the behaviour of intent in terms of policy: 

### Network Policy

#### Prereq

- For the `NetworkPolicy` to work, one should have a [Calico-CNI](https://docs.tigera.io/calico/latest/getting-started/kubernetes/self-managed-onprem/onpremises)  installed in their cluster.

#### Policy Description

- The NetworkPolicy created as a result of this intent defines rules that enforce restricted network access:

    - **Egress Rules:** The policy allows outbound traffic only to specific IP ranges and the kube-dns service, enabling the application to resolve DNS queries while restricting communication to the external network.

    - **Ingress Rules:** The policy specifies that only traffic from defined internal IP ranges can reach the pods, ensuring that only trusted sources can communicate with them.

- By limiting both ingress and egress traffic, this policy significantly reduces the risk of data exfiltration and unauthorized access.

- The application can securely operate within a controlled environment while still being able to resolve DNS queries necessary for its functionality.

