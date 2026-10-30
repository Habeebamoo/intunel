# Intunel

A high-throughput, fault-tolerant notification scheduling and dispatch system written in Go, powered by PostgreSQL for persistent state management and Redis Streams for asynchronous event processing.

Designed to guarantee at-least-once delivery, Intunel withstands worker crashes, network partitions, third-party rate limits ($429$), and thundering herd retry spikes using distributed concurrency, exponential backoff with full jitter, and a Dead Letter Queue (DLQ) replay engine.

### Processing Pipeline

```mermaid
flowchart TD
    Request["Receive HTTP Request"]
    Validate["Validate Payload"]
    Scheduled{"Scheduled?"}
    SaveDB["Save to PostgreSQL\nstatus: scheduled"]
    Publish["Publish to Redis Stream"]
    Scheduler["Scheduler wakes every 60s\nQueries due notifications"]
    Read["Worker Reads Stream"]
    Semaphore["Acquire Semaphore Slot"]
    Goroutine["Start Goroutine"]
    Send["Send Email via Resend"]
    Success{"Success?"}
    Ack["XACK Message"]
    Release["Release Semaphore Slot"]
    PEL["Stay in PEL"]
    Reaper["Reaper Claims after idle threshold"]
    RetryCount{"Retry count?"}
    Retry["Retry Send"]
    DLQ["Move to DLQ"]

    Request --> Validate
    Validate --> Scheduled
    Scheduled -- "yes" --> SaveDB
    Scheduled -- "no" --> Publish
    SaveDB --> Scheduler
    Scheduler --> Publish
    Publish --> Read
    Read --> Semaphore
    Semaphore --> Goroutine
    Goroutine --> Send
    Send --> Success
    Success -- "yes" --> Ack
    Ack --> Release
    Success -- "no" --> PEL
    PEL --> Reaper
    Reaper --> RetryCount
    RetryCount -- "< 3" --> Retry
    Retry --> Success
    RetryCount -- ">= 3" --> DLQ
```

## 🚀 Key Architectural Features

### 1. Concurrency-Bounded Worker Pool with Graceful Drain
- **Semaphore Reservation:** Consumers acquire a semaphore token *before* executing `XReadGroup` calls, preventing memory bloat from fetching unprocessable stream items.
- **Graceful Shutdown Protocol:** Traps `SIGINT`/`SIGTERM` OS signals, immediately stops stream ingestion, and utilizes a `sync.WaitGroup` with an independent `15s` timeout context to drain active in-flight HTTP requests without aborting ongoing provider calls.

### 2. At-Least-Once Delivery & PEL Recovery
- All consumer reads are unacknowledged (`XACK`) until downstream providers return an explicit success status.
- Unacknowledged or stalled messages remain in the Redis **Pending Entries List (PEL)**.

### 3. Anti-Thundering Herd Retry Reaper (Full Jitter)
- Stalled messages in the PEL are swept every $5\text{s}$ by a background `Reaper` process.
- Retries calculate dynamic delays using **Exponential Backoff + Full Jitter**:
  $$\text{RequiredIdle} = \text{random}\left(0, \, \min\left(\text{MaxBackoff}, \, \text{Base} \times 2^{\text{retry\_count}}\right)\right)$$
- Uniform randomizing prevents synchronized retry waves from overloading downstream APIs during recovery.

### 4. Fast-Fail DLQ & Replay Engine
- Messages failing exceeding $3$ consecutive attempts are automatically moved to `notifications:stream:dead`.
- Includes an administrative HTTP endpoint (`POST /api/v1/dlq/redrive`) to inspect and safely re-inject dead letters back into the main processing queue after upstream issues are resolved.

---

## ⚖️ Design Trade-offs

| Engineering Choice | Selected Option | Alternative Considered | Trade-off Justification |
| :--- | :--- | :--- | :--- |
| **Message Queue** | **Redis Streams** | Apache Kafka / RabbitMQ | Lower operational complexity for lightweight Go microservices while providing Consumer Groups, PEL, and explicit `XACK` semantics. |
| **Scheduling Engine** | **Single-Instance Poll** | Distributed Cron / `PG_NOTIFY` | Single-instance tick polling simplifies deployment and prevents cross-replica race conditions without requiring distributed locking (e.g., Redlock). |
| **Retry Backoff** | **Query-Time Full Jitter** | Fixed-Interval Retry | Query-time randomized threshold eliminates thundering herd spikes during outage recoveries without needing persistent state mutations. |
| **DLQ Operations** | **cURL Replay API** | Web GUI / Dashboard | Operational replay capability exposes programmatic control via CLI/HTTP tools without incurring frontend overhead for backend demo projects. |

---

## 🛠️ Tech Stack & Dependencies

* **Language:** Go $1.22+$
* **Database:** PostgreSQL $15+$
* **Cache / Broker:** Redis $7+$ (Streams & Consumer Groups)
* **Email Provider:** Resend SDK / Custom Router Interface

## 🚦 Getting Started & Local Development

### 1. Prerequisites
* Docker & Docker Compose
* PostgreSQL
* Redis
* Go $1.22+$

### 2. Environment Configuration
Create a `.env` file in the project root:
```env
PORT=8080
WORKER_PORT=8081
ENV=development
POSTGRES_URL=postgres://user:password@localhost:port/dbname
REDIS_URL=redis://localhost:port
RESEND_API_KEY=your_resend_key
SENDER_DOMAIN=username@yourdomain.com
DLQ_STREAM=notifications:stream:dead
```

#### 3. Build and Run
```bash
docker-compose up --build
```

#### 4. Without Docker

Start Redis, then run the API and worker separately:

```bash
go run ./cmd/api
go run ./cmd/worker
```

#### 5. Immediate Notification

```http
POST /api/v1/notify
Content-Type: application/json
Idempotency-Key: 550e8400-e29b-41d4-a716-446655440000
```

```json
{
  "channel": "email",
  "to": "john@example.com",
  "title": "Welcome",
  "body": "<h1>Welcome to our platform!</h1>"
}
```

#### 6. Scheduled Notification

```json
{
  "channel": "email",
  "to": "john@example.com",
  "title": "Trial ending soon",
  "body": "<h1>Your trial ends tomorrow</h1>",
  "date": "2026-07-21",
  "time": "09:00:00",
  "timezone": "Africa/Lagos"
}
```
