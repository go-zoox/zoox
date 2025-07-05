# Tutorial 17: Deployment Strategies

## Overview
Learn comprehensive deployment strategies for Zoox applications, including containerization, cloud deployment, CI/CD pipelines, and production best practices.

## Learning Objectives
- Containerize Zoox applications with Docker
- Deploy to various cloud platforms
- Set up CI/CD pipelines
- Implement blue-green deployments
- Configure production monitoring
- Handle scaling and load balancing

## Prerequisites
- Complete Tutorial 16: Security Best Practices
- Basic understanding of Docker and containers
- Familiarity with cloud platforms

## Docker Containerization

### Dockerfile for Zoox Application

```dockerfile
# Multi-stage build for optimized production image
FROM golang:1.21-alpine AS builder

# Install necessary packages
RUN apk add --no-cache git ca-certificates tzdata

# Set working directory
WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./

# Download dependencies
RUN go mod download

# Copy source code
COPY . .

# Build the application
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o main .

# Production stage
FROM alpine:latest

# Install ca-certificates for HTTPS requests
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the binary from builder stage
COPY --from=builder /app/main .

# Copy configuration files if needed
COPY --from=builder /app/config ./config

# Create non-root user for security
RUN adduser -D -s /bin/sh appuser
USER appuser

# Expose port
EXPOSE 8080

# Health check
HEALTHCHECK --interval=30s --timeout=3s --start-period=5s --retries=3 \
  CMD wget --no-verbose --tries=1 --spider http://localhost:8080/health || exit 1

# Run the application
CMD ["./main"]
```

### Docker Compose for Development

```yaml
# docker-compose.yml
version: '3.8'

services:
  app:
    build: .
    ports:
      - "8080:8080"
    environment:
      - ENV=development
      - DB_HOST=postgres
      - REDIS_HOST=redis
    depends_on:
      - postgres
      - redis
    volumes:
      - ./logs:/app/logs
    restart: unless-stopped

  postgres:
    image: postgres:15-alpine
    environment:
      - POSTGRES_DB=zoox_app
      - POSTGRES_USER=user
      - POSTGRES_PASSWORD=password
    ports:
      - "5432:5432"
    volumes:
      - postgres_data:/var/lib/postgresql/data
    restart: unless-stopped

  redis:
    image: redis:7-alpine
    ports:
      - "6379:6379"
    volumes:
      - redis_data:/data
    restart: unless-stopped

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
      - ./ssl:/etc/nginx/ssl
    depends_on:
      - app
    restart: unless-stopped

volumes:
  postgres_data:
  redis_data:
```

## Kubernetes Deployment

### Kubernetes Manifests

```yaml
# k8s/namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: zoox-app

---
# k8s/configmap.yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: app-config
  namespace: zoox-app
data:
  app.env: |
    ENV=production
    PORT=8080
    LOG_LEVEL=info

---
# k8s/secret.yaml
apiVersion: v1
kind: Secret
metadata:
  name: app-secrets
  namespace: zoox-app
type: Opaque
data:
  # Base64 encoded values
  DB_PASSWORD: <base64-encoded-password>
  JWT_SECRET: <base64-encoded-jwt-secret>

---
# k8s/deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: zoox-app
  namespace: zoox-app
  labels:
    app: zoox-app
spec:
  replicas: 3
  selector:
    matchLabels:
      app: zoox-app
  template:
    metadata:
      labels:
        app: zoox-app
    spec:
      containers:
      - name: zoox-app
        image: your-registry/zoox-app:latest
        ports:
        - containerPort: 8080
        env:
        - name: DB_PASSWORD
          valueFrom:
            secretKeyRef:
              name: app-secrets
              key: DB_PASSWORD
        - name: JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: app-secrets
              key: JWT_SECRET
        envFrom:
        - configMapRef:
            name: app-config
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 30
          periodSeconds: 10
        readinessProbe:
          httpGet:
            path: /ready
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 5
        resources:
          requests:
            memory: "128Mi"
            cpu: "100m"
          limits:
            memory: "512Mi"
            cpu: "500m"

---
# k8s/service.yaml
apiVersion: v1
kind: Service
metadata:
  name: zoox-app-service
  namespace: zoox-app
spec:
  selector:
    app: zoox-app
  ports:
  - protocol: TCP
    port: 80
    targetPort: 8080
  type: ClusterIP

---
# k8s/ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: zoox-app-ingress
  namespace: zoox-app
  annotations:
    kubernetes.io/ingress.class: "nginx"
    cert-manager.io/cluster-issuer: "letsencrypt-prod"
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
spec:
  tls:
  - hosts:
    - your-domain.com
    secretName: app-tls
  rules:
  - host: your-domain.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: zoox-app-service
            port:
              number: 80

---
# k8s/hpa.yaml
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: zoox-app-hpa
  namespace: zoox-app
spec:
  scaleTargetRef:
    apiVersion: apps/v1
    kind: Deployment
    name: zoox-app
  minReplicas: 3
  maxReplicas: 10
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
```

## CI/CD Pipeline

### GitHub Actions Workflow

```yaml
# .github/workflows/deploy.yml
name: Build and Deploy

on:
  push:
    branches: [ main, develop ]
  pull_request:
    branches: [ main ]

env:
  REGISTRY: ghcr.io
  IMAGE_NAME: ${{ github.repository }}

jobs:
  test:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Set up Go
      uses: actions/setup-go@v3
      with:
        go-version: 1.21
    
    - name: Run tests
      run: |
        go mod download
        go test -v ./...
        go test -race -coverprofile=coverage.out ./...
    
    - name: Upload coverage
      uses: codecov/codecov-action@v3
      with:
        file: ./coverage.out

  security:
    runs-on: ubuntu-latest
    steps:
    - uses: actions/checkout@v3
    
    - name: Run security scan
      uses: securecodewarrior/github-action-add-sarif@v1
      with:
        sarif-file: 'gosec-report.sarif'
    
    - name: Run Trivy vulnerability scanner
      uses: aquasecurity/trivy-action@master
      with:
        scan-type: 'fs'
        scan-ref: '.'

  build:
    needs: [test, security]
    runs-on: ubuntu-latest
    permissions:
      contents: read
      packages: write
    
    steps:
    - name: Checkout
      uses: actions/checkout@v3
    
    - name: Log in to Container Registry
      uses: docker/login-action@v2
      with:
        registry: ${{ env.REGISTRY }}
        username: ${{ github.actor }}
        password: ${{ secrets.GITHUB_TOKEN }}
    
    - name: Extract metadata
      id: meta
      uses: docker/metadata-action@v4
      with:
        images: ${{ env.REGISTRY }}/${{ env.IMAGE_NAME }}
        tags: |
          type=ref,event=branch
          type=ref,event=pr
          type=sha,prefix=commit-
    
    - name: Build and push Docker image
      uses: docker/build-push-action@v4
      with:
        context: .
        push: true
        tags: ${{ steps.meta.outputs.tags }}
        labels: ${{ steps.meta.outputs.labels }}

  deploy-staging:
    needs: build
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/develop'
    environment: staging
    
    steps:
    - name: Deploy to staging
      run: |
        echo "Deploying to staging environment"
        # Add deployment commands here

  deploy-production:
    needs: build
    runs-on: ubuntu-latest
    if: github.ref == 'refs/heads/main'
    environment: production
    
    steps:
    - name: Deploy to production
      run: |
        echo "Deploying to production environment"
        # Add deployment commands here
```

## Configuration Management

### Environment-based Configuration

```go
// config/config.go
package config

import (
    "os"
    "strconv"
    "time"
)

type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    Redis    RedisConfig
    JWT      JWTConfig
    Logging  LoggingConfig
}

type ServerConfig struct {
    Port         string
    ReadTimeout  time.Duration
    WriteTimeout time.Duration
    IdleTimeout  time.Duration
}

type DatabaseConfig struct {
    Host     string
    Port     string
    User     string
    Password string
    Name     string
    SSLMode  string
    MaxConns int
}

type RedisConfig struct {
    Host     string
    Port     string
    Password string
    DB       int
}

type JWTConfig struct {
    Secret string
    Expiry time.Duration
}

type LoggingConfig struct {
    Level  string
    Format string
}

func Load() *Config {
    return &Config{
        Server: ServerConfig{
            Port:         getEnv("PORT", "8080"),
            ReadTimeout:  getDuration("READ_TIMEOUT", 15*time.Second),
            WriteTimeout: getDuration("WRITE_TIMEOUT", 15*time.Second),
            IdleTimeout:  getDuration("IDLE_TIMEOUT", 60*time.Second),
        },
        Database: DatabaseConfig{
            Host:     getEnv("DB_HOST", "localhost"),
            Port:     getEnv("DB_PORT", "5432"),
            User:     getEnv("DB_USER", "user"),
            Password: getEnv("DB_PASSWORD", "password"),
            Name:     getEnv("DB_NAME", "zoox_app"),
            SSLMode:  getEnv("DB_SSL_MODE", "disable"),
            MaxConns: getInt("DB_MAX_CONNS", 25),
        },
        Redis: RedisConfig{
            Host:     getEnv("REDIS_HOST", "localhost"),
            Port:     getEnv("REDIS_PORT", "6379"),
            Password: getEnv("REDIS_PASSWORD", ""),
            DB:       getInt("REDIS_DB", 0),
        },
        JWT: JWTConfig{
            Secret: getEnv("JWT_SECRET", "change-this-secret"),
            Expiry: getDuration("JWT_EXPIRY", 24*time.Hour),
        },
        Logging: LoggingConfig{
            Level:  getEnv("LOG_LEVEL", "info"),
            Format: getEnv("LOG_FORMAT", "json"),
        },
    }
}

func getEnv(key, defaultValue string) string {
    if value := os.Getenv(key); value != "" {
        return value
    }
    return defaultValue
}

func getInt(key string, defaultValue int) int {
    if value := os.Getenv(key); value != "" {
        if intValue, err := strconv.Atoi(value); err == nil {
            return intValue
        }
    }
    return defaultValue
}

func getDuration(key string, defaultValue time.Duration) time.Duration {
    if value := os.Getenv(key); value != "" {
        if duration, err := time.ParseDuration(value); err == nil {
            return duration
        }
    }
    return defaultValue
}
```

## Health Checks and Monitoring

```go
// health/health.go
package health

import (
    "context"
    "database/sql"
    "net/http"
    "time"

    "github.com/go-redis/redis/v8"
    "github.com/go-zoox/zoox"
)

type HealthChecker struct {
    db    *sql.DB
    redis *redis.Client
}

func NewHealthChecker(db *sql.DB, redis *redis.Client) *HealthChecker {
    return &HealthChecker{
        db:    db,
        redis: redis,
    }
}

func (h *HealthChecker) HealthHandler() zoox.HandlerFunc {
    return func(ctx *zoox.Context) {
        status := h.checkHealth()
        
        if status["status"] == "healthy" {
            ctx.JSON(http.StatusOK, status)
        } else {
            ctx.JSON(http.StatusServiceUnavailable, status)
        }
    }
}

func (h *HealthChecker) ReadinessHandler() zoox.HandlerFunc {
    return func(ctx *zoox.Context) {
        ready := h.checkReadiness()
        
        if ready["ready"] == true {
            ctx.JSON(http.StatusOK, ready)
        } else {
            ctx.JSON(http.StatusServiceUnavailable, ready)
        }
    }
}

func (h *HealthChecker) checkHealth() map[string]interface{} {
    checks := make(map[string]interface{})
    
    // Database health check
    if h.db != nil {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        
        if err := h.db.PingContext(ctx); err != nil {
            checks["database"] = map[string]interface{}{
                "status": "unhealthy",
                "error":  err.Error(),
            }
        } else {
            checks["database"] = map[string]interface{}{
                "status": "healthy",
            }
        }
    }
    
    // Redis health check
    if h.redis != nil {
        ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
        defer cancel()
        
        if err := h.redis.Ping(ctx).Err(); err != nil {
            checks["redis"] = map[string]interface{}{
                "status": "unhealthy",
                "error":  err.Error(),
            }
        } else {
            checks["redis"] = map[string]interface{}{
                "status": "healthy",
            }
        }
    }
    
    // Overall status
    overallStatus := "healthy"
    for _, check := range checks {
        if checkMap, ok := check.(map[string]interface{}); ok {
            if checkMap["status"] == "unhealthy" {
                overallStatus = "unhealthy"
                break
            }
        }
    }
    
    return map[string]interface{}{
        "status":    overallStatus,
        "timestamp": time.Now(),
        "checks":    checks,
    }
}

func (h *HealthChecker) checkReadiness() map[string]interface{} {
    health := h.checkHealth()
    
    return map[string]interface{}{
        "ready":     health["status"] == "healthy",
        "timestamp": time.Now(),
        "checks":    health["checks"],
    }
}
```

## Blue-Green Deployment

```bash
#!/bin/bash
# scripts/blue-green-deploy.sh

set -e

NAMESPACE="zoox-app"
NEW_VERSION=$1
CURRENT_SERVICE="zoox-app-service"

if [ -z "$NEW_VERSION" ]; then
    echo "Usage: $0 <version>"
    exit 1
fi

echo "Starting blue-green deployment for version $NEW_VERSION"

# Check current deployment
CURRENT_DEPLOYMENT=$(kubectl get service $CURRENT_SERVICE -n $NAMESPACE -o jsonpath='{.spec.selector.version}')
echo "Current deployment: $CURRENT_DEPLOYMENT"

# Determine new deployment name
if [ "$CURRENT_DEPLOYMENT" = "blue" ]; then
    NEW_DEPLOYMENT="green"
else
    NEW_DEPLOYMENT="blue"
fi

echo "Deploying to: $NEW_DEPLOYMENT"

# Update deployment manifest with new version
sed "s/{{VERSION}}/$NEW_VERSION/g; s/{{COLOR}}/$NEW_DEPLOYMENT/g" k8s/deployment-template.yaml > k8s/deployment-$NEW_DEPLOYMENT.yaml

# Deploy new version
kubectl apply -f k8s/deployment-$NEW_DEPLOYMENT.yaml

# Wait for deployment to be ready
echo "Waiting for deployment to be ready..."
kubectl wait --for=condition=available --timeout=300s deployment/zoox-app-$NEW_DEPLOYMENT -n $NAMESPACE

# Run health checks
echo "Running health checks..."
POD_NAME=$(kubectl get pods -n $NAMESPACE -l version=$NEW_DEPLOYMENT -o jsonpath='{.items[0].metadata.name}')
kubectl port-forward -n $NAMESPACE $POD_NAME 8080:8080 &
PF_PID=$!

sleep 5

# Check health endpoint
if curl -f http://localhost:8080/health; then
    echo "Health check passed"
    kill $PF_PID
else
    echo "Health check failed"
    kill $PF_PID
    kubectl delete deployment zoox-app-$NEW_DEPLOYMENT -n $NAMESPACE
    exit 1
fi

# Switch traffic to new deployment
echo "Switching traffic to new deployment..."
kubectl patch service $CURRENT_SERVICE -n $NAMESPACE -p '{"spec":{"selector":{"version":"'$NEW_DEPLOYMENT'"}}}'

echo "Deployment completed successfully"

# Clean up old deployment (optional)
read -p "Do you want to delete the old deployment? (y/n): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    kubectl delete deployment zoox-app-$CURRENT_DEPLOYMENT -n $NAMESPACE
    echo "Old deployment deleted"
fi
```

## Monitoring and Observability

### Prometheus Metrics

```go
// monitoring/metrics.go
package monitoring

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    RequestsTotal = promauto.NewCounterVec(
        prometheus.CounterOpts{
            Name: "http_requests_total",
            Help: "Total number of HTTP requests",
        },
        []string{"method", "endpoint", "status"},
    )

    RequestDuration = promauto.NewHistogramVec(
        prometheus.HistogramOpts{
            Name: "http_request_duration_seconds",
            Help: "HTTP request duration in seconds",
            Buckets: prometheus.DefBuckets,
        },
        []string{"method", "endpoint"},
    )

    ActiveConnections = promauto.NewGauge(
        prometheus.GaugeOpts{
            Name: "http_active_connections",
            Help: "Number of active HTTP connections",
        },
    )
)

func MetricsMiddleware() zoox.HandlerFunc {
    return func(ctx *zoox.Context) {
        timer := prometheus.NewTimer(RequestDuration.WithLabelValues(
            ctx.Method(),
            ctx.Request().URL.Path,
        ))
        defer timer.ObserveDuration()

        ctx.Next()

        RequestsTotal.WithLabelValues(
            ctx.Method(),
            ctx.Request().URL.Path,
            fmt.Sprintf("%d", ctx.Writer.Status()),
        ).Inc()
    }
}
```

## Production Checklist

### Pre-deployment Checklist
- [ ] All tests passing
- [ ] Security scan completed
- [ ] Performance benchmarks met
- [ ] Configuration validated
- [ ] Database migrations tested
- [ ] Monitoring setup verified
- [ ] Backup procedures tested
- [ ] Rollback plan prepared

### Post-deployment Checklist
- [ ] Health checks passing
- [ ] Metrics being collected
- [ ] Logs being generated
- [ ] Performance within expected range
- [ ] Error rates acceptable
- [ ] User traffic flowing correctly
- [ ] Backup systems functional

## Key Takeaways

1. **Containerization**: Use multi-stage Docker builds for optimized images
2. **Orchestration**: Leverage Kubernetes for scalable deployments
3. **CI/CD**: Implement automated testing and deployment pipelines
4. **Configuration**: Use environment-based configuration management
5. **Health Checks**: Implement comprehensive health and readiness checks
6. **Monitoring**: Set up proper metrics and observability
7. **Deployment Strategies**: Use blue-green or canary deployments for zero downtime

## Next Steps

- Tutorial 18: Production Monitoring - Advanced monitoring techniques
- Learn about service mesh (Istio, Linkerd)
- Explore GitOps deployment strategies
- Study disaster recovery planning
- Implement advanced security scanning

## Additional Resources

- [Docker Best Practices](https://docs.docker.com/develop/dev-best-practices/)
- [Kubernetes Documentation](https://kubernetes.io/docs/)
- [GitHub Actions](https://docs.github.com/en/actions)
- [Prometheus Monitoring](https://prometheus.io/docs/) 