# Swagger Documentation Setup

This project follows the same Swagger setup pattern as the digital-wallet-demo project, with a dedicated standalone Swagger server.

## Architecture

The Swagger setup consists of:

1. **Server Interface** (`internal/api/interface.go`) - Common interface for all servers
2. **Swagger Server** (`internal/api/swagger_server.go`) - Standalone Swagger documentation server
3. **Logging Middleware** (`internal/api/log.go`) - Request logging for Swagger server
4. **Configuration** (`pkg/config/config.go`) - Swagger server port configuration

## Files Created

### 1. `internal/api/interface.go`
Defines the common `Server` interface used by all servers (API server, Swagger server, etc.):
```go
type Server interface {
    Name() string
    Run() error
    Shutdown(ctx context.Context) error
}
```

### 2. `internal/api/swagger_server.go`
Implements a standalone Swagger documentation server:
- Runs on a separate port from the main API server
- Uses Echo framework for routing
- Includes request logging middleware
- Graceful shutdown support

### 3. `internal/api/log.go`
Provides structured JSON logging for HTTP requests:
- Logs method, host, path, status, latency, etc.
- Uses logrus for structured logging
- Compatible with Echo middleware

### 4. `cmd/swagger.go`
Command to start the standalone Swagger server:
```bash
go run main.go swagger
```

### 5. `pkg/config/config.go`
Added `SwaggerConfig` with port configuration:
```go
type SwaggerConfig struct {
    Port string
}
```

## Configuration

### Environment Variables

Add to your `.env` file:
```env
SWAGGER_PORT=8081
```

### Default Values

- Swagger Server Port: `8081`
- Main API Server Port: `8080`

## Usage

### Starting the Swagger Server

**Option 1: Using go run**
```bash
go run main.go swagger
```

**Option 2: Using compiled binary**
```bash
./bin/ride_engine swagger
```

### Accessing Swagger UI

Once the server is running, access the Swagger UI at:
```
http://localhost:8081/swagger/index.html
```

### API Documentation JSON

The Swagger JSON specification is available at:
```
http://localhost:8081/swagger/doc.json
```

## Running Both Servers

You can run both the API server and Swagger server simultaneously:

**Terminal 1 - API Server:**
```bash
go run main.go serve
```

**Terminal 2 - Swagger Server:**
```bash
go run main.go swagger
```

## Architecture Pattern

This setup follows the digital-wallet-demo pattern:

1. **Separation of Concerns**: Swagger documentation runs independently from the API
2. **Server Interface**: Common interface allows multiple server types
3. **Structured Logging**: Consistent request logging across all servers
4. **Configuration**: Environment-based configuration for flexibility
5. **Graceful Shutdown**: All servers support clean shutdown

## Benefits

1. **Independent Scaling**: Swagger server can be scaled separately
2. **Security**: Swagger docs can be hosted on a different network/port
3. **Performance**: API server isn't affected by documentation traffic
4. **Clean Architecture**: Follows interface-based design
5. **Consistent Logging**: Structured logs for all HTTP requests

## Docker Deployment

To run the Swagger server in Docker, you can update `docker-compose.yml`:

```yaml
  swagger:
    build: .
    container_name: ride_engine-swagger
    command: ["./bin/ride_engine", "swagger"]
    ports:
      - "8081:8081"
    environment:
      - SWAGGER_PORT=8081
    networks:
      - ride_engine_network
```

## Development

### Regenerating Swagger Documentation

When you update API handlers or add new endpoints:

```bash
swag init -g main.go -o docs
```

### Testing

Test the Swagger server:
```bash
# Start server
go run main.go swagger

# Test UI
curl http://localhost:8081/swagger/index.html

# Test JSON
curl http://localhost:8081/swagger/doc.json | jq .info.title
```

## Troubleshooting

### Port Already in Use

If port 8081 is already in use, change the port:
```bash
export SWAGGER_PORT=8082
go run main.go swagger
```

### Swagger UI Not Loading

1. Check if the server is running: `curl http://localhost:8081/swagger/`
2. Verify swagger docs exist: `ls docs/swagger.json`
3. Regenerate docs if needed: `swag init -g main.go -o docs`

### Missing Dependencies

If you get import errors:
```bash
go mod tidy
go mod vendor
```

## Comparison with digital-wallet-demo

This implementation mirrors the digital-wallet-demo structure:

| Feature | digital-wallet-demo | ride_engine |
|---------|---------------------|-------------|
| Server Interface | ✅ | ✅ |
| Standalone Swagger | ✅ | ✅ |
| Request Logging | ✅ | ✅ |
| Configuration | ✅ | ✅ |
| Echo Framework | ✅ | ✅ |
| Logrus Logging | ✅ | ✅ |

## Next Steps

1. Consider adding authentication to the Swagger server
2. Add rate limiting to prevent abuse
3. Implement CORS if needed for browser access
4. Add metrics/monitoring endpoints
5. Consider adding Swagger UI themes/customization

