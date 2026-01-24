-- ==================== 迁移脚本：添加 memory_uuid 字段 ====================
-- 用途：为三张表添加 UUID 字段，用于与 Memory.ID 关联
-- 执行时机：手动执行
-- 日期：2026-01-22

\echo '开始迁移：添加 memory_uuid 字段...'

-- ==================== 1. 添加字段 ====================

-- 对话记忆表
ALTER TABLE dialogue_memory 
    ADD COLUMN IF NOT EXISTS memory_uuid VARCHAR(50);

-- 主题记忆表
ALTER TABLE topic_memory 
    ADD COLUMN IF NOT EXISTS memory_uuid VARCHAR(50);

-- 用户画像表
ALTER TABLE profile_memory 
    ADD COLUMN IF NOT EXISTS memory_uuid VARCHAR(50);

\echo '✅ 字段添加成功'

-- ==================== 2. 创建唯一索引 ====================

-- 对话记忆 UUID 索引
CREATE UNIQUE INDEX IF NOT EXISTS idx_dialogue_uuid 
    ON dialogue_memory(memory_uuid);

-- 主题记忆 UUID 索引
CREATE UNIQUE INDEX IF NOT EXISTS idx_topic_uuid 
    ON topic_memory(memory_uuid);

-- 用户画像 UUID 索引
CREATE UNIQUE INDEX IF NOT EXISTS idx_profile_uuid 
    ON profile_memory(memory_uuid);

\echo '✅ UUID 索引创建成功'

-- ==================== 3. 为现有数据生成 UUID ====================

-- 为对话记忆表现有记录生成 UUID
UPDATE dialogue_memory 
SET memory_uuid = 'dial_' || id::text 
WHERE memory_uuid IS NULL;

-- 为主题记忆表现有记录生成 UUID
UPDATE topic_memory 
SET memory_uuid = 'topic_' || id::text 
WHERE memory_uuid IS NULL;

-- 为用户画像表现有记录生成 UUID
UPDATE profile_memory 
SET memory_uuid = 'prof_' || id::text 
WHERE memory_uuid IS NULL;

\echo '✅ 现有数据 UUID 生成成功'

-- ==================== 4. 添加非空约束 ====================

-- 现在所有记录都有 UUID，可以添加非空约束
ALTER TABLE dialogue_memory 
    ALTER COLUMN memory_uuid SET NOT NULL;

ALTER TABLE topic_memory 
    ALTER COLUMN memory_uuid SET NOT NULL;

ALTER TABLE profile_memory 
    ALTER COLUMN memory_uuid SET NOT NULL;

\echo '✅ 非空约束添加成功'

-- ==================== 迁移完成 ====================

\echo ''
\echo '=========================================='
\echo '🎉 迁移完成！'
\echo '=========================================='
\echo '变更内容:'
\echo '  - 添加 memory_uuid 字段（VARCHAR(50)）'
\echo '  - 创建唯一索引'
\echo '  - 为现有数据生成 UUID'
\echo '  - 添加非空约束'
\echo '=========================================='
