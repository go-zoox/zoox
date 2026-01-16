# 缓存系统

Zoox 提供了灵活的缓存系统，支持内存缓存和 Redis 缓存。

## 基本用法

### 获取 Cache 实例

```go
app := zoox.New()

// 从应用获取 Cache
cache := app.Cache()

// 从 Context 获取 Cache
app.Get("/data", func(ctx *zoox.Context) {
	cache := ctx.Cache()
	// 使用缓存
})
```

**说明**: Cache 实现参考 `application.go:439-446` 和 `context.go:928-935`。

## 内存缓存

### 配置内存缓存

```go
app := zoox.New()

app.Config.Cache = kv.Config{
	Engine: "memory",
}

cache := app.Cache()
```

### 使用内存缓存

```go
cache := app.Cache()

// 设置缓存（1小时过期）
cache.Set("key", "value", time.Hour)

// 获取缓存
var value string
cache.Get("key", &value)

// 检查缓存是否存在
if cache.Has("key") {
	// 缓存存在
}

// 删除缓存
cache.Delete("key")

// 清空所有缓存
cache.Clear()
```

## Redis 缓存

### 配置 Redis 缓存

```go
app := zoox.New()

// 配置 Redis
app.Config.Redis.Host = "localhost"
app.Config.Redis.Port = 6379
app.Config.Redis.Password = "password"
app.Config.Redis.DB = 0

// 配置 Cache 使用 Redis
app.Config.Cache = kv.Config{
	Engine: "redis",
	Config: &redis.Config{
		Host:     app.Config.Redis.Host,
		Port:     app.Config.Redis.Port,
		Password: app.Config.Redis.Password,
		DB:       app.Config.Redis.DB,
	},
}

cache := app.Cache()
```

### 使用 Redis 缓存

使用方式与内存缓存相同：

```go
cache := app.Cache()

// 设置缓存
cache.Set("user:1", userData, time.Hour)

// 获取缓存
var userData User
cache.Get("user:1", &userData)
```

## 缓存操作

### 设置缓存

```go
// 基本设置
cache.Set("key", "value", time.Hour)

// 设置对象
user := User{ID: 1, Name: "Alice"}
cache.Set("user:1", user, time.Hour)

// 永久缓存（不推荐）
cache.Set("key", "value", 0)
```

### 获取缓存

```go
// 获取字符串
var value string
cache.Get("key", &value)

// 获取对象
var user User
cache.Get("user:1", &user)

// 检查是否存在
if cache.Has("key") {
	var value string
	cache.Get("key", &value)
}
```

### 删除缓存

```go
// 删除单个键
cache.Delete("key")

// 删除多个键
cache.Delete("key1", "key2", "key3")
```

### 清空缓存

```go
// 清空所有缓存
cache.Clear()
```

## 缓存模式

### Cache-Aside 模式

```go
app.Get("/user/:id", func(ctx *zoox.Context) {
	id := ctx.Param().Get("id")
	cache := ctx.Cache()
	
	var user User
	cacheKey := "user:" + id
	
	// 尝试从缓存获取
	if cache.Get(cacheKey, &user) == nil {
		ctx.JSON(200, user)
		return
	}
	
	// 缓存未命中，从数据库获取
	user = getUserFromDB(id)
	
	// 写入缓存
	cache.Set(cacheKey, user, time.Hour)
	
	ctx.JSON(200, user)
})
```

### Write-Through 模式

```go
app.Post("/user", func(ctx *zoox.Context) {
	var user User
	ctx.BindJSON(&user)
	
	// 保存到数据库
	user = saveUserToDB(user)
	
	// 同时写入缓存
	cache := ctx.Cache()
	cache.Set("user:"+user.ID, user, time.Hour)
	
	ctx.JSON(200, user)
})
```

### Write-Back 模式

```go
// 写入缓存，异步写入数据库
app.Post("/user", func(ctx *zoox.Context) {
	var user User
	ctx.BindJSON(&user)
	
	cache := ctx.Cache()
	cache.Set("user:"+user.ID, user, time.Hour)
	
	// 异步写入数据库
	go func() {
		saveUserToDB(user)
	}()
	
	ctx.JSON(200, user)
})
```

## 缓存最佳实践

### 1. 使用有意义的键名

```go
// 推荐：使用命名空间
cache.Set("user:1", user, time.Hour)
cache.Set("post:123", post, time.Hour)

// 不推荐：使用简单键名
cache.Set("1", user, time.Hour)
```

### 2. 设置合理的过期时间

```go
// 根据数据特性设置过期时间
cache.Set("user:1", user, 24*time.Hour)        // 用户数据：24小时
cache.Set("session:abc", session, time.Hour)   // Session：1小时
cache.Set("temp:data", data, 5*time.Minute)    // 临时数据：5分钟
```

### 3. 处理缓存穿透

```go
app.Get("/user/:id", func(ctx *zoox.Context) {
	id := ctx.Param().Get("id")
	cache := ctx.Cache()
	cacheKey := "user:" + id
	
	var user User
	if cache.Get(cacheKey, &user) == nil {
		ctx.JSON(200, user)
		return
	}
	
	// 从数据库获取
	user = getUserFromDB(id)
	
	if user.ID == "" {
		// 用户不存在，缓存空值防止穿透
		cache.Set(cacheKey, nil, 5*time.Minute)
		ctx.Error(404, "User not found")
		return
	}
	
	// 缓存用户数据
	cache.Set(cacheKey, user, time.Hour)
	ctx.JSON(200, user)
})
```

### 4. 缓存预热

```go
func warmupCache(cache cache.Cache) {
	// 预加载热点数据
	users := getHotUsers()
	for _, user := range users {
		cache.Set("user:"+user.ID, user, time.Hour)
	}
}
```

## 下一步

- 🍪 学习 [Session 管理](session.md) - Session 和 Cookie
- 🔐 查看 [JWT 认证](jwt.md) - JWT 生成和验证
- 📝 了解 [日志系统](logger.md) - 结构化日志

---

**需要更多帮助？** 👉 [完整文档索引](../README.md)
