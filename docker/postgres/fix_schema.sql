-- ==================== MemoryOS 数据库 Schema 修复脚本 ====================
-- 用途：修复 topic_memory 和 profile_memory 表结构，使其与代码模型匹配
-- 执行方式：docker exec -i memoryos-postgres psql -U memoryos -d memoryos < fix_schema.sql

\echo '开始修复数据库 Schema...'

-- ==================== 1. 修复 topic_memory 表 ====================

\echo '修复 topic_memory 表...'

-- 添加 title 字段
ALTER TABLE topic_memory ADD COLUMN IF NOT EXISTS title VARCHAR(500);

-- 添加 summary 字段
ALTER TABLE topic_memory ADD COLUMN IF NOT EXISTS summary TEXT;

-- 添加 keywords 字段（PostgreSQL 数组）
ALTER TABLE topic_memory ADD COLUMN IF NOT EXISTS keywords TEXT[];

-- 添加 dialogue_ids 字段（关联的对话ID）
ALTER TABLE topic_memory ADD COLUMN IF NOT EXISTS dialogue_ids UUID[];

-- 为 title 添加索引
CREATE INDEX IF NOT EXISTS idx_topic_title ON topic_memory(title);

-- 为 keywords 添加 GIN 索引（支持数组搜索）
CREATE INDEX IF NOT EXISTS idx_topic_keywords ON topic_memory USING GIN(keywords);

\echo '✅ topic_memory 表修复完成'

-- ==================== 2. 修复 profile_memory 表 ====================

\echo '修复 profile_memory 表...'

-- 添加 content 字段
ALTER TABLE profile_memory ADD COLUMN IF NOT EXISTS content TEXT NOT NULL DEFAULT '';

-- 添加 preferences 字段
ALTER TABLE profile_memory ADD COLUMN IF NOT EXISTS preferences JSONB;

-- 添加 habits 字段  
ALTER TABLE profile_memory ADD COLUMN IF NOT EXISTS habits JSONB;

-- 添加 features 字段
ALTER TABLE profile_memory ADD COLUMN IF NOT EXISTS features JSONB;

-- 添加 topic_ids 字段
ALTER TABLE profile_memory ADD COLUMN IF NOT EXISTS topic_ids UUID[];

-- 为 JSONB 字段添加 GIN 索引
CREATE INDEX IF NOT EXISTS idx_profile_preferences ON profile_memory USING GIN(preferences);
CREATE INDEX IF NOT EXISTS idx_profile_habits ON profile_memory USING GIN(habits);
CREATE INDEX IF NOT EXISTS idx_profile_features ON profile_memory USING GIN(features);

\echo '✅ profile_memory 表修复完成'

-- ==================== 3. 验证修复结果 ====================

\echo '验证修复结果...'

-- 显示 topic_memory 表结构
\echo '\ntopic_memory 表结构:'
\d topic_memory

-- 显示 profile_memory 表结构
\echo '\nprofile_memory 表结构:'
\d profile_memory

\echo '\n✅ Schema 修复完成！'
\echo '现在可以重新运行测试：go run ./test/test_complete.go'
