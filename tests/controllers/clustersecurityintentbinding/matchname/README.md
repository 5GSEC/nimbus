# Test: `match-names-add-csib`

1. Adds a csib for matchNames with dev, staging 2. Delete staging ns 3. Update csib for matchNames with dev, prod


## Steps

| # | Name | Bindings | Try | Catch | Finally |
|:-:|---|:-:|:-:|:-:|:-:|
| 1 | [Create prod, dev, staging Namespaces](#step-Create prod, dev, staging Namespaces) | 0 | 1 | 0 | 0 |
| 2 | [Create a SecurityIntent](#step-Create a SecurityIntent) | 0 | 1 | 0 | 0 |
| 3 | [Create a ClusterSecurityIntentBinding](#step-Create a ClusterSecurityIntentBinding) | 0 | 1 | 0 | 0 |
| 4 | [Verify ClusterNimbusPolicy creation](#step-Verify ClusterNimbusPolicy creation) | 0 | 1 | 0 | 0 |
| 5 | [Verify NimbusPolicy creation in dev namespace](#step-Verify NimbusPolicy creation in dev namespace) | 0 | 1 | 0 | 0 |
| 6 | [Verify NimbusPolicy creation in staging namespace](#step-Verify NimbusPolicy creation in staging namespace) | 0 | 1 | 0 | 0 |
| 7 | [Verify NimbusPolicy absence in prod](#step-Verify NimbusPolicy absence in prod) | 0 | 1 | 0 | 0 |
| 8 | [Delete staging namespace](#step-Delete staging namespace) | 0 | 1 | 0 | 0 |
| 9 | [Verify the NimbusPolicy deletion in staging](#step-Verify the NimbusPolicy deletion in staging) | 0 | 1 | 0 | 0 |
| 10 | [Verify ClusterNimbusPolicy presence](#step-Verify ClusterNimbusPolicy presence) | 0 | 1 | 0 | 0 |
| 11 | [Verify NimbusPolicy presence in dev namespace](#step-Verify NimbusPolicy presence in dev namespace) | 0 | 1 | 0 | 0 |
| 12 | [Verify NimbusPolicy absence in prod](#step-Verify NimbusPolicy absence in prod) | 0 | 1 | 0 | 0 |
| 13 | [Update the ClusterSecurityIntentBinding to add prod](#step-Update the ClusterSecurityIntentBinding to add prod) | 0 | 1 | 0 | 0 |
| 14 | [Verify ClusterNimbusPolicy presence](#step-Verify ClusterNimbusPolicy presence) | 0 | 1 | 0 | 0 |
| 15 | [Verify NimbusPolicy presence in dev namespace](#step-Verify NimbusPolicy presence in dev namespace) | 0 | 1 | 0 | 0 |
| 16 | [Verify NimbusPolicy presence in prod namespace](#step-Verify NimbusPolicy presence in prod namespace) | 0 | 1 | 0 | 0 |

### Step: `Create prod, dev, staging Namespaces`

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

### Step: `Create a ClusterSecurityIntentBinding`

*No description*

#### Try

| # | Operation | Bindings | Outputs | Description |
|:-:|---|:-:|:-:|---|
| 1 | `apply` | 0 | 0 | *No description* |

### Step: `Verify ClusterNimbusPolicy creation`

*No description*

#### Try

| # | Operation | Bindings | Outputs | Description |
|:-:|---|:-:|:-:|---|
| 1 | `assert` | 0 | 0 | *No description* |

### Step: `Verify NimbusPolicy creation in dev namespace`

*No description*

#### Try

| # | Operation | Bindings | Outputs | Description |
|:-:|---|:-:|:-:|---|
| 1 | `assert` | 0 | 0 | *No description* |

### Step: `Verify NimbusPolicy creation in staging namespace`

*No description*

#### Try

| # | Operation | Bindings | Outputs | Description |
|:-:|---|:-:|:-:|---|
| 1 | `assert` | 0 | 0 | *No description* |

### Step: `Verify NimbusPolicy absence in prod`

*No description*

#### Try

| # | Operation | Bindings | Outputs | Description |
|:-:|---|:-:|:-:|---|
| 1 | `script` | 0 | 0 | *No description* |

### Step: `Delete staging namespace`

*No description*

#### Try

| # | Operation | Bindings | Outputs | Description |
|:-:|---|:-:|:-:|---|
| 1 | `delete` | 0 | 0 | *No description* |

### Step: `Verify the NimbusPolicy deletion in staging`

*No description*

#### Try

| # | Operation | Bindings | Outputs | Description |
|:-:|---|:-:|:-:|---|
| 1 | `script` | 0 | 0 | *No description* |

### Step: `Verify ClusterNimbusPolicy presence`

*No description*

#### Try

| # | Operation | Bindings | Outputs | Description |
|:-:|---|:-:|:-:|---|
| 1 | `assert` | 0 | 0 | *No description* |

### Step: `Verify NimbusPolicy presence in dev namespace`

*No description*

#### Try

| # | Operation | Bindings | Outputs | Description |
|:-:|---|:-:|:-:|---|
| 1 | `assert` | 0 | 0 | *No description* |

### Step: `Verify NimbusPolicy absence in prod`

*No description*

#### Try

| # | Operation | Bindings | Outputs | Description |
|:-:|---|:-:|:-:|---|
| 1 | `script` | 0 | 0 | *No description* |

### Step: `Update the ClusterSecurityIntentBinding to add prod`

*No description*

#### Try

| # | Operation | Bindings | Outputs | Description |
|:-:|---|:-:|:-:|---|
| 1 | `apply` | 0 | 0 | *No description* |

### Step: `Verify ClusterNimbusPolicy presence`

*No description*

#### Try

| # | Operation | Bindings | Outputs | Description |
|:-:|---|:-:|:-:|---|
| 1 | `assert` | 0 | 0 | *No description* |

### Step: `Verify NimbusPolicy presence in dev namespace`

*No description*

#### Try

| # | Operation | Bindings | Outputs | Description |
|:-:|---|:-:|:-:|---|
| 1 | `assert` | 0 | 0 | *No description* |

### Step: `Verify NimbusPolicy presence in prod namespace`

*No description*

#### Try

| # | Operation | Bindings | Outputs | Description |
|:-:|---|:-:|:-:|---|
| 1 | `assert` | 0 | 0 | *No description* |

---

