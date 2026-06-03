# Distributed Workflow & Project Management System (mini jira)

## Project Overview

This is an enterprise-grade, collaborative project and workflow management platform designed for high scalability and resilience.

Unlike traditional task managers, this utilizes a distributed microservices architecture to handle high-concurrency environments and complex business logic.

The core objective of the project is to provide a seamless collaborative experience where teams can manage projects, track tasks, and automate repetitive workflows. By leveraging Event-Driven Architecture (EDA) and Durable Workflows, the system ensures that no task or automation is ever lost, even in the event of partial system failures.

---

## Key Features

- Workspace & Role Management (Multi-tenant + granular access control)
- Real-time Task Collaboration
- Automated Workflows (SLA & Reminders using durable execution)
- Scalable Notification Engine (Email, push, in-app)
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
- Prometheus & Grafana → Metrics
- Loki → Logging
- Jaeger → Tracing

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
| POST | `/api/v1/auth/logout` | Logout & blacklist token | Yes |  |

#### gRPC (Internal)
| Method | Request | Response | Description |
|--------|---------|----------|-------------|
| `VerifySession` | `token` | `VerifyResponse` | Validates JWT against Redis session & returns user metadata. |

#### Events Produced
- `user-registered`: Triggered when a new user signs up.

---

### 2. Workspace, Task & Project Service (Go)

**Responsibility:** Core business logic, workflow orchestration, and localized user profile snapshotting (via Kafka sync).
**Database:** PostgreSQL  

#### API

| Method | Endpoint | Description | Auth | Sample Input |
| :--- | :--- | :--- | :--- | :--- |
| POST | `/api/v1/workspace` | Create workspace | Yes | `{"name": "test", "slug": "test", "description": "test"}` |
| GET | `/api/v1/workspace/owned` | List workspaces (Owner) | Yes | |
| GET | `/api/v1/workspace/joined` | List workspaces (Member) | Yes | |
| POST | `/api/v1/workspace/:id/invites` | Invite member | Yes | `{"email": "test@test.com", "role": "ADMIN"}` |
| POST | `/api/v1/workspace/invites/accept` | Accept invite | Yes | `{"token": "6fa7dfcd-bfa2-4e13-bdb4-6e7fcb8ee8b5"}` |
| GET | `/api/v1/workspace/:id/members` | Get members | Yes | |
| POST | `/api/v1/workspace/:id/projects` | Create project | Yes | |
| GET | `/api/v1/workspace/:id/projects` | Get project | Yes | |
| POST | `/api/v1/workspace/projects/:id/tasks` | Create task | Yes | |
| GET | `/api/v1/workspace/projects/:id/tasks` | Get task | Yes | |
| PUT | `/api/v1/workspace/tasks/:id` | Update task | Yes | |
| PATCH | `/api/v1/workspace/tasks/:id/stataus` | Update task status | Yes | |
| POST | `/api/v1/workspace/tasks/:id/comments` | Add comment | Yes | |
| GET | `/api/v1/workspace/tasks/:id/comments` | Fetch comments | Yes | |

#### gRPC

- VerifyUser(UserID)

---

### 3. Notification Service (Node.js)

**Responsibility:** Real-time notifications, history, email dispatch  
**Database:** MongoDB  

#### API

| Method | Endpoint | Description | Auth |
|-------|---------|------------|------|
| GET | /api/v1/notifications | Get notifications | Yes |
| PATCH | /api/v1/notifications/:id/read | Mark as read | Yes |
| GET | /api/v1/notifications/unread-count | Unread count | Yes |

#### Socket Events

- connection → join `user_{id}`
- notification_received → real-time push
- unread_sync → sync pending notifications

---

## Kafka Events

- user-registered → Auth → Workspace (Snapshot sync)
- user.invited → Auth → Notification
- task.created → Task → Worker
- task.status.updated → Task → Notification
- notification.system.alert → Temporal → Notification

---

## Temporal Workflows

### Workspace Invite Workflow

- Trigger: Workflow invite with expire date
- Steps:
  - Send a invite email to the workspace with 14 days expire
  - After 10 days send a reminder email
  - After 14 daya maark it as a rexpired

### TaskManagementWorkflow

- Trigger: Task creation with deadline
- Steps:
  - Wait until 24 hours before deadline
  - Send reminder notification
  - Wait for completion signal
  - Trigger escalation if overdue

---

## Targeted Event Routing

- Redis stores user → pod mapping (with TTL)
- Kafka event consumed → lookup Redis → publish to specific pod channel
- Each pod listens only to its own channel

---

## Design Principles

- Stateless Auth → JWT
- Stateful Sessions → Redis
- Eventual Consistency → Kafka (User profile replication across services)
- Data Locality → User Snapshots in Workspace service (Reduces gRPC overhead)

---

## Resilience

- Distributed Rate Limiting (Redis)
- Graceful Shutdown (SIGTERM handling)
- Kubernetes Health Checks
- Dead Letter Queues (DLQ)
- Circuit Breakers (gRPC)
- ACID Transactions (PostgreSQL)

---

## Project Structure

```bash
taskflow-backend/
├── shared-proto/
├── services/
│   ├── auth-service/
│   ├── task-service/
│   └── notification-service/
├── deployments/
│   ├── docker-compose.yml
│   ├── k8s/
│   └── monitoring/
├── scripts/
└── README.md