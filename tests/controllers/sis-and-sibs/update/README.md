# Test: `update`

This test verifies that modifying a SecurityIntent triggers the desired updates in corresponding SecurityIntentBinding's  status subresource and related NimbusPolicy.


### Steps

| # | Name | Try | Catch | Finally |
|:-:|---|:-:|:-:|:-:|
| 1 | [Create a SecurityIntentBinding for multiple SecurityIntents](#step-Create a SecurityIntentBinding for multiple SecurityIntents) | 1 | 0 | 0 |
| 2 | [Create multiple SecurityIntents](#step-Create multiple SecurityIntents) | 1 | 0 | 0 |
| 3 | [Verity NimbusPolicy creation](#step-Verity NimbusPolicy creation) | 1 | 0 | 0 |
| 4 | [Update one SecurityIntent](#step-Update one SecurityIntent) | 1 | 0 | 0 |
| 5 | [Verify NimbusPolicy update](#step-Verify NimbusPolicy update) | 1 | 0 | 0 |
| 6 | [Update SecurityIntentBinding to remove one SecurityIntent](#step-Update SecurityIntentBinding to remove one SecurityIntent) | 1 | 0 | 0 |
| 7 | [Verify the NimbusPolicy update after removal of SecurityIntent](#step-Verify the NimbusPolicy update after removal of SecurityIntent) | 1 | 0 | 0 |
| 8 | [Verify status of SecurityIntentBinding after update](#step-Verify status of SecurityIntentBinding after update) | 1 | 0 | 0 |

## Step: `Create a SecurityIntentBinding for multiple SecurityIntents`

*No description*

### Try

| # | Operation | Description |
|:-:|---|---|
| 1 | `apply` | *No description* |

## Step: `Create multiple SecurityIntents`

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

## Step: `Update one SecurityIntent`

Update the action of one of the previously created SecurityIntents

### Try

| # | Operation | Description |
|:-:|---|---|
| 1 | `apply` | *No description* |

## Step: `Verify NimbusPolicy update`

Verify the update of rule.action for corresponding SecurityIntent update

### Try

| # | Operation | Description |
|:-:|---|---|
| 1 | `assert` | *No description* |

## Step: `Update SecurityIntentBinding to remove one SecurityIntent`

Remove one of the previously created SecurityIntents from the SecurityIntentBinding

### Try

| # | Operation | Description |
|:-:|---|---|
| 1 | `apply` | *No description* |

## Step: `Verify the NimbusPolicy update after removal of SecurityIntent`

*No description*

### Try

| # | Operation | Description |
|:-:|---|---|
| 1 | `assert` | *No description* |

## Step: `Verify status of SecurityIntentBinding after update`

This verifies that upon deletion of a NimbusPolicy, the corresponding SecurityIntentBinding's status subresource is updated to reflect the current information.


### Try

| # | Operation | Description |
|:-:|---|---|
| 1 | `assert` | *No description* |
