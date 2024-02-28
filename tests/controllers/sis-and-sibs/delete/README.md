# Test: `securityintent-deletion-after-creation-of-nimbuspolicy`

This test verifies that when a SecurityIntent is the only one referenced by a SecurityIntentBinding, and that  SecurityIntent is deleted, the corresponding NimbusPolicy is also automatically deleted.


### Steps

| # | Name | Try | Catch | Finally |
|:-:|---|:-:|:-:|:-:|
| 1 | [Create a SecurityIntentBinding](#step-Create a SecurityIntentBinding) | 1 | 0 | 0 |
| 2 | [Create a SecurityIntent](#step-Create a SecurityIntent) | 1 | 0 | 0 |
| 3 | [Verify NimbusPolicy creation](#step-Verify NimbusPolicy creation) | 1 | 0 | 0 |
| 4 | [Delete previously created SecurityIntent](#step-Delete previously created SecurityIntent) | 1 | 0 | 0 |
| 5 | [Verify the NimbusPolicy deletion](#step-Verify the NimbusPolicy deletion) | 1 | 0 | 0 |
| 6 | [Verify status of SecurityIntentBinding](#step-Verify status of SecurityIntentBinding) | 1 | 0 | 0 |

## Step: `Create a SecurityIntentBinding`

*No description*

### Try

| # | Operation | Description |
|:-:|---|---|
| 1 | `apply` | *No description* |

## Step: `Create a SecurityIntent`

*No description*

### Try

| # | Operation | Description |
|:-:|---|---|
| 1 | `apply` | *No description* |

## Step: `Verify NimbusPolicy creation`

*No description*

### Try

| # | Operation | Description |
|:-:|---|---|
| 1 | `assert` | *No description* |

## Step: `Delete previously created SecurityIntent`

*No description*

### Try

| # | Operation | Description |
|:-:|---|---|
| 1 | `delete` | *No description* |

## Step: `Verify the NimbusPolicy deletion`

*No description*

### Try

| # | Operation | Description |
|:-:|---|---|
| 1 | `script` | *No description* |

## Step: `Verify status of SecurityIntentBinding`

This verifies that upon deletion of a NimbusPolicy, the corresponding SecurityIntentBinding's status subresource is updated to reflect the current information.


### Try

| # | Operation | Description |
|:-:|---|---|
| 1 | `assert` | *No description* |
