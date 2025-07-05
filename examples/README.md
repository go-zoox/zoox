# Zoox Framework Examples

This directory contains comprehensive examples demonstrating various features and capabilities of the Zoox Go web framework.

## Quick Start

Each example is self-contained and includes:
- Complete working code
- Detailed README with setup instructions
- API documentation where applicable
- Dependencies and configuration

## Examples Overview

### 🟢 Beginner Level

#### 1. Basic Server (`01-basic-server/`)
**Difficulty:** ⭐⭐☆☆☆  
**Features:** REST API, CRUD operations, Route groups, Basic middleware  
A complete REST API with user management, demonstrating fundamental Zoox concepts including routing, middleware, and JSON handling.

**What you'll learn:**
- Setting up a basic Zoox server
- Creating REST endpoints
- Working with JSON requests/responses
- Using route groups
- Basic error handling

#### 2. Middleware Showcase (`02-middleware-showcase/`)
**Difficulty:** ⭐⭐⭐☆☆  
**Features:** All built-in middleware, Security, Performance, Custom middleware  
Comprehensive demonstration of Zoox's built-in middleware including security, performance optimization, and custom middleware creation.

**What you'll learn:**
- Using built-in middleware (CORS, Logger, Recovery, etc.)
- Security middleware (Helmet, Rate Limiting)
- Performance middleware (Gzip, Caching)
- Creating custom middleware

### 🟡 Intermediate Level

#### 3. WebSocket Chat (`03-websocket-chat/`)
**Difficulty:** ⭐⭐⭐☆☆  
**Features:** WebSocket, Real-time communication, Connection management  
Real-time chat application demonstrating WebSocket integration with user management and message broadcasting.

**What you'll learn:**
- WebSocket implementation
- Real-time data handling
- Connection lifecycle management
- Client-server communication patterns

#### 4. File Upload/Download System (`04-file-upload-download/`)
**Difficulty:** ⭐⭐⭐⭐☆  
**Features:** File handling, Chunked uploads, Validation, Security  
Complete file management system with upload, download, validation, and security features.

**What you'll learn:**
- File upload handling (single/multiple)
- Chunked file transfers
- File validation and security
- File metadata management

### 🔴 Advanced Level

#### 5. JSON-RPC Service (`05-json-rpc-service/`)
**Difficulty:** ⭐⭐⭐⭐☆  
**Features:** JSON-RPC, Service architecture, Error handling  
Professional JSON-RPC service with math and user operations, custom error handling, and interactive testing interface.

**What you'll learn:**
- JSON-RPC protocol implementation
- Service-oriented architecture
- Custom error handling
- Method registration and discovery

#### 6. Production API (`06-production-api/`)
**Difficulty:** ⭐⭐⭐⭐⭐  
**Features:** Authentication, Authorization, Monitoring, Security, Deployment  
Production-ready API with comprehensive security, monitoring, and deployment configurations.

**What you'll learn:**
- JWT authentication and RBAC
- Production security practices
- Monitoring and observability
- Deployment strategies
- Clean architecture patterns

## Learning Paths

### 🎯 Path 1: Web Development Beginner
1. **Basic Server** → Learn fundamental concepts
2. **Middleware Showcase** → Understand request processing
3. **File Upload/Download** → Handle file operations
4. **Production API** → Apply production practices

### 🎯 Path 2: API Development Focus
1. **Basic Server** → REST API fundamentals
2. **JSON-RPC Service** → Alternative API patterns
3. **Production API** → Professional implementation
4. **Middleware Showcase** → Advanced request handling

### 🎯 Path 3: Real-time Applications
1. **Basic Server** → Foundation
2. **WebSocket Chat** → Real-time communication
3. **Middleware Showcase** → Performance optimization
4. **Production API** → Scalable architecture

### 🎯 Path 4: Production Deployment
1. **Basic Server** → Core concepts
2. **Middleware Showcase** → Security and performance
3. **Production API** → Complete production setup
4. **File Upload/Download** → File handling best practices

## Features Matrix

| Example | REST API | WebSocket | Auth | File Handling | JSON-RPC | Monitoring | Deployment |
|---------|----------|-----------|------|---------------|----------|------------|------------|
| Basic Server | ✅ | ❌ | Basic | ❌ | ❌ | ❌ | ❌ |
| Middleware | ✅ | ❌ | ✅ | ❌ | ❌ | ✅ | ❌ |
| WebSocket Chat | ✅ | ✅ | Basic | ❌ | ❌ | ❌ | ❌ |
| File System | ✅ | ❌ | ✅ | ✅ | ❌ | ❌ | ❌ |
| JSON-RPC | ✅ | ❌ | ❌ | ❌ | ✅ | ❌ | ❌ |
| Production | ✅ | ❌ | ✅ | ✅ | ❌ | ✅ | ✅ |

## Getting Started

1. **Prerequisites:**
   - Go 1.19 or higher
   - Git
   - Basic understanding of Go and web concepts

2. **Setup:**
   ```bash
   # Clone the repository
   git clone https://github.com/go-zoox/zoox.git
   cd zoox/examples
   
   # Choose an example
   cd 01-basic-server
   
   # Install dependencies
   go mod tidy
   
   # Run the example
   go run main.go
   ```

3. **Testing:**
   Each example includes test endpoints or interfaces. Check the individual README files for specific testing instructions.

## Troubleshooting

### Common Issues

**Module not found errors:**
```bash
# Ensure you're in the correct directory
cd examples/[example-name]
go mod tidy
```

**Port already in use:**
```bash
# Kill existing processes
sudo lsof -ti:8080 | xargs kill -9
```

**Permission denied:**
```bash
# Check file permissions
chmod +x main.go
```

### Getting Help

- **Documentation:** Check individual example README files
- **Tutorials:** See `../tutorials/README.md` for step-by-step guides
- **Issues:** Report bugs on GitHub
- **Community:** Join discussions on GitHub Discussions

## Contributing

We welcome contributions! To add a new example:

1. Create a new directory following the naming convention
2. Include a complete `main.go` with comments
3. Add a detailed `README.md`
4. Update this index file
5. Test thoroughly
6. Submit a pull request

### Example Structure
```
examples/
├── XX-example-name/
│   ├── main.go          # Main application code
│   ├── README.md        # Detailed documentation
│   ├── go.mod          # Dependencies (if needed)
│   └── static/         # Static files (if needed)
```

## Next Steps

After exploring these examples, check out:
- **Tutorials** (`../tutorials/`) for step-by-step learning
- **Main Documentation** (`../DOCUMENTATION.md`) for API reference
- **Contributing Guide** (`../CONTRIBUTING.md`) to contribute back

Happy coding with Zoox! 🚀 