# High-Frequency Trade Settlement Engine

A high-performance, concurrent backend system simulating a trade settlement engine. This project demonstrates advanced patterns in Distributed Systems, including ACID transactions, worker pools for concurrency, and infrastructure-as-code.

## 🏗 Architecture

The system mimics a real-world fintech settlement pipeline:

1.  **Ingestion API (Producer):** A Go HTTP server accepts trade orders and persists them to Postgres (`PENDING` state) with zero latency.
2.  **Settlement Engine (Consumer):** A background worker pool processes trades concurrently, simulating market latency and execution logic.
3.  **Real-Time Dashboard:** Server-Sent Events (SSE) push updates to the UI instantly upon settlement.

### Tech Stack

* **Language:** Go 1.23+ (Chi Router, Pgx Driver)
* **Database:** PostgreSQL 16 (running in Docker)
* **Infrastructure:** Docker Compose, Taskfile
* **Frontend:** Templ + Datastar (planned)

## 🚀 Quick Start

This project uses `Task` (go-task) for workflow automation.

### Prerequisites

* [Go 1.23+](https://go.dev/dl/)
* [Docker Desktop](https://www.docker.com/products/docker-desktop/)
* [Task](https://taskfile.dev/installation/) (or use `make` if preferred)

### 1. Setup Environment
Copy the example environment file:
```bash
cp .env.example .env
```
### 2. Start Infrastructure
Spin up the Postgres database in a container. This will automatically apply the schema migrations found in /migrations.
```bash
task up
```

### 3. Run Application
Start the Go server locally. It connects to the Dockerized database automatically.
```bash
task run
```

You should see:

Connected to database! Starting server on :8080...

### 4. Developer Commands
We use a Taskfile.yml to standardize commands across the team.

| Command     | Description                                           |
|-------------|-------------------------------------------------------|
| `task up`   | Start the database container (detached)              |
| `task down` | Stop and remove the database container                |
| `task run`  | Run the Go application (hot-reloads .env)            |
| `task db-shell` | Open a PSQL CLI session inside the container     |
| `task logs` | Tail the database logs                                |
| `task build`| Compile the binary to /bin                            |


### 5. Testing  

(Coming soon...)
