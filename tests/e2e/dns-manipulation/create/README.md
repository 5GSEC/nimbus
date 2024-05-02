# Test: `netpol-kubearmor-adapter-policy-creation`

This test validates that creating a `dns-manipulation` SecurityIntent with SecurityIntentBinding generates the expected Network Policy and KubeArmor policy.


## Steps

| # | Name | Bindings | Try | Catch | Finally |
|:-:|---|:-:|:-:|:-:|:-:|
| 1 | [Create a SecurityIntent](#step-Create a SecurityIntent) | 0 | 1 | 0 | 0 |
| 2 | [Create a SecurityIntentBinding](#step-Create a SecurityIntentBinding) | 0 | 1 | 0 | 0 |
| 3 | [Verify NimbusPolicy creation](#step-Verify NimbusPolicy creation) | 0 | 1 | 0 | 0 |
| 4 | [Verify NetworkPolicy creation](#step-Verify NetworkPolicy creation) | 0 | 1 | 0 | 0 |
| 5 | [Verify KubeArmorPolicy creation](#step-Verify KubeArmorPolicy creation) | 0 | 1 | 0 | 0 |
| 6 | [Verify status of created SecurityIntentBinding](#step-Verify status of created SecurityIntentBinding) | 0 | 1 | 0 | 0 |
| 7 | [Verify status of created NimbusPolicy](#step-Verify status of created NimbusPolicy) | 0 | 1 | 0 | 0 |
| 8 | [Verify that the corresponding netpol native policy is being updated in the NimbusPolicy status](#step-Verify that the corresponding netpol native policy is being updated in the NimbusPolicy status) | 0 | 1 | 0 | 0 |
| 9 | [Verify that the corresponding kubearmor native policy is being updated in the NimbusPolicy status](#step-Verify that the corresponding kubearmor native policy is being updated in the NimbusPolicy status) | 0 | 1 | 0 | 0 |

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

### Step: `Verify KubeArmorPolicy creation`

*No description*

#### Try

| # | Operation | Bindings | Outputs | Description |
|:-:|---|:-:|:-:|---|
| 1 | `assert` | 0 | 0 | *No description* |

### Step: `Verify status of created SecurityIntentBinding`

Verify the created SecurityIntentBinding status subresource includes the number and names of bound intents,  along with the generated NimbusPolicy name.


#### Try

| # | Operation | Bindings | Outputs | Description |
|:-:|---|:-:|:-:|---|
| 1 | `assert` | 0 | 0 | *No description* |

### Step: `Verify status of created NimbusPolicy`

*No description*

#### Try

| # | Operation | Bindings | Outputs | Description |
|:-:|---|:-:|:-:|---|
| 1 | `assert` | 0 | 0 | *No description* |

### Step: `Verify that the corresponding netpol native policy is being updated in the NimbusPolicy status`

*No description*

#### Try

| # | Operation | Bindings | Outputs | Description |
|:-:|---|:-:|:-:|---|
| 1 | `script` | 0 | 0 | *No description* |

### Step: `Verify that the corresponding kubearmor native policy is being updated in the NimbusPolicy status`

*No description*

#### Try

| # | Operation | Bindings | Outputs | Description |
|:-:|---|:-:|:-:|---|
| 1 | `script` | 0 | 0 | *No description* |

---

