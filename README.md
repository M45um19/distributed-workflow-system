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

### 1. Auth & Workspace Service (Node.js)

**Responsibility:** Authentication, session management
**Database:** MongoDB + Redis  

#### API

| Method | Endpoint | Description | Auth | Sample Input |
| :--- | :--- | :--- | :--- | :--- |
| POST | `/api/v1/auth/register` | Register user | No | `{"full_name": "test", "email": "test@test.com", "password": "test1234"}` |
| POST | `/api/v1/auth/login` | Login (JWT + Refresh Token) | No | `{"email": "test@test.com", "password": "test1234"}` |
| POST | `/api/v1/auth/logout` | Logout with device id | Yes |  |

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
| GET | `/api/v1/workspace/owned` | List workspaces (Owner) | Yes | N/A |
| GET | `/api/v1/workspace/joined` | List workspaces (Member) | Yes | N/A |
| POST | `/api/v1/workspace/:id/invites` | Invite member | Yes | `{"email":"test@test.com","role":"ADMIN"}` |
| POST | `/api/v1/workspace/invites/accept` | Accept invite | Yes | `{"token":"6fa7dfcd-bfa2-4e13-bdb4-6e7fcb8ee8b5"}` |
| GET | `/api/v1/workspace/:id/members` | Get members | Yes | N/A |
| POST | `/api/v1/workspace/:id/projects` | Create project | Yes | `{"name":"E-Commerce Microservices Backend","description":"This project handles the core ordering and payment workflow systems."}` |
| GET | `/api/v1/workspace/:id/projects` | Get project | Yes | N/A |
| POST | `/api/v1/workspace/projects/:id/tasks` | Create task | Yes | `{"title":"Implement Kafka Event Consumer","description":"Create a robust worker to consume user-registration events from the message queue.","priority":"HIGH","assignee_id":"6a107afad2ac1e59aba88b6f","deadline":"2026-06-15T18:30:00Z"}` |
| GET | `/api/v1/workspace/projects/:id/tasks` | Get task | Yes | N/A |
| PUT | `/api/v1/workspace/tasks/:id` | Update task | Yes | `{"title":"Implement Kafka Event Consumer 2","description":"Create a robust worker to consume user-registration events from the message queue. 2","priority":"HIGH","assignee_id":"6a107afad2ac1e59aba88b6f","deadline":"2026-06-15T18:30:00Z"}` |
| PATCH | `/api/v1/workspace/tasks/:id/status` | Update task status | Yes | `{"status":"DONE"}` |
| POST | `/api/v1/workspace/tasks/:id/comments` | Add comment | Yes | `{"content":"DONE"}` |
| GET | `/api/v1/workspace/tasks/:id/comments` | Fetch comments | Yes | N/A |


#### Events Produced
- `send-notification`: Triggered when need to send a notification.

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