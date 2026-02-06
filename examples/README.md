# Examples

本目录包含 MemoryOS 的使用示例和测试文件。

## 📁 文件说明

### api_test.http
HTTP 请求测试文件，可以使用 VS Code 的 REST Client 扩展或 IntelliJ IDEA 直接执行。

**功能覆盖**:
- ✅ 健康检查 (`GET /health`)
- ✅ 创建记忆 (`POST /api/v1/memories`)
- ✅ 查询记忆 (`GET /api/v1/memories`)
- ✅ 混合召回 (`POST /api/v1/recall/hybrid`)
- ✅ 查看监控指标 (`GET /metrics`)

### 使用方法

**VS Code (推荐)**:
1. 安装扩展: [REST Client](https://marketplace.visualstudio.com/items?itemName=humao.rest-client)
2. 打开 `api_test.http`
3. 点击请求上方的 "Send Request"

**curl 命令**:
```bash
# 健康检查
curl http://localhost:8080/health

# 创建记忆
curl -X POST http://localhost:8080/api/v1/memories \
  -H "Content-Type: application/json" \
  -d '{
    "user_id": "user123",
    "content": "今天学习了 Go 语言的并发编程",
    "metadata": {"source": "chat"}
  }'
```

## 🔗 相关文档

- [API 文档](../docs/api/API_GUIDE.md)
- [部署指南](../docs/deployment/DEPLOYMENT_GUIDE.md)
- [项目结构](../docs/PROJECT_OVERVIEW.md)
