# Production API Example

This example demonstrates a production-ready API built with the Zoox framework, showcasing advanced features, security best practices, monitoring, and proper application architecture.

## 🚀 Features

### 🔐 Security
- **Authentication & Authorization**: Token-based authentication with role-based access control
- **Security Headers**: Helmet middleware for security headers
- **CORS Protection**: Cross-origin resource sharing configuration
- **Rate Limiting**: Request rate limiting to prevent abuse
- **Input Validation**: Request body validation and sanitization

### 📊 Monitoring & Observability
- **Health Checks**: Comprehensive health check endpoint
- **Metrics Collection**: Application metrics and performance monitoring
- **Request Logging**: Detailed request/response logging
- **Error Tracking**: Error counting and monitoring
- **Debug Information**: Development-only debug endpoints

### 🏗️ Architecture
- **Clean Architecture**: Separation of concerns with structured application design
- **Configuration Management**: Environment-based configuration
- **Graceful Shutdown**: Proper server shutdown handling
- **Middleware Pipeline**: Comprehensive middleware stack
- **API Versioning**: Versioned API endpoints

### 🛡️ Production Features
- **Timeouts**: Request timeout handling
- **Compression**: Gzip compression for responses
- **Real IP Detection**: Proper client IP detection behind proxies
- **Recovery**: Panic recovery middleware
- **Request ID**: Unique request ID for tracing

## 🏃 Quick Start

### 1. Run the Application

```bash
cd examples/06-production-api
go run main.go
```

### 2. Test the Endpoints

The server will start on `http://localhost:8080` with the following endpoints:

#### Public Endpoints
- `GET /` - API information
- `GET /health` - Health check
- `GET /metrics` - Application metrics
- `GET /api/v1/status` - API status
- `POST /api/v1/auth/login` - User authentication

#### Protected Endpoints (require authentication)
- `GET /api/v1/protected/users` - List users
- `POST /api/v1/protected/users` - Create user
- `GET /api/v1/protected/users/:id` - Get user
- `PUT /api/v1/protected/users/:id` - Update user
- `DELETE /api/v1/protected/users/:id` - Delete user

#### Admin Endpoints (require admin role)
- `GET /api/v1/protected/admin/stats` - Admin statistics
- `GET /api/v1/protected/admin/logs` - Application logs

#### Development Endpoints (development mode only)
- `GET /debug/pprof/` - Go profiling
- `GET /debug/vars` - Debug variables

## 🔧 Configuration

The application can be configured using environment variables:

```bash
export PORT=8080
export ENV=production
export DB_HOST=localhost
export DB_PORT=5432
export JWT_SECRET=your-secret-key
```

## 📝 API Usage Examples

### Authentication

```bash
# Login to get access token
curl -X POST http://localhost:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"username": "admin", "password": "secret"}'

# Response
{
  "success": true,
  "data": {
    "token": "valid-token",
    "expires_in": 3600,
    "user": {
      "username": "admin",
      "role": "admin"
    }
  }
}
```

### Health Check

```bash
curl http://localhost:8080/health

# Response
{
  "status": "healthy",
  "version": "1.0.0",
  "timestamp": "2024-01-01T12:00:00Z",
  "services": {
    "database": "connected",
    "cache": "connected",
    "queue": "connected"
  }
}
```

### Metrics

```bash
curl http://localhost:8080/metrics

# Response
{
  "success": true,
  "data": {
    "request_count": 1250,
    "error_count": 12,
    "average_latency_ms": 45.2,
    "uptime": "2h30m15s"
  }
}
```

### User Management (Protected)

```bash
# List users (requires authentication)
curl http://localhost:8080/api/v1/protected/users \
  -H "Authorization: Bearer valid-token"

# Create user
curl -X POST http://localhost:8080/api/v1/protected/users \
  -H "Authorization: Bearer valid-token" \
  -H "Content-Type: application/json" \
  -d '{"username": "newuser", "email": "new@example.com", "role": "user"}'

# Get specific user
curl http://localhost:8080/api/v1/protected/users/1 \
  -H "Authorization: Bearer valid-token"

# Update user
curl -X PUT http://localhost:8080/api/v1/protected/users/1 \
  -H "Authorization: Bearer valid-token" \
  -H "Content-Type: application/json" \
  -d '{"username": "updateduser", "email": "updated@example.com"}'

# Delete user
curl -X DELETE http://localhost:8080/api/v1/protected/users/1 \
  -H "Authorization: Bearer valid-token"
```

### Admin Statistics (Admin Only)

```bash
curl http://localhost:8080/api/v1/protected/admin/stats \
  -H "Authorization: Bearer valid-token"

# Response
{
  "success": true,
  "data": {
    "total_users": 3,
    "active_users": 2,
    "total_requests": 1250,
    "error_rate": 0.96,
    "uptime": "2h30m15s",
    "memory_usage": "45MB"
  }
}
```

## 🏗️ Architecture Overview

### Application Structure

```
main.go
├── Config          # Configuration management
├── App             # Application struct with dependencies
├── Middleware      # Custom middleware functions
├── Handlers        # HTTP request handlers
├── Models          # Data structures
└── Helpers         # Utility functions
```

### Middleware Stack

1. **Logger** - Request/response logging
2. **Recovery** - Panic recovery
3. **RequestID** - Unique request identification
4. **CORS** - Cross-origin resource sharing
5. **Helmet** - Security headers
6. **RealIP** - Client IP detection
7. **Gzip** - Response compression
8. **Timeout** - Request timeout handling
9. **RateLimit** - Request rate limiting
10. **Metrics** - Custom metrics collection

### Security Features

- **Token Authentication**: Simple bearer token authentication
- **Role-Based Access Control**: Admin and user roles
- **Rate Limiting**: 100 requests per minute per IP
- **Security Headers**: Comprehensive security headers via Helmet
- **Input Validation**: Request body validation
- **CORS Protection**: Configurable CORS policies

### Error Handling

The API uses a standardized error response format:

```json
{
  "success": false,
  "error": "Error message description"
}
```

Common HTTP status codes:
- `200` - Success
- `201` - Created
- `400` - Bad Request
- `401` - Unauthorized
- `403` - Forbidden
- `404` - Not Found
- `500` - Internal Server Error

## 🔍 Monitoring

### Health Checks

The `/health` endpoint provides:
- Service status
- Version information
- Timestamp
- External service connectivity

### Metrics

The `/metrics` endpoint provides:
- Request count
- Error count
- Average latency
- Uptime information

### Logging

All requests are logged with:
- HTTP method
- Request path
- Client IP
- Response status
- Request duration

## 🚀 Production Deployment

### Docker Deployment

```dockerfile
FROM golang:1.19-alpine AS builder
WORKDIR /app
COPY . .
RUN go mod download
RUN go build -o main .

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/main .
EXPOSE 8080
CMD ["./main"]
```

### Environment Variables

```bash
# Production configuration
ENV=production
PORT=8080
JWT_SECRET=your-production-secret
DB_HOST=your-database-host
DB_PORT=5432

# Optional configuration
RATE_LIMIT=1000  # requests per minute
```

### Kubernetes Deployment

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: production-api
spec:
  replicas: 3
  selector:
    matchLabels:
      app: production-api
  template:
    metadata:
      labels:
        app: production-api
    spec:
      containers:
      - name: api
        image: your-registry/production-api:latest
        ports:
        - containerPort: 8080
        env:
        - name: ENV
          value: "production"
        - name: PORT
          value: "8080"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
```

## 📚 Key Learnings

This example demonstrates:

1. **Structured Application Design**: Clean separation of concerns
2. **Security Best Practices**: Authentication, authorization, and security headers
3. **Monitoring & Observability**: Health checks, metrics, and logging
4. **Production Readiness**: Graceful shutdown, timeouts, and error handling
5. **API Design**: RESTful endpoints with proper HTTP status codes
6. **Configuration Management**: Environment-based configuration
7. **Middleware Usage**: Comprehensive middleware stack
8. **Error Handling**: Standardized error responses

## 🔗 Related Examples

- [Basic Server](../01-basic-server/) - Simple HTTP server basics
- [Middleware Showcase](../02-middleware-showcase/) - Middleware examples
- [WebSocket Chat](../03-websocket-chat/) - Real-time features
- [File Upload System](../04-file-upload-download/) - File handling
- [JSON-RPC Service](../05-json-rpc-service/) - RPC implementation

## 📖 Further Reading

- [Zoox Documentation](../../DOCUMENTATION.md)
- [Security Best Practices](../../DOCUMENTATION.md#security)
- [Production Deployment](../../DOCUMENTATION.md#deployment)
- [Monitoring Guide](../../DOCUMENTATION.md#monitoring) 