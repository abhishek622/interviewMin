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
