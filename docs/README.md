# Zoox 框架文档

欢迎使用 Zoox - 一个轻量级、高性能的 Go Web 框架。

## 📚 文档导航

### 🚀 快速开始（新用户必读）

如果你是第一次使用 Zoox，建议按以下顺序阅读：

1. **[安装指南](getting-started/installation.md)** - 环境要求和安装步骤
2. **[5分钟快速开始](getting-started/quick-start.md)** - 快速上手核心功能
3. **[第一个应用教程](getting-started/first-app.md)** - 完整的应用开发教程
4. **[常见场景示例](getting-started/examples.md)** - 实际场景的快速示例

### 📖 核心指南

深入了解框架的核心功能：

- **[路由系统](guides/routing.md)** - Trie树路由、路由参数、路由组
- **[中间件使用](guides/middleware.md)** - 中间件概念、内置中间件、自定义中间件
- **[Context API](guides/context.md)** - 请求处理、响应方法、数据绑定
- **[模板引擎](guides/templates.md)** - 模板渲染和自定义函数
- **[配置管理](guides/configuration.md)** - 环境变量、配置文件、TLS配置

### 🔧 内置组件

框架提供的开箱即用组件：

- **[缓存系统](components/cache.md)** - Redis和内存缓存
- **[会话管理](components/session.md)** - Session和Cookie处理
- **[JWT认证](components/jwt.md)** - JWT生成和验证
- **[国际化](components/i18n.md)** - 多语言支持
- **[日志系统](components/logger.md)** - 结构化日志

### 🛡️ 中间件

丰富的中间件生态系统：

- **[中间件概览](middleware/overview.md)** - 中间件系统介绍
- **[认证中间件](middleware/authentication.md)** - JWT、BasicAuth、BearerToken
- **[安全中间件](middleware/security.md)** - Helmet、CORS、BodyLimit
- **[性能中间件](middleware/performance.md)** - Gzip、CacheControl、StaticCache
- **[监控中间件](middleware/monitoring.md)** - Prometheus、Sentry、Logger

### 🚀 高级功能

框架的高级特性：

- **[WebSocket支持](advanced/websocket.md)** - 实时通信
- **[JSON-RPC服务](advanced/jsonrpc.md)** - JSON-RPC协议支持
- **[代理功能](advanced/proxy.md)** - 反向代理和路径重写
- **[定时任务](advanced/cron-jobs.md)** - Cron任务调度
- **[任务队列](advanced/job-queue.md)** - 后台任务处理
- **[发布订阅](advanced/pubsub.md)** - Pub/Sub消息系统
- **[消息队列](advanced/message-queue.md)** - MQ消息处理

### 📚 API 参考

完整的API文档：

- **[Application API](api-reference/application.md)** - 应用实例方法、配置、组件访问
- **[Context API](api-reference/context.md)** - 上下文方法、请求响应处理
- **[Router API](api-reference/router.md)** - 路由注册、路由组、静态文件
- **[中间件列表](api-reference/middleware-list.md)** - 所有内置中间件完整列表

### 💡 示例项目

实际项目示例，包含完整可运行的代码：

- **[RESTful API](examples/rest-api.md)** - 完整的用户管理 REST API，包含认证、CRUD操作
- **[WebSocket 应用](examples/real-time-app.md)** - WebSocket 实时聊天应用，支持多用户
- **[静态文件服务](examples/static-files.md)** - 静态文件服务示例，包含缓存、压缩等配置
- **[JSON-RPC 服务器](examples/jsonrpc-server.md)** - JSON-RPC 2.0 服务器示例，包含批量请求和错误处理
- **[API Gateway](examples/api-gateway.md)** - API 网关示例，包含路由、认证、限流、聚合等功能
- **[微服务架构](examples/microservice.md)** - 微服务架构示例，包含 Gateway 和多个服务

### 🎯 最佳实践

- **[最佳实践指南](best-practices.md)** - 项目结构、错误处理、性能优化

## 🔗 相关链接

- [GitHub 仓库](https://github.com/go-zoox/zoox)
- [GoDoc 文档](https://pkg.go.dev/github.com/go-zoox/zoox)
- [问题反馈](https://github.com/go-zoox/zoox/issues)

## 📝 文档贡献

欢迎贡献文档！如果你发现文档有任何问题或需要改进，请提交 Issue 或 Pull Request。

---

**开始使用 Zoox？** 👉 [安装指南](getting-started/installation.md) | [5分钟快速开始](getting-started/quick-start.md)
