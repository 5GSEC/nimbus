# Test: `securityintent-deletion-after-creation-of-nimbuspolicy`

This test verifies that when a SecurityIntent is the only one referenced by a SecurityIntentBinding, and that  SecurityIntent is deleted, the corresponding NimbusPolicy is also automatically deleted.


## Steps

| # | Name | Bindings | Try | Catch | Finally |
|:-:|---|:-:|:-:|:-:|:-:|
| 1 | [Create a SecurityIntentBinding](#step-Create a SecurityIntentBinding) | 0 | 1 | 0 | 0 |
| 2 | [Create a SecurityIntent](#step-Create a SecurityIntent) | 0 | 1 | 0 | 0 |
| 3 | [Verify NimbusPolicy creation](#step-Verify NimbusPolicy creation) | 0 | 1 | 0 | 0 |
| 4 | [Delete previously created SecurityIntent](#step-Delete previously created SecurityIntent) | 0 | 1 | 0 | 0 |
| 5 | [Verify the NimbusPolicy deletion](#step-Verify the NimbusPolicy deletion) | 0 | 1 | 0 | 0 |
| 6 | [Verify status of SecurityIntentBinding](#step-Verify status of SecurityIntentBinding) | 0 | 1 | 0 | 0 |

### Step: `Create a SecurityIntentBinding`

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

### Step: `Verify NimbusPolicy creation`

*No description*

#### Try

| # | Operation | Bindings | Outputs | Description |
|:-:|---|:-:|:-:|---|
| 1 | `assert` | 0 | 0 | *No description* |

### Step: `Delete previously created SecurityIntent`

*No description*

#### Try

| # | Operation | Bindings | Outputs | Description |
|:-:|---|:-:|:-:|---|
| 1 | `delete` | 0 | 0 | *No description* |

### Step: `Verify the NimbusPolicy deletion`

*No description*

#### Try

| # | Operation | Bindings | Outputs | Description |
|:-:|---|:-:|:-:|---|
| 1 | `script` | 0 | 0 | *No description* |

### Step: `Verify status of SecurityIntentBinding`

This verifies that upon deletion of a NimbusPolicy, the corresponding SecurityIntentBinding's status subresource is updated to reflect the current information.


#### Try

| # | Operation | Bindings | Outputs | Description |
|:-:|---|:-:|:-:|---|
| 1 | `assert` | 0 | 0 | *No description* |

---

