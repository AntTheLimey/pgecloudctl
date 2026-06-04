# cloud-accounts

Manage pgEdge Cloud accounts — the AWS, Azure, or GCP credentials that
pgEdge Cloud uses to provision infrastructure in your cloud. Alias: `ca`.

## Subcommands

### list

List all cloud accounts registered in the current tenant.

**Usage:** `pgecloudctl cloud-accounts list [flags]`

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `-h, --help` | No | help for list |

**Example:**

```bash
pgecloudctl cloud-accounts list
```

**Example output (table):**

```
ID                                    NAME          TYPE    STATUS
c3d4e5f6-a7b8-9012-cdef-123456789012  prod-aws      aws     active
d4e5f6a7-b8c9-0123-defa-234567890123  prod-azure    azure   active
```

---

### get

Get details for a specific cloud account.

**Usage:** `pgecloudctl cloud-accounts get <id> [flags]`

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `-h, --help` | No | help for get |

**Example:**

```bash
pgecloudctl cloud-accounts get c3d4e5f6-a7b8-9012-cdef-123456789012
```

**Example output (table):**

```
FIELD         VALUE
ID            c3d4e5f6-a7b8-9012-cdef-123456789012
Name          prod-aws
Type          aws
Role ARN      arn:aws:iam::123456789012:role/pgEdgeCloudRole
Status        active
Description   Production AWS account
Created       2026-01-05T08:00:00Z
```

---

### create

Register a new cloud account for AWS, Azure, or GCP. The required flags
depend on the cloud provider specified with `--type`.

**Usage:** `pgecloudctl cloud-accounts create [flags]`

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `--name string` | Yes | Display name for the cloud account |
| `--type string` | Yes | Cloud provider type: aws, azure, or gcp |
| `--description string` | No | Optional description |
| `--role-arn string` | aws | AWS IAM Role ARN (required for --type aws) |
| `--project-id string` | gcp | GCP project ID (required for --type gcp) |
| `--service-account string` | gcp | GCP service account email (required for --type gcp) |
| `--azure-client-id string` | azure | Azure client/application ID (required for --type azure) |
| `--azure-client-secret string` | azure | Azure client secret (required for --type azure) |
| `--subscription-id string` | azure | Azure subscription ID (required for --type azure) |
| `--tenant-id string` | azure | Azure tenant ID (required for --type azure) |
| `--resource-group string` | No | Azure resource group (optional for --type azure) |
| `-h, --help` | No | help for create |

**Example (AWS):**

```bash
pgecloudctl cloud-accounts create \
    --name prod-aws \
    --type aws \
    --role-arn arn:aws:iam::123456789012:role/pgEdgeCloudRole \
    --description "Production AWS account"
```

**Example (GCP):**

```bash
pgecloudctl cloud-accounts create \
    --name prod-gcp \
    --type gcp \
    --project-id my-gcp-project-123 \
    --service-account pgcloud@my-gcp-project-123.iam.gserviceaccount.com
```

**Example output (table):**

```
FIELD     VALUE
ID        c3d4e5f6-a7b8-9012-cdef-123456789012
Name      prod-aws
Type      aws
Status    active
```

---

### delete

Delete a cloud account by ID.

**Usage:** `pgecloudctl cloud-accounts delete <id> [flags]`

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `-y, --yes` | No | Skip confirmation prompt |
| `-h, --help` | No | help for delete |

**Example:**

```bash
pgecloudctl cloud-accounts delete c3d4e5f6-a7b8-9012-cdef-123456789012 --yes
```

**Example output (table):**

```
Cloud account c3d4e5f6-a7b8-9012-cdef-123456789012 deleted.
```

---

### cloudformation-template

Print the AWS CloudFormation template used to create the IAM role that
pgEdge Cloud requires. Apply this template in your AWS account before
registering an AWS cloud account.

**Usage:** `pgecloudctl cloud-accounts cloudformation-template [flags]`

**Flags:**

| Flag | Required | Description |
|------|----------|-------------|
| `-h, --help` | No | help for cloudformation-template |

**Example:**

```bash
pgecloudctl cloud-accounts cloudformation-template
```

**Example output:**

```
AWSTemplateFormatVersion: "2010-09-09"
Description: pgEdge Cloud IAM role
Resources:
  pgEdgeCloudRole:
    Type: AWS::IAM::Role
    Properties:
      ...
```
