# 定时任务

Zoox 集成了 Cron 任务调度功能，可以轻松创建定时执行的任务。

## 基本用法

### 创建定时任务

```go
package main

import (
	"github.com/go-zoox/zoox"
)

func main() {
	app := zoox.New()
	
	// 获取 Cron 实例
	cron := app.Cron()
	
	// 添加定时任务
	cron.AddJob("daily-cleanup", "0 0 * * *", func() error {
		// 每天午夜执行
		app.Logger().Info("Running daily cleanup")
		return nil
	})
	
	app.Run(":8080")
}
```

**说明**: Cron 实现参考 `application.go:448-455`。

## Cron 表达式

Zoox 使用标准的 Cron 表达式格式：

```
秒 分 时 日 月 星期
*  *  *  *  *  *
```

### 常用表达式

```go
// 每分钟执行
cron.AddJob("every-minute", "* * * * *", handler)

// 每小时执行
cron.AddJob("every-hour", "0 * * * *", handler)

// 每天午夜执行
cron.AddJob("daily", "0 0 * * *", handler)

// 每周一执行
cron.AddJob("weekly", "0 0 * * 1", handler)

// 每月1号执行
cron.AddJob("monthly", "0 0 1 * *", handler)

// 每5分钟执行
cron.AddJob("every-5-minutes", "*/5 * * * *", handler)

// 工作日上午9点执行
cron.AddJob("workday-9am", "0 9 * * 1-5", handler)
```

## 任务处理

### 基本任务

```go
cron.AddJob("simple-task", "0 * * * *", func() error {
	app.Logger().Info("Task executed")
	return nil
})
```

### 带错误处理的任务

```go
cron.AddJob("task-with-error", "0 * * * *", func() error {
	// 执行任务
	if err := doSomething(); err != nil {
		app.Logger().Errorf("Task failed: %v", err)
		return err
	}
	return nil
})
```

### 长时间运行的任务

```go
cron.AddJob("long-task", "0 0 * * *", func() error {
	// 使用 goroutine 处理长时间任务
	go func() {
		// 长时间运行的任务
		processLargeData()
	}()
	return nil
})
```

## 在 Context 中使用

在请求处理中使用 Cron：

```go
app.Get("/trigger", func(ctx *zoox.Context) {
	cron := ctx.Cron()
	
	// 添加一次性任务
	cron.AddJob("one-time", "* * * * *", func() error {
		ctx.Logger().Info("One-time task executed")
		return nil
	})
	
	ctx.JSON(200, zoox.H{"message": "Task scheduled"})
})
```

**说明**: Context 中的 Cron 参考 `context.go:937-944`。

## 完整示例

```go
package main

import (
	"time"
	
	"github.com/go-zoox/zoox"
)

func main() {
	app := zoox.New()
	
	cron := app.Cron()
	
	// 每分钟执行的任务
	cron.AddJob("heartbeat", "* * * * *", func() error {
		app.Logger().Info("Heartbeat: " + time.Now().String())
		return nil
	})
	
	// 每小时执行的数据清理
	cron.AddJob("hourly-cleanup", "0 * * * *", func() error {
		app.Logger().Info("Running hourly cleanup")
		// 清理逻辑
		return nil
	})
	
	// 每天午夜执行的备份
	cron.AddJob("daily-backup", "0 0 * * *", func() error {
		app.Logger().Info("Running daily backup")
		// 备份逻辑
		return nil
	})
	
	// 每周一执行的报告生成
	cron.AddJob("weekly-report", "0 0 * * 1", func() error {
		app.Logger().Info("Generating weekly report")
		// 报告生成逻辑
		return nil
	})
	
	app.Run(":8080")
}
```

## 任务管理

### 获取所有任务

```go
jobs := cron.GetJobs()
for name, job := range jobs {
	app.Logger().Infof("Job: %s, Schedule: %s", name, job.Schedule())
}
```

### 删除任务

```go
cron.RemoveJob("task-name")
```

### 暂停/恢复任务

```go
// 暂停任务
cron.PauseJob("task-name")

// 恢复任务
cron.ResumeJob("task-name")
```

## 最佳实践

### 1. 错误处理

```go
cron.AddJob("task", "0 * * * *", func() error {
	if err := doTask(); err != nil {
		// 记录错误但继续执行
		app.Logger().Errorf("Task error: %v", err)
		// 可以选择返回错误停止任务，或返回 nil 继续
		return nil
	}
	return nil
})
```

### 2. 任务超时

```go
cron.AddJob("task", "0 * * * *", func() error {
	done := make(chan error, 1)
	
	go func() {
		done <- doLongTask()
	}()
	
	select {
	case err := <-done:
		return err
	case <-time.After(5 * time.Minute):
		return errors.New("task timeout")
	}
})
```

### 3. 任务互斥

```go
var taskMutex sync.Mutex

cron.AddJob("task", "0 * * * *", func() error {
	taskMutex.Lock()
	defer taskMutex.Unlock()
	
	// 确保同一时间只有一个任务实例运行
	return doTask()
})
```

### 4. 任务日志

```go
cron.AddJob("task", "0 * * * *", func() error {
	start := time.Now()
	app.Logger().Info("Task started")
	
	err := doTask()
	
	duration := time.Since(start)
	if err != nil {
		app.Logger().Errorf("Task failed after %v: %v", duration, err)
	} else {
		app.Logger().Infof("Task completed in %v", duration)
	}
	
	return err
})
```

## 与 JobQueue 结合

可以将 Cron 任务与 JobQueue 结合使用：

```go
cron.AddJob("schedule-jobs", "0 * * * *", func() error {
	queue := app.JobQueue()
	
	// 将任务添加到队列
	queue.Add("process-data", map[string]interface{}{
		"timestamp": time.Now(),
	})
	
	return nil
})
```

## 下一步

- 📦 学习 [任务队列](job-queue.md) - 后台任务处理
- 📡 查看 [发布订阅](pubsub.md) - 事件驱动架构
- 🚀 探索 [其他高级功能](websocket.md) - WebSocket 等

---

**需要更多帮助？** 👉 [完整文档索引](../README.md)
