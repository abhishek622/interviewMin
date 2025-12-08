# interviewMin

A site where users paste interview experiences from sources (LeetCode discussions, Reddit, GfG, personal) and an AI converts them to structured records.

## commands

1. RUN migration

```bash
migrate -path internal/database/migrations -database "postgres://postgres:postgres@localhost:5432/interviewmin?sslmode=disable" up
```

2. CREATE migration

```bash
migrate create -ext sql -dir ./internal/database/migrations -seq create_questions_table
```

3. Docker compose start

```bash
docker-compose up -d
```
