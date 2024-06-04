# Test: `exclude-names-add-csib`

1. Add dev, staging, prod ns 2. Adds a csib for excludeNames with dev, staging 3. Update csib for excludeNames with dev, prod, regional 4. Add a ns with name management - this ns is not part of the excludeNames
   Add a ns with name regional - this is part of the excludeNames


## Steps

| # | Name | Bindings | Try | Catch | Finally |
|:-:|---|:-:|:-:|:-:|:-:|
| 1 | [1. Create prod, dev, staging Namespaces](#step-1. Create prod, dev, staging Namespaces) | 0 | 1 | 0 | 0 |
| 2 | [Create a SecurityIntent](#step-Create a SecurityIntent) | 0 | 1 | 0 | 0 |
| 3 | [2. Create a ClusterSecurityIntentBinding](#step-2. Create a ClusterSecurityIntentBinding) | 0 | 1 | 0 | 0 |
| 4 | [Verify ClusterNimbusPolicy creation](#step-Verify ClusterNimbusPolicy creation) | 0 | 1 | 0 | 0 |
| 5 | [Verify NimbusPolicy absence in dev namespace](#step-Verify NimbusPolicy absence in dev namespace) | 0 | 1 | 0 | 0 |
| 6 | [Verify NimbusPolicy absence in staging namespace](#step-Verify NimbusPolicy absence in staging namespace) | 0 | 1 | 0 | 0 |
| 7 | [Verify NimbusPolicy creation in prod](#step-Verify NimbusPolicy creation in prod) | 0 | 1 | 0 | 0 |
| 8 | [3. Update the ClusterSecurityIntentBinding to have dev, prod, regional in excludeNames](#step-3. Update the ClusterSecurityIntentBinding to have dev, prod, regional in excludeNames) | 0 | 1 | 0 | 0 |
| 9 | [Verify ClusterNimbusPolicy presence](#step-Verify ClusterNimbusPolicy presence) | 0 | 1 | 0 | 0 |
| 10 | [Verify NimbusPolicy absence in dev namespace](#step-Verify NimbusPolicy absence in dev namespace) | 0 | 1 | 0 | 0 |
| 11 | [Verify NimbusPolicy absence in prod namespace](#step-Verify NimbusPolicy absence in prod namespace) | 0 | 1 | 0 | 0 |
| 12 | [4. Create the management, regional namespaces](#step-4. Create the management, regional namespaces) | 0 | 1 | 0 | 0 |
| 13 | [Verify ClusterNimbusPolicy presence](#step-Verify ClusterNimbusPolicy presence) | 0 | 1 | 0 | 0 |
| 14 | [Verify NimbusPolicy absence in dev namespace](#step-Verify NimbusPolicy absence in dev namespace) | 0 | 1 | 0 | 0 |
| 15 | [Verify NimbusPolicy absence in dev namespace](#step-Verify NimbusPolicy absence in dev namespace) | 0 | 1 | 0 | 0 |
| 16 | [Verify NimbusPolicy absence in prod namespace](#step-Verify NimbusPolicy absence in prod namespace) | 0 | 1 | 0 | 0 |
| 17 | [Verify NimbusPolicy absence in regional namespace](#step-Verify NimbusPolicy absence in regional namespace) | 0 | 1 | 0 | 0 |

### Step: `1. Create prod, dev, staging Namespaces`

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

### Step: `2. Create a ClusterSecurityIntentBinding`

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

### Step: `Verify NimbusPolicy absence in dev namespace`

*No description*

#### Try

| # | Operation | Bindings | Outputs | Description |
|:-:|---|:-:|:-:|---|
| 1 | `script` | 0 | 0 | *No description* |

### Step: `Verify NimbusPolicy absence in staging namespace`

*No description*

#### Try

| # | Operation | Bindings | Outputs | Description |
|:-:|---|:-:|:-:|---|
| 1 | `script` | 0 | 0 | *No description* |

### Step: `Verify NimbusPolicy creation in prod`

*No description*

#### Try

| # | Operation | Bindings | Outputs | Description |
|:-:|---|:-:|:-:|---|
| 1 | `assert` | 0 | 0 | *No description* |

### Step: `3. Update the ClusterSecurityIntentBinding to have dev, prod, regional in excludeNames`

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

### Step: `Verify NimbusPolicy absence in dev namespace`

*No description*

#### Try

| # | Operation | Bindings | Outputs | Description |
|:-:|---|:-:|:-:|---|
| 1 | `script` | 0 | 0 | *No description* |

### Step: `Verify NimbusPolicy absence in prod namespace`

*No description*

#### Try

| # | Operation | Bindings | Outputs | Description |
|:-:|---|:-:|:-:|---|
| 1 | `script` | 0 | 0 | *No description* |

### Step: `4. Create the management, regional namespaces`

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

### Step: `Verify NimbusPolicy absence in dev namespace`

*No description*

#### Try

| # | Operation | Bindings | Outputs | Description |
|:-:|---|:-:|:-:|---|
| 1 | `script` | 0 | 0 | *No description* |

### Step: `Verify NimbusPolicy absence in dev namespace`

*No description*

#### Try

| # | Operation | Bindings | Outputs | Description |
|:-:|---|:-:|:-:|---|
| 1 | `script` | 0 | 0 | *No description* |

### Step: `Verify NimbusPolicy absence in prod namespace`

*No description*

#### Try

| # | Operation | Bindings | Outputs | Description |
|:-:|---|:-:|:-:|---|
| 1 | `script` | 0 | 0 | *No description* |

### Step: `Verify NimbusPolicy absence in regional namespace`

*No description*

#### Try

| # | Operation | Bindings | Outputs | Description |
|:-:|---|:-:|:-:|---|
| 1 | `script` | 0 | 0 | *No description* |

---

