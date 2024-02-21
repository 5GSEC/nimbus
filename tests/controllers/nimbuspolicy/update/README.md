# Test: `nimbuspolicy-update`

This test validates that direct updates to a NimbusPolicy resource are ignored, to maintain consistency and  prevent unintended modifications.


### Steps

| # | Name | Try | Catch | Finally |
|:-:|---|:-:|:-:|:-:|
| 1 | [Create a SecurityIntent](#step-Create a SecurityIntent) | 1 | 0 | 0 |
| 2 | [Create a SecurityIntentBinding](#step-Create a SecurityIntentBinding) | 1 | 0 | 0 |
| 3 | [Verity NimbusPolicy creation](#step-Verity NimbusPolicy creation) | 1 | 0 | 0 |
| 4 | [Update existing NimbusPolicy](#step-Update existing NimbusPolicy) | 1 | 0 | 0 |
| 5 | [Verify discarding of changes to NimbusPolicy](#step-Verify discarding of changes to NimbusPolicy) | 1 | 0 | 0 |

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

## Step: `Update existing NimbusPolicy`

*No description*

### Try

| # | Operation | Description |
|:-:|---|---|
| 1 | `apply` | *No description* |

## Step: `Verify discarding of changes to NimbusPolicy`

*No description*

### Try

| # | Operation | Description |
|:-:|---|---|
| 1 | `assert` | *No description* |
