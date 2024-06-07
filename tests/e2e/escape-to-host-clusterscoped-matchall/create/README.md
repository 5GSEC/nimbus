# Test: `kyverno-adapter-policy-creation`

This test validates that creating a `escapeToHost` SecurityIntent with SecurityIntentBinding generates the expected Kyverno Policy.


## Steps

| # | Name | Bindings | Try | Catch | Finally |
|:-:|---|:-:|:-:|:-:|:-:|
| 1 | [Create a SecurityIntent](#step-Create a SecurityIntent) | 0 | 1 | 0 | 0 |
| 2 | [Create a SecurityIntentBinding](#step-Create a SecurityIntentBinding) | 0 | 1 | 0 | 0 |
| 3 | [Verify NimbusPolicy creation](#step-Verify NimbusPolicy creation) | 0 | 1 | 0 | 0 |
| 4 | [Verify KyvernoPolicy creation](#step-Verify KyvernoPolicy creation) | 0 | 1 | 0 | 0 |
| 5 | [Verify status of created SecurityIntentBinding](#step-Verify status of created SecurityIntentBinding) | 0 | 1 | 0 | 0 |
| 6 | [Verify status of created NimbusPolicy](#step-Verify status of created NimbusPolicy) | 0 | 1 | 0 | 0 |
| 7 | [Verify that the corresponding NimbusPolicy status has been updated with the generated Kyverno Policy](#step-Verify that the corresponding NimbusPolicy status has been updated with the generated Kyverno Policy) | 0 | 1 | 0 | 0 |

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

### Step: `Verify KyvernoPolicy creation`

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

### Step: `Verify that the corresponding NimbusPolicy status has been updated with the generated Kyverno Policy`

*No description*

#### Try

| # | Operation | Bindings | Outputs | Description |
|:-:|---|:-:|:-:|---|
| 1 | `script` | 0 | 0 | *No description* |

---

