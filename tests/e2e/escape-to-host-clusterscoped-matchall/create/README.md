# Test: `escape-to-host-clusterscoped-matchall-adapter-policy-creation`

This test validates that creating a `escapeToHost` SecurityIntent with ClusterSecurityIntentBinding with a matchNames of "*" generates the expected cluster Kyverno Policy, and kubearmor policies in all the namespaces in the cluster except kube-system


## Steps

| # | Name | Bindings | Try | Catch | Finally |
|:-:|---|:-:|:-:|:-:|:-:|
| 1 | [Create the dev, staging namespaces ](#step-Create the dev, staging namespaces ) | 0 | 1 | 0 | 0 |
| 2 | [Create a SecurityIntent](#step-Create a SecurityIntent) | 0 | 1 | 0 | 0 |
| 3 | [Create a ClusterSecurityIntentBinding](#step-Create a ClusterSecurityIntentBinding) | 0 | 1 | 0 | 0 |
| 4 | [Verify ClusterNimbusPolicy creation](#step-Verify ClusterNimbusPolicy creation) | 0 | 1 | 0 | 0 |
| 5 | [Verify Nimbus Policy creation in dev](#step-Verify Nimbus Policy creation in dev) | 0 | 1 | 0 | 0 |
| 6 | [Verify Nimbus Policy creation in staging](#step-Verify Nimbus Policy creation in staging) | 0 | 1 | 0 | 0 |
| 7 | [Verify Nimbus Policy creation in default](#step-Verify Nimbus Policy creation in default) | 0 | 1 | 0 | 0 |
| 8 | [Verify NimbusPolicy absence in kube-system](#step-Verify NimbusPolicy absence in kube-system) | 0 | 1 | 0 | 0 |
| 9 | [Verify Cluster KyvernoPolicy creation](#step-Verify Cluster KyvernoPolicy creation) | 0 | 1 | 0 | 0 |
| 10 | [Verify spec, status of created ClusterSecurityIntentBinding](#step-Verify spec, status of created ClusterSecurityIntentBinding) | 0 | 2 | 2 | 0 |

### Step: `Create the dev, staging namespaces `

*No description*

#### Try

| # | Operation | Bindings | Outputs | Description |
|:-:|---|:-:|:-:|---|
| 1 | `apply` | 0 | 0 | *No description* |

### Step: `Create a SecurityIntent`

*No description*

#### Try

| # | Operation | Bindings | Outputs | Description |
|:-:|---|:-:|:-:|---|
| 1 | `apply` | 0 | 0 | *No description* |

### Step: `Create a ClusterSecurityIntentBinding`

*No description*

#### Try

| # | Operation | Bindings | Outputs | Description |
|:-:|---|:-:|:-:|---|
| 1 | `apply` | 0 | 0 | *No description* |

### Step: `Verify ClusterNimbusPolicy creation`

*No description*

#### Try

| # | Operation | Bindings | Outputs | Description |
|:-:|---|:-:|:-:|---|
| 1 | `assert` | 0 | 0 | *No description* |

### Step: `Verify Nimbus Policy creation in dev`

*No description*

#### Try

| # | Operation | Bindings | Outputs | Description |
|:-:|---|:-:|:-:|---|
| 1 | `assert` | 0 | 0 | *No description* |

### Step: `Verify Nimbus Policy creation in staging`

*No description*

#### Try

| # | Operation | Bindings | Outputs | Description |
|:-:|---|:-:|:-:|---|
| 1 | `assert` | 0 | 0 | *No description* |

### Step: `Verify Nimbus Policy creation in default`

*No description*

#### Try

| # | Operation | Bindings | Outputs | Description |
|:-:|---|:-:|:-:|---|
| 1 | `assert` | 0 | 0 | *No description* |

### Step: `Verify NimbusPolicy absence in kube-system`

*No description*

#### Try

| # | Operation | Bindings | Outputs | Description |
|:-:|---|:-:|:-:|---|
| 1 | `script` | 0 | 0 | *No description* |

### Step: `Verify Cluster KyvernoPolicy creation`

*No description*

#### Try

| # | Operation | Bindings | Outputs | Description |
|:-:|---|:-:|:-:|---|
| 1 | `assert` | 0 | 0 | *No description* |

### Step: `Verify spec, status of created ClusterSecurityIntentBinding`

Verify the created SecurityIntentBinding status subresource includes the number and names of bound intents, along with the generated NimbusPolicy name.


#### Try

| # | Operation | Bindings | Outputs | Description |
|:-:|---|:-:|:-:|---|
| 1 | `script` | 0 | 1 | *No description* |
| 2 | `assert` | 0 | 0 | *No description* |

#### Catch

| # | Operation | Bindings | Outputs | Description |
|:-:|---|:-:|:-:|---|
| 1 | `script` | 0 | 0 | *No description* |
| 2 | `script` | 0 | 0 | *No description* |

---

