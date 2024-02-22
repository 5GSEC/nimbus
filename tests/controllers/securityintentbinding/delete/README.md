# Test: `securityintentbinding-deletion`

This test validates the expected behavior of NimbusPolicy deletion upon the removal of a corresponding SecurityIntentBinding resource.


### Steps

| # | Name | Try | Catch | Finally |
|:-:|---|:-:|:-:|:-:|
| 1 | [Create a SecurityIntent](#step-Create a SecurityIntent) | 1 | 0 | 0 |
| 2 | [Create a SecurityIntentBinding](#step-Create a SecurityIntentBinding) | 1 | 0 | 0 |
| 3 | [Delete existing SecurityIntentBinding](#step-Delete existing SecurityIntentBinding) | 1 | 0 | 0 |
| 4 | [Verify the NimbusPolicy deletion](#step-Verify the NimbusPolicy deletion) | 1 | 0 | 0 |

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

## Step: `Delete existing SecurityIntentBinding`

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
