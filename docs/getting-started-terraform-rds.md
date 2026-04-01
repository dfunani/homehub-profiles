# Getting started with Terraform and Amazon RDS

This guide walks you through installing Terraform, configuring the AWS provider, and creating a PostgreSQL RDS instance in a safe, repeatable way. It is aimed at **real AWS accounts**. If you only use **LocalStack** for RDS, see [documentation/localstack-rds.md](../documentation/localstack-rds.md)—RDS there typically requires LocalStack Pro; plain Postgres for local dev is usually the Compose `database` service instead.

## Prerequisites

- An **AWS account** with permission to create VPC resources, security groups, RDS instances, and (optionally) subnets.
- **AWS CLI** installed and configured (`aws configure` or environment variables / IAM role).
- A **default VPC** in your target region, or existing **subnet IDs** where RDS should live (RDS needs subnets in at least two Availability Zones for a standard multi-AZ or subnet group setup).

## Install Terraform

Install a recent 1.x release from [Terraform Install](https://developer.hashicorp.com/terraform/install).

Verify:

```bash
terraform version
```

## Project layout (recommended)

Create a dedicated directory for infrastructure (for example next to `homehub-profiles` or in a separate repo):

```text
terraform/
  main.tf
  variables.tf
  terraform.tfvars.example
  outputs.tf
  versions.tf
```

Keep **secrets out of Git**: use `terraform.tfvars` (gitignored), environment variables, or AWS Secrets Manager / SSM Parameter Store.

## Pin provider versions

In `versions.tf`:

```hcl
terraform {
  required_version = ">= 1.5.0"

  required_providers {
    aws = {
      source  = "hashicorp/aws"
      version = "~> 5.0"
    }
  }
}

provider "aws" {
  region = var.aws_region
}
```

## Variables

Example `variables.tf`:

```hcl
variable "aws_region" {
  type    = string
  default = "us-east-1"
}

variable "db_identifier" {
  type    = string
  default = "homehub-dev"
}

variable "db_name" {
  type    = string
  default = "homehub"
}

variable "db_username" {
  type      = string
  sensitive = true
}

variable "db_password" {
  type      = string
  sensitive = true
}

variable "instance_class" {
  type    = string
  default = "db.t4g.micro"
}
```

Create `terraform.tfvars` locally (do not commit):

```hcl
db_username = "dbadmin"
db_password = "<a strong password>"
```

## Minimal RDS (PostgreSQL) example

This example uses the **default VPC** and a simple **DB subnet group** spanning the default subnets. For production, use explicit VPC IDs, private subnets, and tight security groups.

`main.tf` (illustrative—adjust AZs and subnet IDs to your account):

```hcl
data "aws_availability_zones" "available" {
  state = "available"
}

resource "aws_db_subnet_group" "this" {
  name       = "${var.db_identifier}-subnet-group"
  subnet_ids = data.aws_subnets.default.ids
}

data "aws_subnets" "default" {
  filter {
    name   = "vpc-id"
    values = [data.aws_vpc.default.id]
  }
}

data "aws_vpc" "default" {
  default = true
}

resource "aws_security_group" "rds" {
  name        = "${var.db_identifier}-rds-sg"
  description = "RDS access"
  vpc_id      = data.aws_vpc.default.id

  ingress {
    description = "Postgres from your IP or app SG"
    from_port   = 5432
    to_port     = 5432
    protocol    = "tcp"
    cidr_blocks = ["YOUR_IP/32"] # tighten: use a bastion or app security group
  }

  egress {
    from_port   = 0
    to_port     = 0
    protocol    = "-1"
    cidr_blocks = ["0.0.0.0/0"]
  }
}

resource "aws_db_instance" "this" {
  identifier                 = var.db_identifier
  engine                     = "postgres"
  engine_version             = "16"
  instance_class             = var.instance_class
  allocated_storage          = 20
  storage_type               = "gp3"
  db_name                    = var.db_name
  username                   = var.db_username
  password                   = var.db_password
  db_subnet_group_name       = aws_db_subnet_group.this.name
  vpc_security_group_ids     = [aws_security_group.rds.id]
  skip_final_snapshot        = true
  publicly_accessible        = false # set true only if you need direct internet access (not typical)
  backup_retention_period    = 7
  deletion_protection        = false   # set true for production
}
```

Replace `YOUR_IP/32` with your office IP, or attach an **application security group** instead of open CIDR. For production, place RDS in **private subnets** and access via VPN, bastion, or ECS/Lambda in the same VPC.

`outputs.tf`:

```hcl
output "rds_endpoint" {
  value = aws_db_instance.this.endpoint
}

output "rds_address" {
  value = aws_db_instance.this.address
}

output "rds_port" {
  value = aws_db_instance.this.port
}
```

## Initialize and apply

```bash
cd terraform
terraform init
terraform fmt -recursive
terraform validate
terraform plan -out=tfplan
terraform apply tfplan
```

After apply, read the endpoint:

```bash
terraform output rds_endpoint
```

Connection string (conceptually):

```text
postgres://USERNAME:PASSWORD@ADDRESS:5432/DATABASE_NAME?sslmode=require
```

Use **SSL** against RDS in AWS; `sslmode=require` or stricter is typical.

## Destroy (dev only)

```bash
terraform destroy
```

Confirm you are not deleting production data. For `skip_final_snapshot = false`, Terraform will create a final snapshot on destroy if configured.

## State and collaboration

- **Local state** (`terraform.tfstate`) is fine for solo experiments; add `terraform.tfstate*` to `.gitignore`.
- For teams, use a **remote backend** (for example S3 + DynamoDB locking) as described in [HashiCorp’s AWS backend tutorial](https://developer.hashicorp.com/terraform/language/settings/backends/s3).

## Further reading

- [aws_db_instance](https://registry.terraform.io/providers/hashicorp/aws/latest/docs/resources/db_instance) (Terraform AWS provider)
- [Amazon RDS for PostgreSQL](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/CHAP_PostgreSQL.html)
- [Terraform AWS provider](https://registry.terraform.io/providers/hashicorp/aws/latest/docs)

## Relation to this repository

The **homehub-profiles** service’s database URL and migrations are configured in this repo (GORM, Atlas migrations under `migrations/`). Point `DATABASE_URL` at the Terraform-created RDS endpoint after the instance is available and security groups allow your application’s traffic.

For **private RDS**, operator and app access patterns (**IAM DB auth**, **bastion**, **Lambda in VPC**, **migrations**), see [aws-rds-connectivity-iam-bastion.md](./aws-rds-connectivity-iam-bastion.md).
