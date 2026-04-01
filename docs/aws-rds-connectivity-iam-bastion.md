# AWS RDS connectivity: IAM auth, bastion, Lambda, and migrations

This guide ties together how **private RDS** in a VPC is reached, how **IAM database authentication** fits the **homehub-profiles** Go code, and how a **bastion host** (or Lambda in the VPC) avoids exposing Postgres to the public internet.

For provisioning RDS with Terraform, see [getting-started-terraform-rds.md](./getting-started-terraform-rds.md). For Atlas migrations in this repo, see [documentation/migrations.md](../documentation/migrations.md).

---

## 1. Why your laptop cannot reach a private RDS

- RDS in a **private subnet** gets a **private IP** (for example `172.31.x.x`). DNS for the RDS endpoint resolves to that address inside the VPC.
- Your **home machine** and **AWS CloudShell** (by default) are **not** in that VPC, so there is **no route** to that IP. You typically see **`connect: operation timed out`** before any username/password or IAM token is validated.
- An **Internet Gateway (IGW)** alone does **not** fix that: IGW is for **public** subnets and **public** endpoints. Private RDS is intentionally **not** on a public path.

**Ways to get application or operator traffic to private RDS:**

| Approach | Use case |
|----------|-----------|
| **Bastion EC2** (or SSM port-forward) | You, DBeaver, `psql`, Atlas CLI from your laptop |
| **Lambda / ECS / EC2 in the same VPC** | Production API (for example this service built for Lambda) |
| **VPN or Direct Connect** | Corporate / site-to-site access to the VPC |
| **Public RDS** (dev only) | Simple but wider blast radius; needs public subnet, IGW route, strict security groups |

---

## 2. RDS IAM database authentication (DB user + AWS IAM)

The app in this repo may use the AWS SDK to build a **short-lived IAM auth token** and pass it as the Postgres **password** (see `src/database/connection.go`). That is **RDS IAM database authentication**, not the same thing as only creating an “IAM user” in the AWS console—though an IAM **user** or **role** must be allowed to call **`rds-db:connect`**.

### 2.1 Enable IAM auth on the RDS instance

- In the AWS console: modify the instance and enable **IAM DB authentication**, or in Terraform set **`iam_database_authentication_enabled = true`** on `aws_db_instance`.

### 2.2 Create a database user in PostgreSQL

Connect as a privileged user (often the **master** user) and create a user that may authenticate via IAM:

```sql
CREATE USER homehub_iam;
GRANT rds_iam TO homehub_iam;
-- Grant schema/table privileges your app needs, for example:
-- GRANT CONNECT ON DATABASE yourdb TO homehub_iam;
-- GRANT USAGE ON SCHEMA public TO homehub_iam;
-- GRANT SELECT, INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public TO homehub_iam;
```

The **Postgres user name** (here `homehub_iam`) must match what you pass as `POSTGRES_USER` and what you encode in the IAM policy below.

### 2.3 IAM policy: `rds-db:connect`

Attach a policy to the identity that obtains credentials at runtime:

- **Lambda**: execution **role** of the function.
- **Your laptop**: IAM **user** or **role** used by `aws configure` / SSO / environment variables when the Go app or AWS CLI runs.

Minimal permission:

```json
{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Action": "rds-db:connect",
      "Resource": "arn:aws:rds-db:REGION:ACCOUNT_ID:dbuser:DB_RESOURCE_ID/POSTGRES_USERNAME"
    }
  ]
}
```

Replace:

- **`REGION`**, **`ACCOUNT_ID`** — your AWS region and 12-digit account id.
- **`DB_RESOURCE_ID`** — the RDS **DbiResourceId** (in the console: RDS instance → **Configuration** → **Resource ID**). This is **not** the same as the DB **identifier** string.
- **`POSTGRES_USERNAME`** — the PostgreSQL user name (for example `homehub_iam`).

### 2.4 Application environment variables

Align with `src/config/database.go`:

| Variable | Purpose |
|----------|---------|
| `POSTGRES_HOST` | RDS endpoint hostname |
| `POSTGRES_PORT` | `5432` |
| `POSTGRES_USER` | IAM-enabled Postgres user name |
| `POSTGRES_DB` | Database name |
| `AWS_REGION` or `AWS_DEFAULT_REGION` | Same region as RDS |
| `AWS_PROFILE` | Optional; local dev only |

Lambda uses the **execution role** automatically; do not rely on long-lived access keys on the function.

### 2.5 Security note

Do **not** log IAM auth tokens or full DSNs in production. Tokens are credentials and expire after about **15 minutes**.

---

## 3. Bastion EC2: create and use

A **bastion** is a small EC2 in a **public subnet** (route **`0.0.0.0/0` → Internet Gateway**). You SSH (or use SSM) to the bastion; from there the VPC can reach **RDS on 5432**. RDS stays private; only the bastion exposes a small surface (SSH or SSM) to you.

### 3.1 Before you launch (VPC, public subnet, key pair, security groups)

Do this planning **once** per environment so the bastion can receive SSH from you and forward traffic to RDS without misconfiguring routes or security groups.

#### Same VPC as RDS

Traffic from the bastion to RDS stays **inside the VPC** (private IPv4 to private IPv4). The bastion and RDS must live in the **same VPC** so that networking and security groups work as expected.

**Find the VPC your RDS uses (console):**

1. Open **RDS** → **Databases** → select your instance.
2. In the **Connectivity & security** section, note **VPC** (for example `vpc-0abc1234…`). That is the VPC you must choose when launching the bastion.

**Why “default VPC” is not enough:** The **default VPC** is just one VPC in the account. If RDS was created in a **custom VPC** or another region’s VPC, picking “default” in the launch wizard will put the bastion in the **wrong** VPC—you would never reach RDS on layer 3, and security-group references would be invalid across VPCs.

**Same VPC, different subnets is normal:** RDS is usually in **private** subnets; the bastion goes in a **public** subnet. Both subnets must belong to the **same VPC ID** you copied from RDS.

---

#### Confirm the subnet is public (route table)

A subnet is **public** (for IPv4) when instances with a public IP can send and receive traffic **to/from the internet** because the subnet’s associated **route table** sends default traffic to an **Internet Gateway**.

**What you are verifying:** There is a route **`0.0.0.0/0`** → **`igw-…`** (target type *Internet Gateway*), and that route table is **associated with the subnet** where you will place the bastion.

**Console: inspect a subnet**

1. **VPC** → **Subnets** → filter by the **VPC ID** from RDS.
2. Select a candidate subnet (often one labeled or named with “public” in Terraform/your design).
3. Open the **Route table** tab (or note the **Route table ID** and open **VPC** → **Route tables**).
4. In the **Routes** sub-tab, look for:
   - **`0.0.0.0/0`** → target **`igw-xxxxxxxx`** (Internet Gateway).

If **`0.0.0.0/0`** points to **`nat-…`** (NAT Gateway) or **there is no `0.0.0.0/0`**, that subnet is **not** a classic public subnet for a bastion you SSH to from the internet. Pick another subnet in the **same VPC** that uses a route table with **`0.0.0.0/0` → IGW**.

**Private subnet (contrast):** Often has **`0.0.0.0/0` → NAT** for outbound-only internet, or no default route. Fine for Lambda/RDS; **not** where you place an internet-SSH bastion unless you use **SSM** only and never open port 22 (see §3.6).

**Subnet must be in an AZ that can reach RDS:** Any subnet in the same VPC can route to RDS private IPs if the VPC is set up normally. If your RDS subnet group only covers certain AZs, you still only need **one** bastion subnet in the VPC; SG rules control access, not AZ matching.

---

#### Internet Gateway checklist

The **Internet Gateway** is attached to the **VPC**, not to a single subnet. The **subnet** becomes “public” when its **route table** sends **`0.0.0.0/0`** to that IGW.

**Console: confirm IGW exists for the VPC**

1. **VPC** → **Internet gateways** → filter by **VPC** = your RDS VPC.
2. You should see **`igw-…`** in state **Attached**.

If there is no IGW, a subnet cannot be internet-public in the usual sense; you would need to **create and attach** an IGW and **add** **`0.0.0.0/0` → igw-…** to the bastion subnet’s route table (typical Terraform/console VPC work—outside the scope of a single EC2 launch).

---

#### EC2 key pair (for SSH)

SSH to the bastion uses **asymmetric keys**: AWS stores the **public** key on the instance at launch; you keep the **private** **`.pem`** file on your laptop.

**Create a key pair (console)**

1. **EC2** → **Key pairs** (under **Network & Security** in the left nav).
2. **Create key pair**.
3. **Name**: e.g. `bastion-dev-eu-north-1`.
4. **Key pair type**: **RSA** (widely compatible) or **ED25519** if your SSH client supports it.
5. **Private key file format**: **`.pem`** for OpenSSH.
6. **Create** — the browser downloads **one** private key file. **You cannot download it again.**

**Important details**

- **Region:** Key pairs are **per Region**. Create or import the key in the **same Region** as the bastion and RDS.
- **Permissions on your machine:** `chmod 400 your-key.pem` before `ssh -i …`.
- **Whoever has the `.pem` can SSH** as the default user if the security group allows their IP—protect the file like a password.

If you use **Session Manager only** and never SSH, a key pair is optional; the instance still needs an IAM instance profile for SSM (§3.6).

---

#### RDS security group: what to note and how to allow the bastion

RDS inbound rules control **who** may open TCP **5432** to the database **ENI**. For a bastion pattern, the best practice is **source = bastion security group**, not your home IP on **5432** (you are not connecting from home to RDS directly—you connect to **localhost** after the tunnel, and the **bastion** opens the connection to RDS from **inside** the VPC).

**Find the RDS security group ID (console)**

1. **RDS** → **Databases** → your instance → **Connectivity & security**.
2. Under **VPC security groups**, click the group (for example `sg-rds-…`). The ID looks like **`sg-0123456789abcdef0`**. Keep this page handy—you will edit **Inbound rules**.

**Recommended order of operations**

1. **Create the bastion security group first** (you can do this before launching EC2): **EC2** → **Security Groups** → **Create security group** — VPC = RDS VPC, name e.g. `bastion-sg`, description “SSH from my IP”. Add **inbound** SSH **22** from **My IP** (or a specific `/32`). **Create**.
2. **Update the RDS security group**: **Inbound rules** → **Edit** → **Add rule**: **Type** PostgreSQL, **Port** `5432`, **Source** = **Custom** → pick **`bastion-sg`** (the security group ID, not a CIDR). Save. Now **any ENI using `bastion-sg`** may connect to RDS on 5432.
3. **Launch the bastion EC2** and assign **`bastion-sg`** as its security group (§3.2).

That way you never temporarily open **5432** to `0.0.0.0/0`. If you already launched the bastion, create or identify its SG and add the same RDS inbound rule referencing **`sg-…`** of the bastion.

**Outbound on the bastion SG:** Default **allow all outbound** is enough for the bastion to initiate TCP to RDS **5432** on a private IP.

**Optional diagram**

```text
Internet ──IGW──► public subnet ──► bastion (ENI in bastion-sg)
                                        │
                                        └──► RDS :5432 (RDS SG allows source = bastion-sg)
```

#### Optional: verify routes with AWS CLI

After you know the **subnet ID** you intend to use for the bastion:

```bash
aws ec2 describe-route-tables \
  --filters "Name=association.subnet-id,Values=subnet-YOUR_PUBLIC_SUBNET_ID" \
  --query "RouteTables[0].Routes" \
  --output table
```

Confirm one route has **`DestinationCidrBlock`** (or IPv6 equivalent) for **`0.0.0.0/0`** and **`GatewayId`** starting with **`igw-`**. Replace the subnet ID with your Region’s credentials/profile as usual (`AWS_PROFILE`, `--region`).

---

### 3.2 Launch the instance (console sketch)

1. **EC2** → **Launch instance** — name e.g. `bastion-dev`.
2. **AMI**: Amazon Linux 2023 (or Ubuntu).
3. **Instance type**: `t4g.nano` or `t3.micro`.
4. **Key pair**: your key (for SSH).
5. **Network**: correct **VPC**, **public subnet**, **auto-assign public IP** enabled (or attach an **Elastic IP** after launch).
6. **Security group** (e.g. `bastion-sg`):
   - **Inbound**: **SSH (22)** from **My IP** only (avoid `0.0.0.0/0`).
   - **Outbound**: default is usually fine for dev.

### 3.3 RDS security group (checklist)

If you followed §3.1, the RDS inbound rule is already in place: **PostgreSQL / 5432 / source = bastion security group**. After launch, confirm on **RDS** → your instance → **VPC security groups** → **Inbound rules** that **5432** lists the bastion’s **`sg-…`** and not an open CIDR.

### 3.4 SSH local port forward to RDS

On your laptop:

```bash
chmod 400 /path/to/your-key.pem
ssh -i /path/to/your-key.pem -N -L 5433:YOUR_RDS_ENDPOINT:5432 ec2-user@BASTION_PUBLIC_IP
```

- **Amazon Linux 2023**: `ec2-user`. **Ubuntu**: `ubuntu`.
- **`YOUR_RDS_ENDPOINT`**: hostname from RDS (no `https`).

Then point tools at **`127.0.0.1:5433`**:

- **DBeaver** / **psql**: host `127.0.0.1`, port `5433`, SSL as required (`require` for RDS).
- **Atlas** (see [migrations.md](../documentation/migrations.md)): use a URL with `127.0.0.1:5433`. If you use **IAM auth** for migrations, generate a token (CLI or SDK) and use it as the password; tokens expire in ~15 minutes.

### 3.5 Troubleshooting

| Symptom | Likely cause |
|---------|----------------|
| SSH times out | Instance not in public subnet, no public IP, or SG blocks **22** from your IP. |
| Tunnel up, DB times out | RDS SG does not allow **5432** from **bastion SG**, or wrong endpoint. |
| Connection OK, auth fails | DB user, password, or IAM policy / token — not the bastion path. |

### 3.6 Alternative: Session Manager

Use **AWS Systems Manager Session Manager** with **port forwarding** so you do not open **22** to the internet. The bastion (or any managed instance in the VPC) needs an instance profile with **`AmazonSSMManagedInstanceCore`** and outbound connectivity to SSM endpoints (often via **VPC interface endpoints** or a **NAT gateway** in private subnets).

### 3.7 Cost and hygiene

- Stop the bastion when idle to save cost.
- Prefer **SG-to-SG** rules for RDS, not open CIDRs.

---

## 4. Lambda in the VPC (no bastion for the API)

For the API itself, run **Lambda in the same VPC** as RDS (private subnets are normal). The execution role needs:

- **`AWSLambdaVPCAccessExecutionRole`** (or equivalent) for ENIs.
- **`rds-db:connect`** for your IAM DB user (same ARN pattern as in §2.3).

Security group: **Lambda SG** → **RDS SG** on **5432**. No IGW is required for **Lambda → RDS** traffic inside the VPC.

Build and package the Go binary as **`bootstrap`** for `provided.al2` / `provided.al2023`; see `scripts/build-lambda.sh` and `lambda/main.go`.

---

## 5. Internet Gateway: when it matters

- **Bastion in a public subnet**: the subnet’s route table needs **`0.0.0.0/0` → IGW** so **you** can reach the bastion.
- **Private RDS + Lambda**: traffic stays inside the VPC; **IGW is not** what connects Lambda to RDS.
- **Publicly accessible RDS** (discouraged for production): public subnet, IGW route, and tight **5432** rules are all part of that pattern; private-only RDS does not use that path.

---

## 6. Migrations from your laptop through a bastion

1. Start the SSH tunnel (§3.4).
2. Use `postgres://...@127.0.0.1:5433/...` (or IAM token as password if applicable).
3. Run Atlas from the repo root per [migrations.md](../documentation/migrations.md).

For teams, running **`atlas migrate apply`** from **CodeBuild in the VPC** often replaces ad-hoc laptop tunnels.

---

## Further reading

- [IAM database authentication for RDS for PostgreSQL](https://docs.aws.amazon.com/AmazonRDS/latest/UserGuide/UsingWithRDS.IAMDBAuth.html)
- [AWS Lambda VPC networking](https://docs.aws.amazon.com/lambda/latest/dg/configuration-vpc.html)
- [Session Manager port forwarding](https://docs.aws.amazon.com/systems-manager/latest/userguide/session-manager-working-with-sessions-start.html#start-port-forwarding)
