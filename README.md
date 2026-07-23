# Distributed Workflow & Project Management System (mini jira)

## Project Overview

This is an enterprise-grade, collaborative project and workflow management platform designed for high scalability and resilience.

Unlike traditional task managers, this utilizes a distributed microservices architecture to handle high-concurrency environments and complex business logic.

The core objective of the project is to provide a seamless collaborative experience where teams can manage projects and track tasks. By leveraging Event-Driven Architecture (EDA) and Durable Workflows, the system ensures that no task is ever lost, even in the event of partial system failures.

---

## Key Features

- Workspace & Role Management (Multi-tenant + granular access control)
- Project Task Collaboration
- Automated Workflows for workspace invitation (SLA & Reminders using durable execution)
- Scalable Notification Engine (Email, in-app)
- System Observability (Logging, monitoring, tracing)

---

## Technology Stack

### Backend Languages
- Go (Golang) → Performance-critical services & workers
- Node.js → Auth service & notification engine

### Communication & Orchestration
- gRPC → Internal service communication
- Apache Kafka → Event streaming
- Temporal.io → Durable workflows

### Data & Caching
- PostgreSQL → Projects & tasks (ACID compliance)
- MongoDB → Users, workspaces, notifications
- Redis → Caching, sessions, Pub/Sub

### Infrastructure
- Docker → Containerization
- Kubernetes → Orchestration
- Nginx Ingress → API Gateway

### Observability
- Prometheus → Metrics
- Loki → Logging
- OpenTelemetry → Tracing
- Grafana → Dashboard View

---

## Microservices

### 1. Auth Service (Node.js)

**Responsibility:** Authentication, session management
**Database:** MongoDB + Redis  

#### API

| Method | Endpoint | Description | Auth | Sample Input |
| :--- | :--- | :--- | :--- | :--- |
| POST | `/api/v1/auth/register` | Register user | No | `{"full_name": "test", "email": "test@test.com", "password": "test1234"}` |
| POST | `/api/v1/auth/login` | Login (JWT + Refresh Token) | No | `{"email": "test@test.com", "password": "test1234"}` |
| POST | `/api/v1/auth/logout` | Logout with device id | Yes |  |
| POST | `/api/v1/auth/refresh-token` | Get new access token | No | `{"refreshToken": "eyJhbGc...."}` |
| GET | `/api/v1/users/profile` | Get user profile | Yes |  |
| PUT | `/api/v1/users/profile` | Update user profile | Yes | `{"full_name": "Jane Doe", "bio": "developer"}` |
| GET | `/api/v1/users/sessions` | Get all active sessions | Yes |  |

#### gRPC
| Method | Request | Response | Description |
|--------|---------|----------|-------------|
| `VerifySession` | `token` | `VerifyResponse` | Validates JWT against Redis session & returns user metadata. |

#### Events Produced
- `user-registered`: Triggered when a new user signs up.
- `user-logout`: Triggered when a new user logout from a device.

---

### 2. Workspace, Task & Project Service (Go)

**Responsibility:** Core business logic, workflow orchestration, and localized user profile snapshotting (via Kafka sync).
**Database:** PostgreSQL + Redis

#### API

| Method | Endpoint | Description | Auth | Sample Input |
| :--- | :--- | :--- | :--- | :--- |
| POST | `/api/v1/workspace` | Create workspace | Yes | `{"name":"test","slug":"test","description":"test"}` |
| GET | `/api/v1/workspace/owned?limit=10&cursor=...` | List workspaces (Owner) | Yes | N/A |
| GET | `/api/v1/workspace/joined?limit=10&cursor=...` | List workspaces (Member) | Yes | N/A |
| POST | `/api/v1/workspace/:id/invites` | Invite member | Yes | `{"email":"test@test.com","role":"ADMIN"}` |
| POST | `/api/v1/workspace/invites/accept` | Accept invite | Yes | `{"token":"6fa7dfcd-bfa2-4e13-bdb4-6e7fcb8ee8b5"}` |
| GET | `/api/v1/workspace/:id/members` | Get members | Yes | N/A |
| POST | `/api/v1/workspace/:id/projects` | Create project | Yes | `{"name":"E-Commerce Microservices Backend","description":"This project handles the core ordering and payment workflow systems."}` |
| GET | `/api/v1/workspace/:id/projects?limit=10&cursor=...` | Get projects with cursor pagination | Yes | N/A |
| POST | `/api/v1/workspace/:id/projects/:projectId/tasks` | Create task | Yes | `{"title":"Implement Kafka Event Consumer","description":"Create a robust worker to consume user-registration events from the message queue.","priority":"HIGH","assignee_id":"6a107afad2ac1e59aba88b6f","deadline":"2026-06-15T18:30:00Z"}` |
| GET | `/api/v1/workspace/:id/projects/:projectId/tasks?limit=10&cursor=...&status=TODO` | Get tasks with cursor pagination | Yes | N/A |
| PUT | `/api/v1/workspace/:id/tasks/:taskId` | Update task | Yes | `{"title":"Implement Kafka Event Consumer 2","description":"Create a robust worker to consume user-registration events from the message queue. 2","priority":"HIGH","assignee_id":"6a107afad2ac1e59aba88b6f","deadline":"2026-06-15T18:30:00Z"}` |
| PATCH | `/api/v1/workspace/:id/tasks/:taskId/status` | Update task status | Yes | `{"status":"DONE"}` |
| POST | `/api/v1/workspace/:id/tasks/:taskId/comments` | Add comment | Yes | `{"content":"DONE"}` |
| GET | `/api/v1/workspace/:id/tasks/:taskId/comments` | Fetch comments | Yes | N/A |


#### Events Produced
- `send-notification`: Triggered when need to send a notification.

#### Redis Caching Architecture

##### Why use Redis in Workspace & Project:
- **High-Frequency Read Optimization**: Workspace listings, membership roles, and project lists are highly read-intensive. Redis shields the PostgreSQL database from high query volumes.
- **Lexicographical Pagination**: Allows paginating workspaces and projects chronologically descending directly inside Redis without transferring the entire set of IDs to the Go application memory.
- **Fast Authorization Checks**: Workspace roles are cached to evaluate permissions instantly during API handler authorization.

##### Workspace Caching:
- **Workspace Metadata Cache (`workspace:<workspaceId>:meta`)**: Redis Hash storing core workspace fields with a 24-hour TTL.
- **Lexicographical ZSET Indexing (`user:<userId>:workspaces:owned` and `user:<userId>:workspaces:joined`)**: Stored as Sorted Sets where all members have a score of `0`. Redis sorts them lexicographically. Since IDs are UUIDv7, lexicographical sorting corresponds to chronological sorting.
  - Pagination fetches exactly `limit` IDs using `ZRevRangeByLex` with exclusive boundary offsets (`Max: "(" + cursor`).
- **Workspace Role Cache (`workspace:<workspaceId>:roles`)**: Hash mapping `user_id -> role` for fast permission lookup.
- **Workspace Members Cache (`workspace:<workspaceId>:members`)**: Hash storing JSON strings of `WorkspaceMemberResponse` indexed by `user_id` for quick collection retrieval and single member updates.
- **Consistency & Invalidation**: We follow the Cache-Aside pattern. On creating workspaces or accepting invites, the corresponding ZSET caches are dynamically appended (`ZAdd`) and hashes updated/invalidated to guarantee strong read-after-write consistency.

##### Project Caching:
- **Project Metadata Cache (`project:<projectId>:meta`)**: Redis Hash storing core project fields (`id`, `workspace_id`, `name`, `description`, `status`, `created_by`, `created_at`) with a 24-hour TTL.
- **Workspace Projects ZSET Indexing (`workspace:<workspaceId>:projects`)**: Sorted set storing project IDs in a workspace with score `0`. Redis sorted sets sort members lexicographically (matching UUIDv7 time sorting).
  - Paginated queries fetch specific pages using `ZRevRangeByLex` with cursor boundaries (`Max: "(" + cursor`).
- **Consistency & Invalidation**: We follow the Cache-Aside pattern. On creating a project, the metadata is cached and the ID is added to the ZSET list (`ZAdd`). If a cache miss occurs, the system queries the database (fetching up to 1000 items) to repopulate both the metadata hash and ZSET list. Updates or invalidations delete the list keys from Redis to trigger a reload.

##### Task Caching:
- **Task Metadata Cache (`task:<taskId>:data`)**: Redis Hash storing detailed task fields (`id`, `workspace_id`, `project_id`, `title`, `description`, `status`, `priority`, `assignee_id`, `assignee_name`, `deadline`, `created_at`) with a 3-day TTL.
- **Project Column ZSET Indexing (`project:<projectId>:col:<columnName>`)**: Sorted set storing task IDs inside a project's column.
  - **UUIDv7 Scoring**: Scores are set using the 48-bit millisecond timestamp extracted from the task's UUIDv7 ID, maintaining chronological sorting.
  - **Pruning**: Restricts the column ZSET to only store the latest 100 task IDs using `ZRemRangeByRank` to prune older tasks.
  - **Cursor-based Pagination**: Fetches paginated tasks using `ZRevRangeByScore` based on the score/timestamp extracted from the cursor task ID (`Max: "(" + cursorScore`).
  - **Empty Column Caching**: Inserts a dummy `__empty__` member with score `-1` when a column is empty, allowing cache hits for empty columns and avoiding database roundtrips.
- **Consistency & Invalidation**: We follow the Cache-Aside pattern. Creating or editing a task populates the metadata Hash and ZSET (if it exists) with a 3-day TTL. Moving a task's column (PATCH status) triggers an atomic Redis transaction (`TxPipeline` / `MULTI`/`EXEC` block) to update the status in the metadata Hash, remove the task ID from the old column ZSET, add it to the new column ZSET, and prune the new ZSET.

#### Kafka Architecture

The workspace and notification services leverage Kafka to support Event-Driven Architecture (EDA), ensuring loose coupling and eventual consistency.

##### Core Architecture Patterns:
- **User Snapshot Synchronization (`user-registered` topic)**: When a user registers, the Auth Service publishes a `user-registered` event containing the user's basic profile details. The Workspace Service subscribes to this topic and replicates a local read-only copy of the user profiles (`SyncUserSnapshot` method) to execute fast workspace member joins and project listings without synchronous cross-service HTTP calls.
- **Session Termination Handling (`user-logout` topic)**: Logging out of a device triggers a `user-logout` event. The Workspace Service listens to this topic and instantly deletes the active login session cached under `session:<userId>:<deviceId>` in Redis, validating logout across all microservices.
- **Event-Driven Notifications (`send-notification` topic)**: Services publish notification events to trigger delivery. The Notification Service consumes these events and pushes real-time notifications to the client over WebSockets and records the notification history in MongoDB.
- **Asynchronous Task Creation (`task-created` topic)**: To support high-throughput, write-heavy task ingestion, tasks are not inserted directly into the database. Instead, the creation request generates a UUIDv7, queries the assignee name using a cache-first approach (checking the workspace members cache first, then the user repo database on miss), updates the Redis cache immediately for real-time reads, and publishes the task object to the `task-created` Kafka topic. A background consumer buffers incoming tasks and performs bulk inserts into PostgreSQL in batches of 1,000 tasks or every 2 seconds.

---

### 3. Notification Service (Node.js)

**Responsibility:** Real-time notifications, history
**Database:** MongoDB + Redis

#### API

| Method | Endpoint | Description | Auth | Sample Input |
| :--- | :--- | :--- | :--- | :--- |
| GET | `/api/v1/notifications` | Get notifications | Yes | N/A |
| PATCH | `/api/v1/notifications/:id/read` | Mark as read | Yes | `{"notificationIds":["6a245a7d7441e2849b9e9e6b"]}` |

#### Socket Events

- connection → `domain?token=Bearer (jwt token)`
- real-time push → event listen → `notification-received`

---

## Kafka Events

- user-registered → Auth → all service (Snapshot sync)
- user-logout → Auth → all service (for delete login session)
- send-notification → all service → notification
- task-created → Workspace → Workspace (async bulk task creation)

---

## Temporal Workflows

### Workspace Invite Workflow

- Trigger: Workflow invite with expire date
- Steps:
  - Send a invite email to the workspace with 14 days expire
  - After 10 days send a reminder email
  - After 14 daya maark it as a rexpired

---

## Cursor-Based Pagination Strategy

To support high scalability, low latency, and seamless infinite scrolling, the platform implements **Cursor-Based Pagination** using UUIDv7.

### Why Cursor-Based Pagination?
- **Consistent Retrieval Performance**: Traditional offset pagination (`LIMIT X OFFSET Y`) scales quadratically `O(N^2)` with page depth because PostgreSQL must scan and discard `Y` rows. Cursor-based pagination is `O(log N)` as it navigates directly using the index.
- **Drift Resilience**: If elements are inserted or deleted while a client is traversing lists, offset-based lists skip or duplicate rows. Cursor boundaries prevent list drifts.

### Technical Implementation:
1. **Chronological Cursors (UUIDv7)**: Workspace, projects and task IDs are stored as time-ordered UUIDv7s. Since the highest 48 bits encode Unix milliseconds, sorting alphabetically/lexicographically matches chronological creation time.
2. **Base64 Obfuscation**: Cursors are passed in the HTTP query string encoded in URL-safe base64 (e.g. `?cursor=MDE5Zjc1Y2YtNTh...`). This abstracts internal UUID representation and simplifies request parsing.
3. **Database Execution**:
   - Page 1 query: `SELECT * FROM workspaces WHERE owner_id = $1 ORDER BY id DESC LIMIT $2`
   - Next pages query: `SELECT * FROM workspaces WHERE owner_id = $1 AND id < $2 ORDER BY id DESC LIMIT $3`
4. **Pipelined Cache Pagination**: Redis lists are maintained in Sorted Sets (ZSETs) with a score of `0`. Cursors are paginated inside Redis using lexicographical range commands (`ZRevRangeByLex`) to query only the requested slide before looking up metadata hashes.

---

## Project Structure

```bash
taskflow-backend/
├── shared-proto/                  # Shared gRPC contracts
│   └── auth/
│       └── auth.proto
│
├── deployments/                   # Infrastructure & deployment configs
│   ├── docker-compose.yaml
│   └── k8s/
│       ├── global-ingress.yaml    # API Gateway
│       ├── otel-collector-values.yaml
│       ├── auth/
│       ├── workspace/
│       └── notification/
│
├── scripts/
│   └── gen-proto.sh               # Generate gRPC code
│   └── setup-and-build.sh         # Build full backend for local development 
│
├── services/
│   │
│   ├── auth-service/              # Authentication & session management
│   │   ├── src/
│   │   │   ├── config/
│   │   │   ├── middleware/
│   │   │   ├── modules/
│   │   │   │   ├── auth/
│   │   │   │   └── user/
│   │   │   ├── monitoring/
│   │   │   ├── utils/
│   │   │   ├── app.container.ts
│   │   │   ├── app.ts
│   │   │   └── server.ts
│   │   └── Dockerfile
│   │
│   ├── workspace-service/         # Workspace, Project & Task management
│   │   ├── cmd/
│   │   │   ├── api/
│   │   │   └── worker/
│   │   ├── config/
│   │   ├── internal/
│   │   │   ├── app/
│   │   │   ├── domain/
│   │   │   ├── middleware/
│   │   │   ├── workspace/
│   │   │   ├── project/
│   │   │   ├── task/
│   │   │   ├── user/
│   │   │   ├── kafka/
│   │   │   └── temporal/
│   │   ├── migrations/
│   │   └── pkg/
│   │
│   └── notification-service/      # Notifications & real-time delivery
│       ├── src/
│       │   ├── config/
│       │   ├── middleware/
│       │   ├── kafka/
│       │   ├── modules/
│       │   │   ├── notification/
│       │   │   └── user/
│       │   ├── monitoring/
│       │   ├── utils/
│       │   ├── app.container.ts
│       │   ├── app.ts
│       │   └── server.ts
│       │   └── worker.ts
│       └── Dockerfile
│
├── .gitignore
├── LICENSE
└── README.md

---

## Deployment Guide

This system supports two deployment profiles: Localhost Development and Production-Grade Enterprise Deployment.

### 1. Localhost Development Deployment

For local development, a pre-configured orchestrator script handles the full bootstrap, from spinning up docker databases to setting up a Kubernetes cluster, building service images, and installing monitoring charts.

#### Prerequisites
Ensure the following tools are installed on your machine and are added to your shell's PATH:
- **Docker & Docker Desktop** (Make sure the daemon is running)
- **KinD (Kubernetes in Docker)**
- **kubectl**
- **Helm**

#### Step-by-Step Run Instructions
1. **Execute the Setup Script**:
   Run the orchestrator script from the root of the workspace:
   ```bash
   ./scripts/setup-and-build.sh
   ```
   *The script will prompt you step-by-step to start database containers, build Docker images, configure KinD, install Helm charts (Prometheus, Grafana, Loki, Tempo, OpenTelemetry, Nginx Ingress), and deploy services.*

2. **Accessing Local Ingress & APIs**:
   Port-forward the Nginx Ingress Controller to route external HTTP requests to internal microservices:
   ```bash
   kubectl port-forward svc/ingress-nginx-controller -n ingress-nginx 8080:80
   ```
   Verify services are alive on port `8080`:
   - Auth Service: `GET http://localhost:8080/api/v1/auth/health`
   - Workspace Service: `GET http://localhost:8080/api/v1/workspace/health`
   - Notification Service: `GET http://localhost:8080/api/v1/notification/health`

3. **Accessing Dashboards**:
   Port-forward Grafana to inspect logs, traces, and metrics:
   ```bash
   kubectl port-forward svc/kube-stack-grafana 3000:80 -n monitoring
   ```
   - Open: `http://localhost:3000` (User: `admin`, get password using instructions printed by the setup script).

---

### 2. Production Deployment on AWS (Enterprise & Scale to Millions)

For highly resilient, enterprise-scale production running on AWS, we pivot away from local containers to fully managed cloud services:

#### Infrastructure Layer
- **Managed Kubernetes**: Run on **Amazon EKS** with autoscaling powered by **Karpenter** (or AWS Cluster Autoscaler) to dynamically provision EC2 worker nodes based on real-time resource demands.
- **API Gateway & Ingress**: Route public requests through an **AWS Network Load Balancer (NLB)** integrated with the Nginx Ingress controller. Manage SSL/TLS certificates securely via **AWS Certificate Manager (ACM)**.
- **Secrets Management**: Do not store passwords, connection strings, or JWT signing keys in standard plain-text Kubernetes ConfigMaps or Secrets. Integrate **AWS Secrets Manager** with the **External Secrets Operator (ESO)** to securely sync and inject secret variables into application pods.

#### Data & Streaming Layer
- **Relational Storage**: Migrate PostgreSQL to **Amazon Aurora PostgreSQL** (Serverless v2 or multi-AZ provisioned instances) for sub-millisecond read scaling, connection pooling (via PgBouncer or AWS RDS Proxy), and automated point-in-time recovery.
- **Document Storage**: Replace MongoDB with **Amazon DocumentDB (with MongoDB compatibility)** configured across multiple availability zones.
- **Cache & Session Management**: Upgrade Redis to **Amazon ElastiCache for Redis** in Cluster Mode to handle high-throughput session lookups and user caching.
- **Event Streaming**: Move the Apache Kafka cluster to **Amazon MSK (Managed Streaming for Apache Kafka)** to ensure high availability, partition auto-scaling, and secure VPC peering.
- **Durable Orchestration**: Migrate from self-hosted local Temporal instances to **Temporal Cloud** (fully managed Temporal namespace) to guarantee 99.99% availability, automated history/matching scale, and zero cluster maintenance, while only running workflow worker processes inside EKS.

#### Production Monitoring & Observability
- **Observability Stack (Prometheus, Grafana, Loki, Tempo)**: Route all telemetry (metrics, logs, traces) to **Grafana Cloud** rather than managing storage, compaction, and query layers for a self-hosted cluster on EKS. This provides a fully managed SaaS observability suite with auto-scaling ingestion, high query speeds, and zero-maintenance dashboard visualizations.
- **CI/CD Pipeline**: Automate deployments to EKS using GitOps controllers like **ArgoCD** or **FluxCD** integrated with GitHub Actions and AWS ECR (Elastic Container Registry).