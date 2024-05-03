# Test: `securityintentbinding-and-securityintent-independent-creation`

This test verifies the independent creation of SecurityIntent and SecurityIntentBinding custom resources. It ensures users can create these custom resources individually without requiring one to exist beforehand.


## Steps

| # | Name | Bindings | Try | Catch | Finally |
|:-:|---|:-:|:-:|:-:|:-:|
| 1 | [Create a SecurityIntentBinding](#step-Create a SecurityIntentBinding) | 0 | 1 | 0 | 0 |
| 2 | [Create a SecurityIntent](#step-Create a SecurityIntent) | 0 | 1 | 0 | 0 |
| 3 | [Verity NimbusPolicy creation](#step-Verity NimbusPolicy creation) | 0 | 1 | 0 | 0 |
| 4 | [Verify status of created SecurityIntentBinding](#step-Verify status of created SecurityIntentBinding) | 0 | 1 | 0 | 0 |
| 5 | [Verify status of created SecurityIntent](#step-Verify status of created SecurityIntent) | 0 | 1 | 0 | 0 |
| 6 | [Verify status of created NimbusPolicy](#step-Verify status of created NimbusPolicy) | 0 | 1 | 0 | 0 |

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

### Step: `Verity NimbusPolicy creation`

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

### Step: `Verify status of created SecurityIntent`

*No description*

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

---

