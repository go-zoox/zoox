# gormx 查询参数适配

Zoox 提供了 `gormx` 的上下文适配层，用于把 HTTP Query/Param 解析为 `gormx` 的 `Params`、`Where`、`OrderBy`、`ListParams`。

## 安装

`zoox` 依赖 `gormx` 后即可使用：

```go
import (
	"github.com/go-zoox/zoox"
	contextgormx "github.com/go-zoox/zoox/components/context/gormx"
)
```

## 基本用法

```go
app.Get("/users/:id", func(ctx *zoox.Context) {
	params := contextgormx.NewParams(ctx)

	// 读取路由参数 id
	id, err := params.ID()
	if err != nil {
		ctx.Error(400, err.Error())
		return
	}

	// 读取分页、过滤、排序
	list, err := params.GetList()
	if err != nil {
		ctx.Error(400, err.Error())
		return
	}

	whereQuery, whereArgs, _ := list.Where.Build()

	ctx.JSON(200, zoox.H{
		"id":         id,
		"page":       list.Page,
		"page_size":  list.PageSize,
		"whereQuery": whereQuery,
		"whereArgs":  whereArgs,
		"orderBy":    list.OrderBy.Build(),
	})
})
```

## 支持的查询参数格式

- 分页：`?page=2&pageSize=20`
- 排序：`?orderBy=created_at:desc,id:asc`
- 模糊匹配：`?name=alice:*`
- 集合查询：`?status=active,pending:in`
- 区间查询：`?age=18,65:range[)`（左闭右开）

## 设计说明

- `gormx` 从 `v1.7.0` 开始使用通用 `ParamsContext` 接口，不再直接依赖 `zoox.Context`。
- `zoox/components/context/gormx` 负责把 `zoox.Context` 适配为 `gormx` 需要的接口能力。
- 这种拆分让 `gormx` 保持数据库查询增强库定位，`zoox` 负责 Web 上下文集成。

## 下一步

- 📝 查看 [Context API](../guides/context.md) - 请求参数与绑定
- 🚀 参考 [RESTful API 示例](../examples/rest-api.md) - 组合实际业务路由

---

**需要更多帮助？** 👉 [完整文档索引](../README.md)
