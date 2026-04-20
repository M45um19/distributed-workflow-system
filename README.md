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

**Responsibility:** Authentication, session management, workspace membership  
**Database:** MongoDB + Redis  

#### API

| Method | Endpoint | Description | Auth |
|-------|---------|------------|------|
| POST | /api/v1/auth/register | Register user | No |
| POST | /api/v1/auth/login | Login (JWT + Refresh Token) | No |
| POST | /api/v1/auth/logout | Logout & blacklist token | Yes |
| POST | /api/v1/workspaces | Create workspace | Yes |
| GET | /api/v1/workspaces | List workspaces | Yes |
| POST | /api/v1/workspaces/:id/invite | Invite member | Yes |
| GET | /api/v1/workspaces/:id/members | Get members | Yes |

---

### 2. Task & Project Service (Go)

**Responsibility:** Core business logic & workflow orchestration  
**Database:** PostgreSQL  

#### API

| Method | Endpoint | Description | Auth |
|-------|---------|------------|------|
| POST | /api/v1/projects | Create project | Yes |
| GET | /api/v1/projects/:id | Get project (cached) | Yes |
| POST | /api/v1/tasks | Create task (Temporal workflow) | Yes |
| GET | /api/v1/tasks/:id | Get task | Yes |
| PATCH | /api/v1/tasks/:id | Update task | Yes |
| POST | /api/v1/tasks/:id/comments | Add comment | Yes |

#### gRPC

- VerifyUser(UserID)
- GetWorkspaceContext(WSID)

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

- user.invited → Auth → Notification
- task.created → Task → Worker
- task.status.updated → Task → Notification
- notification.system.alert → Temporal → Notification

---

## Temporal Workflows

### TaskManagementWorkflow

- Trigger: Task creation with deadline
- Steps:
  - Wait until 24 hours before deadline
  - Send reminder notification
  - Wait for completion signal
  - Trigger escalation if overdue

### RecurringTaskWorkflow

- Trigger: Cron schedule
- Action:
  - Auto-create recurring tasks via gRPC

---

## Targeted Event Routing

- Redis stores user → pod mapping (with TTL)
- Kafka event consumed → lookup Redis → publish to specific pod channel
- Each pod listens only to its own channel

---

## Design Principles

- Stateless Auth → JWT
- Stateful Sessions → Redis
- Eventual Consistency → Kafka
- Durable Execution → Temporal

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