# Election Backend

A RESTful API for managing elections, candidacies, and votes built with Go and Echo framework.

## 🏗️ Project Structure

```
ElectTNBackend/
├── cmd/server/              # Application entry point
├── internal/                # Private application code
│   ├── api/                # Generated API code and OpenAPI spec
│   ├── app/                # Server setup and routing
│   ├── auth/               # Session management
│   ├── config/             # Configuration management
│   ├── database/           # Database connection
│   ├── handlers/           # HTTP request handlers
│   ├── models/             # Data models
│   └── services/           # Business logic
├── scripts/                # Utility scripts
│   ├── db.sh              # Database management
│   └── integration-test.sh # Integration testing
├── deployments/            # Docker configurations
└── generated/              # Generated files
```

## 🚀 Features

- **UUID v4**: All entities use random UUIDs for security and scalability
- **Anonymous Voting**: Hash-based voting system with no vote logs
- **OAuth2 Authentication**: Google OAuth with session management
- **RESTful API**: Complete OpenAPI 3.0 specification
- **Production Ready**: Environment-specific configurations
- **Comprehensive Testing**: Unit and integration tests

## 🛠️ Getting Started

### Prerequisites

- Go 1.24.6+
- PostgreSQL 15+
- Docker (optional)

### Development

```bash
# Clone and build
git clone <repository>
cd elect-tn-backend
go mod tidy

# Set environment variables
export APP_MODE=development
export DATABASE_URL="postgres://user:password@localhost/elections_dev?sslmode=disable"

# Run tests
go test ./...

# Run server
go run ./cmd/server
```

### Using Docker

```bash
# Build and run with Docker Compose
docker-compose up --build

# Run tests in container
docker-compose -f docker-compose.test.yml up --build --abort-on-container-exit
```

## 📝 API Documentation

The API follows OpenAPI 3.0 specification. To view the documentation:

```bash
# Generate API code
./generate.sh

# View spec
cat api/openapi.yml
```

## 🔒 Security Features

- **Anonymous Voting**: Uses SHA-256 hashing for voter identification
- **No Vote Logs**: No detailed vote tracking, only counts
- **Session Management**: Secure session handling with gorilla/sessions
- **Role-Based Access**: Admin and user roles with proper authorization
- **UUID Security**: No sequential IDs to prevent enumeration

## 🧪 Testing

```bash
# Run all tests
go test ./...

# Run specific package tests
go test ./pkg/election
go test ./pkg/config

# Run with coverage
go test -cover ./...

# Run integration tests
go test -tags=integration ./test/...
```

## 🚦 Environment Variables

| Variable | Required | Development Default | Description |
|----------|----------|-------------------|-------------|
| `APP_MODE` | Yes | `development` | Application mode (`development`/`production`) |
| `SESSION_SECRET` | Production | `dev-secret-key-change-in-production` | Session encryption key |
| `DATABASE_URL` | Production | `postgres://...` | PostgreSQL connection string |
| `GOOGLE_CLIENT_ID` | Production | - | Google OAuth client ID |
| `GOOGLE_CLIENT_SECRET` | Production | - | Google OAuth client secret |
| `PORT` | No | `8080` | Server port |

## 📦 API Endpoints

### Authentication
- `GET /auth/google` - Start OAuth2 flow
- `POST /auth/callback` - OAuth2 callback
- `POST /auth/logout` - Logout user
- `GET /auth/me` - Get current user

### Elections
- `GET /elections` - List all elections
- `POST /elections` - Create new election (admin)
- `GET /elections/{id}` - Get election details
- `PUT /elections/{id}` - Update election (admin)
- `DELETE /elections/{id}` - Delete election (admin)
- `POST /elections/{id}/open` - Open/close election (admin)

### Candidacies
- `GET /elections/{id}/candidacies` - List candidates
- `POST /elections/{id}/candidacies` - Submit candidacy
- `DELETE /elections/{id}/candidacies/{id}` - Withdraw candidacy

### Voting
- `POST /elections/{id}/vote` - Cast vote
- `GET /elections/{id}/results` - Get results (admin)
- `GET /elections/{id}/live` - Get live results (if enabled)

## 🏛️ Architecture

The application follows a clean architecture pattern:

1. **cmd/**: Entry points for different applications
2. **pkg/**: Reusable library code organized by domain
3. **internal/**: Application-specific private code
4. **api/**: API specifications and generated code
5. **test/**: Integration and e2e tests

## 📄 License

[License name here]

## 🤝 Contributing

[Contributing guidelines here]