# Students API

A REST API server for managing student data, built with Go.

## Features

- **Configuration Management**: Using cleanenv for flexible config handling
- **SQLite Database**: Lightweight storage for student data
- **RESTful API**: Clean REST endpoints for student operations

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

## Project Structure

```
students_api/
├── cmd/
│   └── go_students_api/
│       └── main.go           # Application entry point
├── config/
│   ├── local.yml             # Local environment config
│   ├── production.yml        # Production config
│   └── README.md             # Config documentation
├── internal/
│   └── config/
│       └── config.go         # Config loading logic
├── storage/
│   └── storage.db            # SQLite database
└── go.mod
```