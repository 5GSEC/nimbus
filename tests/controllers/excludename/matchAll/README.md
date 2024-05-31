# Test: `multiple-ns-add-csib`

1. Create prod, dev, staging ns 2. Add security intent 3. Add csib  4. Delete prod namespace 5. Create prod namespace again


## Steps

| # | Name | Bindings | Try | Catch | Finally |
|:-:|---|:-:|:-:|:-:|:-:|
| 1 | [Create prod, dev, staging Namespaces](#step-Create prod, dev, staging Namespaces) | 0 | 1 | 0 | 0 |
| 2 | [Create a SecurityIntent](#step-Create a SecurityIntent) | 0 | 1 | 0 | 0 |
| 3 | [Create a ClusterSecurityIntentBinding](#step-Create a ClusterSecurityIntentBinding) | 0 | 1 | 0 | 0 |
| 4 | [Verify ClusterNimbusPolicy creation](#step-Verify ClusterNimbusPolicy creation) | 0 | 1 | 0 | 0 |
| 5 | [Verify NimbusPolicy creation in default](#step-Verify NimbusPolicy creation in default) | 0 | 1 | 0 | 0 |
| 6 | [Verify NimbusPolicy creation in dev namespace](#step-Verify NimbusPolicy creation in dev namespace) | 0 | 1 | 0 | 0 |
| 7 | [Verify NimbusPolicy creation in staging namespace](#step-Verify NimbusPolicy creation in staging namespace) | 0 | 1 | 0 | 0 |
| 8 | [Verify NimbusPolicy creation in prod namespace](#step-Verify NimbusPolicy creation in prod namespace) | 0 | 1 | 0 | 0 |
| 9 | [Delete prod namespace](#step-Delete prod namespace) | 0 | 1 | 0 | 0 |
| 10 | [Verify the NimbusPolicy deletion](#step-Verify the NimbusPolicy deletion) | 0 | 1 | 0 | 0 |
| 11 | [Verify ClusterNimbusPolicy presence](#step-Verify ClusterNimbusPolicy presence) | 0 | 1 | 0 | 0 |
| 12 | [Verify NimbusPolicy presence in default](#step-Verify NimbusPolicy presence in default) | 0 | 1 | 0 | 0 |
| 13 | [Verify NimbusPolicy presence in dev namespace](#step-Verify NimbusPolicy presence in dev namespace) | 0 | 1 | 0 | 0 |
| 14 | [Verify NimbusPolicy presence in staging namespace](#step-Verify NimbusPolicy presence in staging namespace) | 0 | 1 | 0 | 0 |
| 15 | [Create prod Namespace again](#step-Create prod Namespace again) | 0 | 1 | 0 | 0 |
| 16 | [Verify ClusterNimbusPolicy presence](#step-Verify ClusterNimbusPolicy presence) | 0 | 1 | 0 | 0 |
| 17 | [Verify NimbusPolicy presence in default](#step-Verify NimbusPolicy presence in default) | 0 | 1 | 0 | 0 |
| 18 | [Verify NimbusPolicy presence in dev namespace](#step-Verify NimbusPolicy presence in dev namespace) | 0 | 1 | 0 | 0 |
| 19 | [Verify NimbusPolicy presence in staging namespace](#step-Verify NimbusPolicy presence in staging namespace) | 0 | 1 | 0 | 0 |
| 20 | [Verify NimbusPolicy creation in prod namespace](#step-Verify NimbusPolicy creation in prod namespace) | 0 | 1 | 0 | 0 |

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

### Step: `Verify NimbusPolicy creation in default`

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

### Step: `Verify NimbusPolicy creation in prod namespace`

*No description*

#### Try

| # | Operation | Bindings | Outputs | Description |
|:-:|---|:-:|:-:|---|
| 1 | `assert` | 0 | 0 | *No description* |

### Step: `Delete prod namespace`

*No description*

#### Try

| # | Operation | Bindings | Outputs | Description |
|:-:|---|:-:|:-:|---|
| 1 | `delete` | 0 | 0 | *No description* |

### Step: `Verify the NimbusPolicy deletion`

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

### Step: `Verify NimbusPolicy presence in default`

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

### Step: `Verify NimbusPolicy presence in staging namespace`

*No description*

#### Try

| # | Operation | Bindings | Outputs | Description |
|:-:|---|:-:|:-:|---|
| 1 | `assert` | 0 | 0 | *No description* |

### Step: `Create prod Namespace again`

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

### Step: `Verify NimbusPolicy presence in default`

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

### Step: `Verify NimbusPolicy presence in staging namespace`

*No description*

#### Try

| # | Operation | Bindings | Outputs | Description |
|:-:|---|:-:|:-:|---|
| 1 | `assert` | 0 | 0 | *No description* |

### Step: `Verify NimbusPolicy creation in prod namespace`

*No description*

#### Try

| # | Operation | Bindings | Outputs | Description |
|:-:|---|:-:|:-:|---|
| 1 | `assert` | 0 | 0 | *No description* |

---

