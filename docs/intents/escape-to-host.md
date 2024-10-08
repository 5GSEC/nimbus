## Objective

- The escapeToHost intent focuses on preventing potential vulnerabilities that could allow an attacker to escape from a containerized environment and gain unauthorized access to the host machine. This is a crucial aspect of container security, as it aims to ensure that the isolation provided by containers is not compromised.

- The goal of the escapeToHost intent is to enforce strict security standards on pods to prevent any breaches that could lead to container escape. This is particularly relevant for applications, which may have extensive network exposure.

- The escapeToHost intent is also focused on preventing attackers from breaching container boundaries and gaining access to the host machine. This is a critical aspect of securing containerized environments, as it helps maintain the isolation provided by containers, thereby protecting the host and other workloads from potential compromises.

**Note** : For the escapeToHost intent one needs to have either [nimbus-kyverno](../../deployments/nimbus-kyverno/Readme.md) adapter or [nimbus-kubearmor](../../deployments/nimbus-kubrarmor/Readme.md) or both adapters running in their cluster. (it is recommended to have both the adapters in place to make sure that the intent work with its full potential). To install the complete suite with all the adapters pls follow the steps mentioned [here](../getting-started.md#nimbus)

## Policy Creation

The escapeToHost intent results in `KyvernoPolicy` and a couple of `KubearmorPolicy`. Below is the behaviour of intent in terms of policy: 

### Kyverno Policy

#### Policy Description

- The KyvernoPolicy defined here implements rules that align with the escape prevention objectives:

    - **Admission Control:** The policy is applied at admission time, meaning it will validate pod creation requests before they are accepted by the Kubernetes API.

    - **Background Scanning:** The policy also runs in the background to continuously validate existing pods against the specified security criteria. And provide the PolicyReports for the same.

- The policy enforces a pod security standard at the baseline level, which includes basic security measures to mitigate risks associated with container escape.

- By requiring a baseline pod security level, the policy enforces essential security practices, such as ensuring that pods do not run as root or have excessive privileges. This helps mitigate the risk of attackers exploiting container vulnerabilities to gain access to the host system.

- User can change the pod security level by specifying any pod security level enforced by KyvernoPolicy in the intent params as:

   ```
    params:
      psa_level: ["restricted"]
   ```

- The `escapeToHost` intent and corresponding policy work together to establish a strong security posture for the application. By enforcing pod security standards, the policy reduces the risk of container escape, which is critical for maintaining the integrity of the host system.


- The use of admission control ensures that potential security issues are addressed before they can affect the running environment.


### KubeArmor Policy

#### Prereq

- For the `KubeArmorPolicy` to work, one should have a [BPF-LSM](https://github.com/kubearmor/KubeArmor/blob/main/getting-started/FAQ.md#checking-and-enabling-support-for-bpf-lsm) enabled for each node in their cluster.


#### Policy Description

 The intent is implemented through three distinct KubeArmorPolicy configurations, each addressing different aspects of container security:

 - **Disallow Capabilities Policy**

    - Capabilities: The policy blocks specific Linux capabilities that are critical for system-level access:

        - `sys_admin`: Allows a process to perform administrative tasks.
        - `sys_ptrace`: Enables processes to observe and control other processes.
        - `sys_ptrace`: Allows loading and unloading of kernel modules.
        - `dac_read_search` and `dac_override`: Affect discretionary access control.

- **Disallow Chroot Policy**

    - Process: Specifically blocks access to chroot binaries, which can be used to change the root filesystem of a process. This could allow an attacker to escape from their container environment.

        - Match Paths: Includes `/usr/sbin/chroot` and `/sbin/chroot`.

- **Disallow Deployment Tools Policy**

    - Process: Blocks access to various package management and build tools that could be exploited to manipulate the container or the host.
    
        - Match Paths: Includes package managers like `apt`, `yum`, `dnf`, `zypper`, as well as build tools like `make`, and network utilities like `curl` and `wget`.


- Each policy is designed to actively block actions that could lead to a compromise of the host system. By preventing access to critical capabilities and processes, these policies effectively reduce the risk of container escape.

- These policies together create a robust security mechanism to protect the host from potential breaches originating from the application. They ensure that even if an attacker manages to compromise a container, their ability to impact the host is severely limited.
