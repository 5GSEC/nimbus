# Test: `securityintentbinding-creation`

This test validates the automated creation of a NimbusPolicy resource when a corresponding SecurityIntent  and SecurityIntentBinding are created.


## Steps

| # | Name | Bindings | Try | Catch | Finally |
|:-:|---|:-:|:-:|:-:|:-:|
| 1 | [Create a SecurityIntent](#step-Create a SecurityIntent) | 0 | 1 | 0 | 0 |
| 2 | [Create a SecurityIntentBinding](#step-Create a SecurityIntentBinding) | 0 | 1 | 0 | 0 |
| 3 | [Verify NimbusPolicy creation](#step-Verify NimbusPolicy creation) | 0 | 1 | 0 | 0 |
| 4 | [Verify status of created SecurityIntentBinding](#step-Verify status of created SecurityIntentBinding) | 0 | 1 | 0 | 0 |
| 5 | [Verify status of created NimbusPolicy](#step-Verify status of created NimbusPolicy) | 0 | 1 | 0 | 0 |

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

---

