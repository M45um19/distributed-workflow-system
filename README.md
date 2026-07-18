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
| GET | `/api/v1/workspace/:id/projects` | Get project | Yes | N/A |
| POST | `/api/v1/workspace/:id/projects/:projectId/tasks` | Create task | Yes | `{"title":"Implement Kafka Event Consumer","description":"Create a robust worker to consume user-registration events from the message queue.","priority":"HIGH","assignee_id":"6a107afad2ac1e59aba88b6f","deadline":"2026-06-15T18:30:00Z"}` |
| GET | `/api/v1/workspace/:id/projects/:projectId/tasks` | Get task | Yes | N/A |
| PUT | `/api/v1/workspace/:id/tasks/:taskId` | Update task | Yes | `{"title":"Implement Kafka Event Consumer 2","description":"Create a robust worker to consume user-registration events from the message queue. 2","priority":"HIGH","assignee_id":"6a107afad2ac1e59aba88b6f","deadline":"2026-06-15T18:30:00Z"}` |
| PATCH | `/api/v1/workspace/:id/tasks/:taskId/status` | Update task status | Yes | `{"status":"DONE"}` |
| POST | `/api/v1/workspace/:id/tasks/:taskId/comments` | Add comment | Yes | `{"content":"DONE"}` |
| GET | `/api/v1/workspace/:id/tasks/:taskId/comments` | Fetch comments | Yes | N/A |


#### Events Produced
- `send-notification`: Triggered when need to send a notification.

#### Redis Caching Architecture

##### Why use Redis in workspace:
- **High-Frequency Read Optimization**: Workspace owned/joined listings and membership checks are heavily read-intensive. Redis shields the Postgres database from high query volumes.
- **Lexicographical Pagination**: Allows paginating workspaces chronologically descending directly inside Redis without transferring the entire set of IDs to the Go application memory.
- **Fast Authorization Checks**: Workspace roles are cached to evaluate permissions instantly during handler auth.

##### How use Redis in workspace:
- **Workspace Metadata Cache (`workspace:<workspaceId>:meta`)**: Redis Hash storing core workspace fields with a 24-hour TTL.
- **Lexicographical ZSET Indexing (`user:<userId>:workspaces:owned` and `user:<userId>:workspaces:joined`)**: Stored as Sorted Sets where all members have a score of `0`. Redis sorts them lexicographically. Since IDs are UUIDv7, lexicographical sorting corresponds to chronological sorting.
  - Pagination fetches exactly `limit` IDs using `ZRevRangeByLex` with exclusive boundary offsets (`Max: "(" + cursor`).
- **Workspace Role Cache (`workspace:<workspaceId>:roles`)**: Hash mapping `user_id -> role` for fast permission lookup.
- **Workspace Members Cache (`workspace:<workspaceId>:members`)**: Hash storing JSON strings of `WorkspaceMemberResponse` indexed by `user_id` for quick collection retrieval and single member updates.
- **Consistency & Invalidation**: We follow the Cache-Aside pattern. On creating workspaces or accepting invites, the corresponding ZSET caches are dynamically appended (`ZAdd`) and hashes updated/invalidated to guarantee strong read-after-write consistency.

#### Kafka Architecture
*(Section detail to be written later)*

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