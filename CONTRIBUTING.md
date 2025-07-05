# 贡献指南

感谢你对 Zoox 框架的兴趣！我们欢迎各种形式的贡献，包括代码、文档、示例、问题报告和功能建议。

## 🤝 贡献方式

### 1. 文档改进
- 修正错误和拼写
- 完善示例代码
- 添加新的使用场景
- 翻译文档

### 2. 示例代码
- 创建新的示例应用
- 改进现有示例
- 添加最佳实践示例
- 性能优化示例

### 3. 问题报告
- 报告文档中的错误
- 提出改进建议
- 请求新功能文档

### 4. 代码贡献
- 修复 bug
- 添加新功能
- 性能优化
- 测试覆盖率提升

## 📋 贡献流程

### 1. 准备工作

```bash
# 1. Fork 仓库到你的 GitHub 账户

# 2. 克隆你的 fork
git clone https://github.com/your-username/zoox.git
cd zoox

# 3. 添加上游仓库
git remote add upstream https://github.com/go-zoox/zoox.git

# 4. 创建开发分支
git checkout -b feature/your-feature-name
```

### 2. 开发规范

#### 代码风格
```bash
# 使用 gofmt 格式化代码
go fmt ./...

# 使用 golint 检查代码
golint ./...

# 使用 go vet 检查代码
go vet ./...
```

#### 提交规范
使用 [Conventional Commits](https://www.conventionalcommits.org/) 规范：

```
<type>[optional scope]: <description>

[optional body]

[optional footer(s)]
```

**类型 (type):**
- `feat`: 新功能
- `fix`: 修复 bug
- `docs`: 文档更改
- `style`: 代码格式（不影响代码运行的变动）
- `refactor`: 重构（既不是新增功能，也不是修复 bug 的代码变动）
- `test`: 增加测试
- `chore`: 构建过程或辅助工具的变动

**示例:**
```
feat(middleware): add rate limiting middleware
fix(router): fix path parameter parsing bug
docs(readme): update installation instructions
test(context): add tests for context binding methods
```

### 3. 测试要求

```bash
# 运行所有测试
go test ./...

# 运行测试并生成覆盖率报告
go test -cover ./...

# 运行特定包的测试
go test ./middleware
```

#### 测试指南
- 所有新功能必须包含测试
- 测试覆盖率应保持在 80% 以上
- 测试应该清晰、可读且可维护
- 使用表驱动测试处理多种情况

### 4. 文档要求

#### 代码文档
```go
// ExampleFunction 演示如何使用某个功能
// 参数 param1 用于...
// 返回值表示...
func ExampleFunction(param1 string) error {
    // 实现...
}
```

#### README 和文档
- 所有新功能必须在文档中说明
- 提供清晰的使用示例
- 更新相关的 API 参考
- 确保文档与代码同步

### 5. 提交和审查

```bash
# 1. 提交更改
git add .
git commit -m "feat(middleware): add rate limiting middleware"

# 2. 推送到你的 fork
git push origin feature/your-feature-name

# 3. 创建 Pull Request
# 在 GitHub 上创建 PR，详细描述你的更改
```

#### PR 要求
- 清晰的标题和描述
- 关联相关的 issue
- 包含测试和文档
- 通过所有 CI 检查

## 🎯 贡献重点领域

### 高优先级
1. **测试覆盖率提升**
   - 为现有功能添加测试
   - 边界情况测试
   - 性能测试

2. **文档完善**
   - API 文档补充
   - 使用示例增加
   - 最佳实践指南

3. **性能优化**
   - 路由性能优化
   - 内存使用优化
   - 并发安全改进

### 中等优先级
1. **中间件扩展**
   - 新的中间件开发
   - 现有中间件改进
   - 中间件文档

2. **示例应用**
   - 真实场景示例
   - 集成示例
   - 部署示例

### 低优先级
1. **工具改进**
   - 开发工具优化
   - 构建脚本改进
   - CI/CD 优化

## 📝 问题报告

### 报告 Bug
使用 [Bug 报告模板](https://github.com/go-zoox/zoox/issues/new?template=bug_report.md)

**包含信息:**
- Go 版本
- Zoox 版本
- 操作系统
- 重现步骤
- 期望行为
- 实际行为
- 错误日志

### 功能请求
使用 [功能请求模板](https://github.com/go-zoox/zoox/issues/new?template=feature_request.md)

**包含信息:**
- 功能描述
- 使用场景
- 预期 API 设计
- 相关资源

## 🏆 贡献者认可

### 贡献者列表
所有贡献者都会在 CONTRIBUTORS.md 中列出。

### 贡献统计
- 代码贡献
- 文档贡献
- 问题报告
- 功能建议

## 📞 联系方式

### 讨论
- GitHub Discussions
- Issues
- Pull Requests

### 社区
- Slack/Discord（如果有）
- 邮件列表
- 社交媒体

## 📜 行为准则

我们致力于营造一个开放、友好的社区环境：

1. **尊重他人** - 尊重不同的观点和经验
2. **建设性反馈** - 提供有建设性的批评和建议
3. **协作精神** - 共同努力改进项目
4. **包容性** - 欢迎所有背景的贡献者
5. **专业性** - 保持专业和礼貌的交流

## 🎉 开始贡献

1. 浏览 [Issues](https://github.com/go-zoox/zoox/issues) 寻找合适的任务
2. 查看 [Good First Issue](https://github.com/go-zoox/zoox/labels/good%20first%20issue) 标签
3. 阅读相关文档和代码
4. 开始你的第一个贡献！

感谢你的贡献！🚀 