# Test: `nimbuspolicy-deletion`

This test validates that when a NimbusPolicy is directly deleted, nimbus automatically re-creates the deleted NimbusPolicy or not.


### Steps

| # | Name | Try | Catch | Finally |
|:-:|---|:-:|:-:|:-:|
| 1 | [Create a SecurityIntent](#step-Create a SecurityIntent) | 1 | 0 | 0 |
| 2 | [Create a SecurityIntentBinding](#step-Create a SecurityIntentBinding) | 1 | 0 | 0 |
| 3 | [Verity NimbusPolicy creation](#step-Verity NimbusPolicy creation) | 1 | 0 | 0 |
| 4 | [Delete existing NimbusPolicy](#step-Delete existing NimbusPolicy) | 1 | 0 | 0 |
| 5 | [Verify NimbusPolicy recreation](#step-Verify NimbusPolicy recreation) | 1 | 0 | 0 |

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

## Step: `Delete existing NimbusPolicy`

*No description*

### Try

| # | Operation | Description |
|:-:|---|---|
| 1 | `delete` | *No description* |

## Step: `Verify NimbusPolicy recreation`

*No description*

### Try

| # | Operation | Description |
|:-:|---|---|
| 1 | `assert` | *No description* |
