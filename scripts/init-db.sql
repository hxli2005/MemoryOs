-- MemoryOS PostgreSQL 初始化脚本
-- 创建 pgvector 扩展

-- 创建 pgvector 扩展（如果不存在）
CREATE EXTENSION IF NOT EXISTS vector;

-- 验证扩展安装
SELECT extname, extversion FROM pg_extension WHERE extname = 'vector';

-- 设置默认搜索路径
SET search_path TO public;

-- 完成提示
DO $$
BEGIN
    RAISE NOTICE 'MemoryOS database initialized successfully!';
    RAISE NOTICE 'pgvector extension version: %', (SELECT extversion FROM pg_extension WHERE extname = 'vector');
END $$;
