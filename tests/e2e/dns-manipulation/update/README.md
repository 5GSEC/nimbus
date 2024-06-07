# Test: `netpol-kubearmor-adapter-policy-updation`

This test validates that direct updates to the generated adapter's policies are discarded, to maintain consistency and prevent unintended modifications.


## Steps

| # | Name | Bindings | Try | Catch | Finally |
|:-:|---|:-:|:-:|:-:|:-:|
| 1 | [Create a SecurityIntent](#step-Create a SecurityIntent) | 0 | 1 | 0 | 0 |
| 2 | [Create a SecurityIntentBinding](#step-Create a SecurityIntentBinding) | 0 | 1 | 0 | 0 |
| 3 | [Verify NimbusPolicy creation](#step-Verify NimbusPolicy creation) | 0 | 1 | 0 | 0 |
| 4 | [Verify NetworkPolicy creation](#step-Verify NetworkPolicy creation) | 0 | 1 | 0 | 0 |
| 5 | [Verify KubearmorPolicy creation](#step-Verify KubearmorPolicy creation) | 0 | 1 | 0 | 0 |
| 6 | [Update existing NetworkPolicy](#step-Update existing NetworkPolicy) | 0 | 1 | 0 | 0 |
| 7 | [Verify discarding of the changes made in NetworkPolicy](#step-Verify discarding of the changes made in NetworkPolicy) | 0 | 1 | 0 | 0 |
| 8 | [Update existing KubearmorPolicy](#step-Update existing KubearmorPolicy) | 0 | 1 | 0 | 0 |
| 9 | [Verify discarding of the changes made in KubeArmorPolicy](#step-Verify discarding of the changes made in KubeArmorPolicy) | 0 | 1 | 0 | 0 |

### Step: `Create a SecurityIntent`

*No description*

#### Try

| # | Operation | Bindings | Outputs | Description |
|:-:|---|:-:|:-:|---|
| 1 | `apply` | 0 | 0 | *No description* |

### Step: `Create a SecurityIntentBinding`

*No description*

#### Try

| # | Operation | Bindings | Outputs | Description |
|:-:|---|:-:|:-:|---|
| 1 | `apply` | 0 | 0 | *No description* |

### Step: `Verify NimbusPolicy creation`

*No description*

#### Try

| # | Operation | Bindings | Outputs | Description |
|:-:|---|:-:|:-:|---|
| 1 | `assert` | 0 | 0 | *No description* |

### Step: `Verify NetworkPolicy creation`

*No description*

#### Try

| # | Operation | Bindings | Outputs | Description |
|:-:|---|:-:|:-:|---|
| 1 | `assert` | 0 | 0 | *No description* |

### Step: `Verify KubearmorPolicy creation`

*No description*

#### Try

| # | Operation | Bindings | Outputs | Description |
|:-:|---|:-:|:-:|---|
| 1 | `assert` | 0 | 0 | *No description* |

### Step: `Update existing NetworkPolicy`

*No description*

#### Try

| # | Operation | Bindings | Outputs | Description |
|:-:|---|:-:|:-:|---|
| 1 | `apply` | 0 | 0 | *No description* |

### Step: `Verify discarding of the changes made in NetworkPolicy`

*No description*

#### Try

| # | Operation | Bindings | Outputs | Description |
|:-:|---|:-:|:-:|---|
| 1 | `assert` | 0 | 0 | *No description* |

### Step: `Update existing KubearmorPolicy`

*No description*

#### Try

| # | Operation | Bindings | Outputs | Description |
|:-:|---|:-:|:-:|---|
| 1 | `apply` | 0 | 0 | *No description* |

### Step: `Verify discarding of the changes made in KubeArmorPolicy`

*No description*

#### Try

| # | Operation | Bindings | Outputs | Description |
|:-:|---|:-:|:-:|---|
| 1 | `assert` | 0 | 0 | *No description* |

---

