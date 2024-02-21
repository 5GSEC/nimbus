# Test: `securityintentbinding-creation`

This test validates the automated creation of a NimbusPolicy resource when a corresponding SecurityIntent  and SecurityIntentBinding are created.


### Steps

| # | Name | Try | Catch | Finally |
|:-:|---|:-:|:-:|:-:|
| 1 | [Create a SecurityIntent](#step-Create a SecurityIntent) | 1 | 0 | 0 |
| 2 | [Create a SecurityIntentBinding](#step-Create a SecurityIntentBinding) | 1 | 0 | 0 |
| 3 | [Verity NimbusPolicy creation](#step-Verity NimbusPolicy creation) | 1 | 0 | 0 |
| 4 | [Verify status of created SecurityIntentBinding](#step-Verify status of created SecurityIntentBinding) | 1 | 0 | 0 |
| 5 | [Verify status of created NimbusPolicy](#step-Verify status of created NimbusPolicy) | 1 | 0 | 0 |

## Step: `Create a SecurityIntent`

*No description*

### Try

| # | Operation | Description |
|:-:|---|---|
| 1 | `apply` | *No description* |

## Step: `Create a SecurityIntentBinding`

*No description*

### Try

| # | Operation | Description |
|:-:|---|---|
| 1 | `apply` | *No description* |

## Step: `Verity NimbusPolicy creation`

*No description*

### Try

| # | Operation | Description |
|:-:|---|---|
| 1 | `assert` | *No description* |

## Step: `Verify status of created SecurityIntentBinding`

Verify the created SecurityIntentBinding status subresource includes the number and names of bound intents,  along with the generated NimbusPolicy name.


### Try

| # | Operation | Description |
|:-:|---|---|
| 1 | `assert` | *No description* |

## Step: `Verify status of created NimbusPolicy`

*No description*

### Try

| # | Operation | Description |
|:-:|---|---|
| 1 | `assert` | *No description* |
