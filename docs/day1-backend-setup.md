# Day 1 Backend Setup Checklist

## Completed scaffold
- cmd/server/main.go
- internal/config/config.go
- internal/bootstrap/db.go
- internal/router/router.go
- internal/handler/health_handler.go
- pkg/response/response.go
- configs/config.example.yaml
- configs/config.yaml

## What this scaffold provides
- Project entry and startup flow.
- Config loader (yaml + env overrides).
- MySQL bootstrap with connection pool and ping check.
- Gin router with `/ping` and `/api/v1/ping` health check endpoints.
- Unified JSON response format.

## Run commands
1. Ensure MySQL is running and schema created from `db/schema.sql`.
2. Update `configs/config.yaml` database dsn and jwt secret.
3. Run commands:

```bash
go mod tidy
go run ./cmd/server
```

## Expected result
- Service starts on configured host and port (default `0.0.0.0:8080`).
- Request `GET /ping` returns JSON with `code`, `message`, and `data`.

## Environment variable override examples
- `APP_SERVER_PORT=9090`
- `APP_DATABASE_DSN=root:pwd@tcp(127.0.0.1:3306)/task_system?charset=utf8mb4&parseTime=True&loc=Local`
- `APP_JWT_SECRET=your-secret`
