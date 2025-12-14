# Students API

A REST API server for managing student data, built with Go.

## Features

- **Configuration Management**: Using cleanenv for flexible config handling
- **SQLite Database**: Lightweight storage for student data
- **RESTful API**: Clean REST endpoints for student operations
- **Pagination**: Production-grade offset-based pagination for large datasets
- **Domain-Driven Errors**: Proper error handling with sentinel errors
- **Graceful Shutdown**: Safe server shutdown with timeout handling
- **Validation**: Request body validation using go-playground/validator
- **Clean Architecture**: Separation of concerns with handlers, storage, and types

## Getting Started

### Prerequisites

- Go 1.23.2 or higher

### Installation

1. Clone the repository
2. Install dependencies:
```bash
cd students_api
go mod download
```

### Configuration

The project uses YAML configuration files located in the `config/` directory:
- `config/local.yml` - Development environment
- `config/production.yml` - Production environment

You can specify which config to use via the `CONFIG_PATH` environment variable:
```bash
export CONFIG_PATH=config/production.yml
```

See [config/README.md](config/README.md) for detailed configuration documentation.

### Running the Application

```bash
# Run with default local config
go run cmd/go_students_api/main.go

# Run with custom config
CONFIG_PATH=config/production.yml go run cmd/go_students_api/main.go
```

## API Endpoints

### Create Student
```bash
POST /students
Content-Type: application/json

{
  "name": "John Doe",
  "email": "john@example.com",
  "age": 22
}
```

### Get Student by ID
```bash
GET /students/{id}
```

### Get Students List (Paginated)
```bash
# Default: page=1, limit=20
GET /students

# Custom pagination
GET /students?page=2&limit=50

# Response includes metadata
{
  "data": [...],
  "page": 2,
  "limit": 50,
  "total_items": 1500,
  "total_pages": 30,
  "has_next": true,
  "has_prev": true
}
```

See [docs/PAGINATION_GUIDE.md](docs/PAGINATION_GUIDE.md) for detailed pagination documentation.

## Project Structure

```
students_api/
├── cmd/
│   └── go_students_api/
│       └── main.go                     # Application entry point
├── config/
│   ├── local.yml                       # Local environment config
│   ├── production.yml                  # Production config
│   └── README.md                       # Config documentation
├── docs/
│   └── PAGINATION_GUIDE.md             # Pagination strategies guide
├── examples/                           # Example usage and patterns
├── internal/
│   ├── config/
│   │   └── config.go                   # Config loading logic
│   ├── http/
│   │   ├── handlers/
│   │   │   └── students/
│   │   │       └── students.go         # Student handlers (CRUD + List)
│   │   ├── helpers/
│   │   │   └── helper.go               # HTTP helper functions (pagination parsing)
│   │   └── response/
│   │       └── response.go             # JSON response utilities
│   ├── storage/
│   │   ├── sqlite/
│   │   │   └── sqlite.go               # SQLite implementation
│   │   └── storage.go                  # Storage interface & domain errors
│   └── types/
│       └── types.go                    # Domain types & structs
├── storage/
│   └── storage.db                      # SQLite database file
├── go.mod                              # Go module definition
├── go.sum                              # Dependency checksums
└── README.md                           # This file
```

## Architecture & Design Patterns

This project follows production-grade Go best practices:

### 1. **Clean Architecture**
- **Separation of Concerns**: Handlers, storage, and domain logic are separated
- **Dependency Inversion**: Handlers depend on storage interface, not concrete implementations
- **Domain-Driven Design**: Business logic in domain layer, infrastructure details in implementation

### 2. **Error Handling**
```go
// Domain errors defined in storage package
var (
    ErrNotFound    = errors.New("student not found")
    ErrDuplicate   = errors.New("student already exists")
    ErrDatabase    = errors.New("database error")
)

// Handlers check using errors.Is()
if errors.Is(err, storage.ErrNotFound) {
    // Handle not found case
}
```

### 3. **Pagination Strategy**
- Offset-based pagination with `LIMIT` and `OFFSET`
- Maximum limit enforcement (100 items per request)
- Rich metadata (total count, page info, navigation flags)
- Memory-safe for large datasets

### 4. **Validation**
- Request body validation using `go-playground/validator/v10`
- Type-safe validation with struct tags
- Clear error messages for clients

### 5. **Graceful Shutdown**
- Signal handling (SIGINT, SIGTERM)
- Graceful server shutdown with timeout
- Active requests completion before shutdown

## Dependencies

```go
require (
    github.com/ilyakaznacheev/cleanenv v1.5.0
    github.com/mattn/go-sqlite3 v1.14.24
    github.com/go-playground/validator/v10 v10.23.0
)
```

### Building
```bash
# Build for current platform
go build -o bin/students-api cmd/go_students_api/main.go

# Build for Linux
GOOS=linux GOARCH=amd64 go build -o bin/students-api-linux cmd/go_students_api/main.go

# Build for Windows
GOOS=windows GOARCH=amd64 go build -o bin/students-api.exe cmd/go_students_api/main.go
```

## Kubernetes Deployment (Kind Cluster)

### Prerequisites
- Docker installed
- Kind installed: `curl -Lo ./kind https://kind.sigs.k8s.io/dl/v0.20.0/kind-linux-amd64 && chmod +x ./kind && sudo mv ./kind /usr/local/bin/`
- kubectl installed: `curl -LO "https://dl.k8s.io/release/$(curl -L -s https://dl.k8s.io/release/stable.txt)/bin/linux/amd64/kubectl" && chmod +x kubectl && sudo mv kubectl /usr/local/bin/`

### Quick Deploy (5 Steps)

```bash
# 1. Create Kind cluster
kind create cluster --config k8s/kind-config.yaml

# 2. Build Docker image
docker build -t students-api:latest .

# 3. Load image into Kind
kind load docker-image students-api:latest --name students-api-cluster

# 4. Deploy to Kubernetes
kubectl apply -f k8s/namespace.yaml
kubectl apply -f k8s/configmap.yaml
kubectl apply -f k8s/pvc.yaml
kubectl apply -f k8s/deployment.yaml
kubectl apply -f k8s/service.yaml

# 5. Verify deployment
kubectl rollout status deployment/students-api -n students-api
```

**Access API:** http://localhost:30080

### Test Deployment

```bash
# Health check
curl http://localhost:30080/

# List students
curl http://localhost:30080/students

# Create a student
curl -X POST http://localhost:30080/students \
  -H "Content-Type: application/json" \
  -d '{"name":"John Doe","email":"john@example.com","age":22}'
```

### Useful Commands

```bash
# View logs
kubectl logs -f -l app=students-api -n students-api

# Check status
kubectl get all -n students-api

# Scale replicas
kubectl scale deployment students-api --replicas=3 -n students-api

# Update after code changes
docker build -t students-api:latest .
kind load docker-image students-api:latest --name students-api-cluster
kubectl rollout restart deployment/students-api -n students-api
```

### Cleanup

```bash
# Delete namespace (removes all app resources)
kubectl delete namespace students-api

# Delete Kind cluster
kind delete cluster --name students-api-cluster

# Delete Docker image
docker rmi students-api:latest
```

### Features
- ✅ 2 replicas for high availability
- ✅ Health checks (liveness, readiness, startup probes)
- ✅ Resource limits (CPU: 500m, Memory: 256Mi)
- ✅ Persistent storage for SQLite database
- ✅ Rolling updates (zero-downtime deployments)
- ✅ Graceful shutdown (40s termination grace period)