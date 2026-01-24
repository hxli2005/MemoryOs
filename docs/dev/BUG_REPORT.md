# MemoryOS Bug 诊断报告

## 测试环境
- 日期: 2026-01-23
- Docker 容器: PostgreSQL, Redis, Milvus (含 etcd, minio)
- 数据库: memoryos@127.0.0.1:15432

---

## 🐛 发现的严重 Bug

### Bug #2: 数据库 Schema 与代码不匹配 ❌ **[P0 - 阻塞]**

**位置**: 
- `docker/postgres/init.sql` (原始 schema)
- `internal/storage/postgres/models.go` (代码中的模型)

**问题详情**:

#### Topic Memory 表
**数据库实际字段**:
```sql
topic_memory (
  id, user_id, content, memory_type, importance, 
  access_count, last_accessed, metadata, created_at, updated_at, embedding
)
```

**代码期望字段**:
```go
TopicMemoryPO {
  ID, UserID, Title, Summary, Content, Embedding,
  Keywords, DialogueIDs, Metadata, MemoryType, ...
}
```

**缺失字段**: `title`, `summary`, `keywords`, `dialogue_ids`

**错误信息**:
```
ERROR: column "title" of relation "topic_memory" does not exist (SQLSTATE 42703)
```

**影响**: 
- ❌ 无法创建话题层记忆
- ❌ 聚合功能完全不可用

---

#### Profile Memory 表
**数据库实际字段**:
```sql
profile_memory (
  id, user_id, content, memory_type, importance,
  access_count, last_accessed, metadata, created_at, updated_at, embedding
)
```

**代码期望字段**:
```go
ProfileMemoryPO {
  ID, UserID, Preferences, Habits, Features, Embedding,
  TopicIDs, Metadata, MemoryType, ...
}
```

**缺失字段**: `preferences`, `habits`, `features`, `topic_ids`, `content`

**错误信息**:
```
ERROR: column "preferences" of relation "profile_memory" does not exist (SQLSTATE 42703)
```

**影响**: 
- ❌ 无法创建画像层记忆
- ❌ 用户画像功能完全不可用

---

**根本原因**: 
1. 数据库初始化脚本 (`init.sql`) 使用了旧的简化 schema
2. 代码中的 ORM 模型使用了新的详细 schema
3. 两者未同步更新

**修复方案**:

**方案 A: 更新数据库 schema** (推荐)
```sql
-- 为 topic_memory 添加缺失字段
ALTER TABLE topic_memory ADD COLUMN title VARCHAR(500);
ALTER TABLE topic_memory ADD COLUMN summary TEXT;
ALTER TABLE topic_memory ADD COLUMN keywords TEXT[];
ALTER TABLE topic_memory ADD COLUMN dialogue_ids BIGINT[];

-- 为 profile_memory 添加缺失字段  
ALTER TABLE profile_memory ADD COLUMN preferences JSONB;
ALTER TABLE profile_memory ADD COLUMN habits JSONB;
ALTER TABLE profile_memory ADD COLUMN features JSONB;
ALTER TABLE profile_memory ADD COLUMN topic_ids BIGINT[];
ALTER TABLE profile_memory ADD COLUMN content TEXT NOT NULL DEFAULT '';
```

**方案 B: 简化代码模型** (不推荐，功能受限)
```go
// 移除 TopicMemoryPO 中的 Title, Summary, Keywords, DialogueIDs
// 移除 ProfileMemoryPO 中的 Preferences, Habits, Features, TopicIDs
```

---

### Bug #1: 测试代码中的表名错误 ⚠️ (已文档化)

**位置**: `test/test_simple.go:74`, `test/test_integration.go`

**错误代码**:
```go
app.DB.Table("memories").Where("user_id = ?", userID).Count(&count)
```

**修复**: 使用正确的表名 `dialogue_memory`, `topic_memory`, `profile_memory`

---

### Bug #3: Milvus 会话警告 ⚠️ (非致命)

**日志**:
```
[WARN] [rootcoord/root_coord.go:1582] ["failed to updateTimeTick"] 
[error="skip ChannelTimeTickMsg from un-recognized session 4"]
```

**状态**: 不影响功能，可忽略

---

## ✅ 已验证正常的功能

### 1. 对话层 (Dialogue Layer)
- ✅ 创建记忆: 正常
- ✅ Embedding 生成: 768 维正确
- ✅ PostgreSQL 存储: 正常
- ✅ Milvus 向量存储: 正常
- ✅ 向量搜索: 正常工作

**测试结果**:
```
✅ 创建 4 条对话记忆
✅ 数据库存储: 4 条
✅ 向量搜索返回: 3 条相似记忆
✅ 相似度分数: 0.1274 ~ 0.1821
```

### 2. 混合召回 (Hybrid Recall)
- ✅ 对话层召回: 正常 (4 条)
- ⚠️  话题层召回: 返回 6 条 (但创建失败)
- ❌ 画像层召回: 0 条 (创建失败)

**分析**: 话题层返回 6 条是历史数据（可能是之前测试留下的）

### 3. 核心基础设施
- ✅ PostgreSQL 连接
- ✅ Redis 连接
- ✅ Milvus 连接
- ✅ Embedding API (降维 2560→768)

---

## 📊 完整性检查

| 模块 | 对话层 | 话题层 | 画像层 |
|------|--------|--------|--------|
| 创建记忆 | ✅ | ❌ | ❌ |
| 数据库存储 | ✅ | ❌ | ❌ |
| Milvus 存储 | ✅ | ⚠️ | ⚠️ |
| 向量搜索 | ✅ | ⚠️ | ⚠️ |
| 混合召回 | ✅ | ⚠️ | ❌ |

**综合状态**: 
- 对话层: 100% 可用 ✅
- 话题层: 0% 可用 ❌ (Schema 不匹配)
- 画像层: 0% 可用 ❌ (Schema 不匹配)

---

## 🎯 修复优先级

### P0 - 立即修复 (阻塞功能)
1. **修复数据库 Schema**
   - 更新 `docker/postgres/init.sql`
   - 或者执行 ALTER TABLE 迁移脚本
   - 预计时间: 30 分钟

### P1 - 重新测试
2. **验证话题层和画像层**
   - 修复 schema 后重新运行测试
   - 确认三层架构完整可用

### P2 - 代码优化
3. **添加 Schema 验证**
   - 启动时检查表结构
   - 自动提示 schema 不匹配

---

## 📝 测试结论

**当前状态**: ⚠️ 部分功能可用

**已验证可用**:
- ✅ 对话层完整功能 (100%)
- ✅ Embedding 生成与降维
- ✅ 向量存储与检索
- ✅ 混合召回框架

**阻塞问题**:
- ❌ 话题层和画像层无法使用 (Schema 不匹配)
- ⚠️ 三层架构无法完整验证

**建议下一步**:
1. **立即执行**: 修复数据库 schema (见修复方案 A)
2. **重新测试**: 运行完整功能验证
3. **文档更新**: 同步 schema 文档

---

**生成时间**: 2026-01-23 18:07  
**测试人员**: GitHub Copilot  
**环境**: Docker (PostgreSQL 15 + Milvus 2.3.3 + Redis 7)
