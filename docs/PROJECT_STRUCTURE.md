# Go Project Structure

This project follows golang-standards/project-layout conventions, similar to Kubernetes and other production Go applications.

## Directory Structure

```
elect-tn-backend/
├── cmd/                    # Application entry points
│   └── server/
│       └── main.go         # Main application entry
├── internal/               # Private application code
│   ├── app/               # Application-specific code
│   │   └── server.go      # Server setup and routing
│   └── handlers/          # HTTP request handlers
│       └── election_handler.go
├── pkg/                    # Public library code
│   ├── api/               # Generated API types and interfaces
│   ├── config/            # Configuration management
│   ├── models/            # Data models and database schemas
│   └── services/          # Business logic services
├── deployments/           # Deployment configurations
│   └── dockers/          # Docker files
├── api/                   # OpenAPI specifications
│   └── openapi.yml
└── docs/                  # Documentation
```

## Principles

### cmd/
- Contains only main packages and minimal initialization code
- Each application has its own directory (server, cli, etc.)

### internal/
- **app/**: Application-specific code, server setup, routing
- **handlers/**: HTTP request handlers specific to this application
- Code in internal/ cannot be imported by external projects

### pkg/
- **api/**: Auto-generated OpenAPI types and server interfaces
- **config/**: Reusable configuration management
- **models/**: Data models that could be shared across services
- **services/**: Business logic that could be reused

### deployments/
- All deployment-related configurations
- Dockerfiles, Kubernetes manifests, etc.

This structure provides:
- ✅ Clear separation between public and private code
- ✅ Reusable components in pkg/
- ✅ Application-specific logic in internal/
- ✅ Standard layout familiar to Go developers
- ✅ Easy testing and dependency management
