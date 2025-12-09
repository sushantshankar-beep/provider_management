# Provider Management - AI Coding Instructions

## Architecture Overview

This is a **microservices-based Go backend** using DDD (Domain-Driven Design) for complaint and payout management. Two main entrypoints:
- **API Server** (`cmd/api/main.go`): HTTP API for complaint operations via Gin
- **Worker** (`cmd/worker/main.go`): Async message consumer from RabbitMQ

**Data Flow**: HTTP requests → API (MongoDB) → RabbitMQ → Worker processes events asynchronously.

### Key Components

| Component | Purpose | Tech |
|-----------|---------|------|
| **Complaint Domain** (`internal/domain/complaint/`) | Core business logic; entities with statuses (pending/in_review/resolved) | Plain Go structs |
| **Repositories** (`internal/ports/repository/`) | Interfaces for MongoDB access (booking, complaint) | Port pattern for DI |
| **UseCases** (`internal/usecase/`) | Orchestrate domain + repositories (e.g., create complaint from booking) | Interactors |
| **RabbitMQ Queue** (`internal/infrastructure/queue/`) | Async event publishing/consuming; `GoroutinePool` for worker concurrency | AMQP |
| **HTTP API** (`internal/infrastructure/http/`) | REST endpoints and Gin routing; controllers inject DB instance | Gin framework |

## Critical Patterns

### 1. Dependency Injection via Constructor
Controllers and UseCases receive dependencies in constructors, NOT singletons. Example:
```go
// ✓ Correct: inject at controller creation
func NewComplaintController(db *mongo.Database, cfg *config.Config) *ComplaintController {
    pool := shared.NewGoroutinePool(cfg.AsyncWorkerPoolSize)
    return &ComplaintController{db: db, cfg: cfg, pool: pool}
}

// In router: pass db instance
ch := controllers.NewComplaintController(db)
```

### 2. Repository Interface Pattern
All data access goes through interfaces defined in `ports/repository/`. MongoDB implementations in `infrastructure/db/`. 
- Add query methods to the interface first, then implement in `*_repo_mongo.go`
- Use `bson` tags for MongoDB field mapping

```go
// ports/repository/complaint.go
type ComplaintRepository interface {
    Create(ctx context.Context, c *complaint.Complaint) error
    FindByBookingID(ctx context.Context, bookingID string) (*complaint.Complaint, error)
}
```

### 3. Context & Timeouts
Every function accepts `context.Context`. Controllers set 10-second timeouts:
```go
ctx, cancel := context.WithTimeout(c.Request.Context(), 10*time.Second)
defer cancel()
```

### 4. Queue Message Structure
All RabbitMQ messages use `QueueMessage` struct (`queue/message.go`). When publishing:
```go
notifier.Publish(&queue.QueueMessage{
    Type: "complaint_created",
    ComplaintID: complaintID,
    Payload: rawData,
})
```

### 5. GoroutinePool for Async Work
Instead of `go func()`, submit to pool to avoid unbounded goroutine growth:
```go
pool.Submit(func() {
    // async work here
})
```
Pool gracefully recovers from panics and falls back to new goroutines if queue is full.

## Development Workflows

### Build & Run

**API Server**:
```powershell
cd cmd/api; go run main.go
# Requires: MONGO_URI, RABBITMQ_URL env vars set
```

**Worker**:
```powershell
cd cmd/worker; go run main.go
# Same env vars; listens on RABBITMQ_QUEUE (default: "complaints.booking")
```

### Configuration
All config via environment variables in `internal/config/config.go`:
- `MONGO_URI` (default: "mongodb://localhost:27017")
- `MONGO_DB` (default: "provider_db")
- `RABBITMQ_URL` (default: "amqp://guest:guest@localhost:5672/")
- `RABBITMQ_QUEUE` (default: "complaints.booking")
- `API_PORT` (default: "8080")
- `CONSUMER_WORKER_COUNT` (default: 5)
- `ASYNC_WORKER_POOL_SIZE` (default: 10)

### Adding a New Feature

1. **Define domain entity** in `internal/domain/<domain>/entity.go`
2. **Add repository interface** in `internal/ports/repository/<domain>.go`
3. **Implement MongoDB repo** in `internal/infrastructure/db/<domain>_repo_mongo.go`
4. **Create usecase** in `internal/usecase/<domain>/`
5. **Add HTTP endpoint** in router and create controller in `internal/infrastructure/http/controllers/`
6. **If async work needed**: publish `QueueMessage` to RabbitMQ; handle in consumer

## Error Handling

- Repository methods return `nil, nil` for "not found" (check with `if obj == nil`)
- MongoDB `ErrNoDocuments` handled explicitly:
  ```go
  if errors.Is(err, mongo.ErrNoDocuments) {
      return nil, nil
  }
  ```
- Controllers respond with HTTP status codes; logs use `log.Println("[CONTEXT]", msg)`

## Testing Considerations

- Inject repositories via interface so tests can mock
- Use context with timeouts to prevent test hangs
- Repositories use MongoDB's context, so integration tests need running MongoDB

## External Dependencies

- **MongoDB 3.14.0+**: Complaint/booking document storage
- **RabbitMQ (amqp091-go v1.4.0)**: Async event stream
- **Gin v1.11.0**: HTTP framework
- **Go 1.23.0**: Required version
