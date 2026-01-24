# Changelog

本文档记录 MemoryOS 项目的所有重要变更。

格式基于 [Keep a Changelog](https://keepachangelog.com/zh-CN/1.0.0/)，
版本号遵循 [Semantic Versioning](https://semver.org/lang/zh-CN/)。

## [Unreleased]

### Added
- OpenAI LLM Client 支持（兼容灵娅AI中转接口）
- 混合召回策略（HybridRecall）
- LLM 聚合功能（对话→话题→画像）
- 项目结构标准化重构

### Changed
- 重组测试文件到 test/integration/ 和 test/e2e/
- 重组配置文件到 config/test/
- 重组文档到 docs/api/、docs/guides/、docs/dev/
- 移动脚本到 scripts/dev/、scripts/build/、scripts/test/

### Fixed
- PostgreSQL Schema 不完整问题
- UTF-8 字符截断问题
- ProfileMemoryPO 缺失字段
- OpenAI API BaseURL 路径问题

### Removed
- 空目录 internal/api/
- 根目录散乱的 .bat 脚本
- examples/chatbot/ 示例（API 已变更，功能被集成测试覆盖）
- docs/guides/CHATBOT_USAGE.md（已过时）

## [0.1.0] - 2026-01-20

### Added
- 三层记忆架构（Dialogue/Topic/Profile）
- PostgreSQL 元数据存储
- Milvus 向量存储
- Gemini LLM Client
- 基础 HTTP API
- Docker Compose 部署方案

### Security
- 配置文件敏感信息保护（.gitignore）

---

## 版本规范

- **Major (X.0.0)**: 不兼容的 API 变更
- **Minor (0.X.0)**: 向后兼容的功能新增
- **Patch (0.0.X)**: 向后兼容的 Bug 修复

## 类型说明

- `Added`: 新增功能
- `Changed`: 既有功能的变更
- `Deprecated`: 即将移除的功能
- `Removed`: 已移除的功能
- `Fixed`: Bug 修复
- `Security`: 安全相关修复
