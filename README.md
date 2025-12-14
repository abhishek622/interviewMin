# InterviewMin Backend

InterviewMin is a platform that simplifies the process of tracking interview experiences. Users can paste interview experiences from various sources (LeetCode, Reddit, GeeksForGeeks, or personal notes), and the system uses AI (Groq) to extract and organize them into structured records.

![Go Version](https://img.shields.io/badge/go-1.21+-00ADD8?style=flat&logo=go)
![License](https://img.shields.io/badge/license-MIT-blue.svg)

## 🚀 Features

- **Multi-Source Support**: Import experiences from LeetCode, Reddit, GfG, and custom sources.
- **AI-Powered Extraction**: Automatically extracts company, position, rounds, location, and questions using Large Language Models (LLM).
- **Structured Records**: distinct separation of interview metadata and specific questions.
- **Interview Management**: Create, read, update, and delete (CRUD) interview experiences.
- **Company Tracking**: Automatic organization of interviews by company hierarchy.
- **Analytics**: Statistics on interview trends, top companies, and user activity.
- **Secure Authentication**: JWT-based authentication with refresh token rotation and session management.
- **Role-Based Access**: Standard user and admin roles.

## 🛠 Tech Stack

- **Language**: Go (Golang)
- **Framework**: Gin Web Framework
- **Database**: PostgreSQL
- **AI Integration**: Groq API
- **Authentication**: JWT (JSON Web Tokens)
- **Logging**: Zap (Structured logging)
- **Containerization**: Docker & Docker Compose

## 📂 Project Structure

```bash
backend/
├── cmd/api/            # Application entrypoint
├── internal/
│   ├── auth/           # JWT & auth logic
│   ├── config/         # Configuration loading
│   ├── database/       # DB connection & migrations
│   ├── fetcher/        # External content fetchers
│   ├── groq/           # AI Client integration
│   ├── handler/        # HTTP Request handlers
│   ├── logger/         # Zap logger setup
│   └── repository/     # Data access layer
├── pkg/
│   ├── model/          # Data models
│   ├── response/       # Standardized API response helpers
│   └── *.go          # Utility functions
└── README.md
```

## 📖 Useful Commands

### Database Migrations

**Run Migrations:**

```bash
migrate -path internal/database/migrations -database "postgres://postgres:postgres@localhost:5432/interviewmin?sslmode=disable" up
```

**Create New Migration:**

```bash
migrate create -ext sql -dir ./internal/database/migrations -seq create_questions_table
```

> **Note**: You need to have `golang-migrate` installed on your machine.

### Docker

**Start Services:**

```bash
docker-compose up -d
```

**Stop Services:**

```bash
docker-compose down
```

## 🛡️ Security Features

- **Rate Limiting**: Configurable token bucket rate limiting per IP.
- **Graceful Shutdown**: Proper handling of in-flight requests during shutdown.
- **Secure Headers**: CORS policy enforcement.
- **Input Validation**: Strict JSON binding and validation.

## 🚀 Deployment (Leapcell.io)

This application is configured for deployment on [Leapcell.io](https://leapcell.io).

### Build Configuration

| Setting           | Value                                              |
| ----------------- | -------------------------------------------------- |
| **Runtime**       | Go (Any version)                                   |
| **Build Command** | `make build` or use the command in `leapcell.yaml` |
| **Start Command** | `./app`                                            |
| **Port**          | `8080`                                             |
| **Health Check**  | `/v1/healthcheck`                                  |

### Required Environment Variables

| Variable               | Description                              |
| ---------------------- | ---------------------------------------- |
| `ENV`                  | Set to `production`                      |
| `PORT`                 | `8080` (default)                         |
| `DATABASE_URL`         | PostgreSQL connection string             |
| `JWT_SECRET`           | JWT secret key (min 32 characters)       |
| `AES_SECRET_KEY`       | AES encryption key (16, 24, or 32 bytes) |
| `GROQ_API_KEY`         | Groq API key for AI features             |
| `CORS_TRUSTED_ORIGINS` | Comma-separated list of allowed origins  |

### Optional Environment Variables

| Variable                | Default                                         | Description                       |
| ----------------------- | ----------------------------------------------- | --------------------------------- |
| `DB_MAX_OPEN_CONNS`     | `25`                                            | Max open database connections     |
| `DB_MAX_IDLE_CONNS`     | `25`                                            | Max idle database connections     |
| `DB_MAX_IDLE_TIME`      | `15m`                                           | Max idle time for connections     |
| `RATE_LIMIT_RPS`        | `10`                                            | Requests per second limit         |
| `RATE_LIMIT_BURST`      | `20`                                            | Burst limit for rate limiter      |
| `RATE_LIMIT_ENABLED`    | `true`                                          | Enable/disable rate limiting      |
| `JWT_ACCESS_TOKEN_TTL`  | `15m`                                           | Access token expiration           |
| `JWT_REFRESH_TOKEN_TTL` | `168h`                                          | Refresh token expiration (7 days) |
| `GROQ_MODEL`            | `meta-llama/llama-4-maverick-17b-128e-instruct` | Groq model                        |
| `GROQ_TIMEOUT`          | `30s`                                           | Groq API timeout                  |

### Makefile Commands

```bash
make build        # Build production binary
make build-local  # Build for local development
make run          # Run locally (with air hot-reload if available)
make test         # Run all tests
make lint         # Run linter
make version      # Display version info
make help         # Show all available commands
```
