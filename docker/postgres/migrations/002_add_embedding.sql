-- ==================== 迁移脚本：添加 embedding 向量字段 ====================
-- 用途：为三张表添加向量字段和索引
-- 日期：2026-01-22

\echo '开始迁移：添加 embedding 字段...'

-- ==================== 1. 添加 embedding 字段 ====================

ALTER TABLE dialogue_memory 
    ADD COLUMN IF NOT EXISTS embedding vector(768);

ALTER TABLE topic_memory 
    ADD COLUMN IF NOT EXISTS embedding vector(768);

ALTER TABLE profile_memory 
    ADD COLUMN IF NOT EXISTS embedding vector(768);

\echo '✅ embedding 字段添加成功'

-- ==================== 2. 创建向量索引 ====================

-- IVF 索引适合大规模向量检索
-- lists 参数：聚类中心数量，通常设为 sqrt(记录数)
-- 对于百万级数据，lists=100 是合理的起点

CREATE INDEX IF NOT EXISTS idx_dialogue_embedding 
    ON dialogue_memory 
    USING ivfflat (embedding vector_cosine_ops)
    WITH (lists = 100);

CREATE INDEX IF NOT EXISTS idx_topic_embedding 
    ON topic_memory 
    USING ivfflat (embedding vector_cosine_ops)
    WITH (lists = 100);

CREATE INDEX IF NOT EXISTS idx_profile_embedding 
    ON profile_memory 
    USING ivfflat (embedding vector_cosine_ops)
    WITH (lists = 50);

\echo '✅ 向量索引创建成功'

-- ==================== 迁移完成 ====================

\echo ''
\echo '=========================================='
\echo '🎉 Embedding 字段迁移完成！'
\echo '=========================================='
\echo '变更内容:'
\echo '  - 添加 embedding vector(768) 字段'
\echo '  - 创建 IVF 向量索引（cosine 距离）'
\echo '  - dialogue/topic: lists=100'
\echo '  - profile: lists=50'
\echo '=========================================='
