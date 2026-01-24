-- ==================== MemoryOS PostgreSQL 初始化脚本 ====================
-- 用途：自动创建数据库表结构、扩展和索引
-- 执行时机：PostgreSQL 容器首次启动时自动执行

-- ==================== 1. 启用扩展 ====================

-- pgvector: 向量存储和检索
CREATE EXTENSION IF NOT EXISTS vector;

-- uuid-ossp: UUID 生成
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

\echo '✅ 扩展创建成功: vector, uuid-ossp'

-- ==================== 2. 创建记忆表 ====================

-- 对话记忆表 (Dialogue Layer)
CREATE TABLE IF NOT EXISTS dialogue_memory (
    id BIGSERIAL PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    session_id VARCHAR(255),
    content TEXT NOT NULL,
    role VARCHAR(50) DEFAULT 'user',
    embedding vector(768),  -- qwen3-embedding-4b 768维向量
    metadata JSONB,         -- 额外元数据
    memory_type VARCHAR(50) DEFAULT 'dialogue',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 主题记忆表 (Topic Layer)
CREATE TABLE IF NOT EXISTS topic_memory (
    id BIGSERIAL PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL,
    title VARCHAR(500) NOT NULL,
    summary TEXT,
    content TEXT NOT NULL,
    embedding vector(768),
    keywords TEXT[],        -- 关键词数组
    dialogue_ids BIGINT[],  -- 关联的对话ID
    memory_type VARCHAR(50) DEFAULT 'topic',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 用户画像表 (Profile Layer)
CREATE TABLE IF NOT EXISTS profile_memory (
    id BIGSERIAL PRIMARY KEY,
    user_id VARCHAR(255) NOT NULL UNIQUE,
    preferences JSONB,      -- 用户偏好
    habits JSONB,           -- 用户习惯
    features JSONB,         -- 用户特征
    embedding vector(768),
    topic_ids BIGINT[],     -- 关联的主题ID
    memory_type VARCHAR(50) DEFAULT 'profile',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

\echo '✅ 记忆表创建成功: dialogue_memory, topic_memory, profile_memory'

-- ==================== 3. 创建索引 ====================

-- 对话记忆索引
CREATE INDEX IF NOT EXISTS idx_dialogue_user_created 
    ON dialogue_memory(user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_dialogue_session 
    ON dialogue_memory(session_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_dialogue_type 
    ON dialogue_memory(memory_type);

-- 向量索引 (IVF 索引，适合百万级向量)
CREATE INDEX IF NOT EXISTS idx_dialogue_embedding 
    ON dialogue_memory 
    USING ivfflat (embedding vector_cosine_ops)
    WITH (lists = 100);

-- 主题记忆索引
CREATE INDEX IF NOT EXISTS idx_topic_user_created 
    ON topic_memory(user_id, created_at DESC);

CREATE INDEX IF NOT EXISTS idx_topic_keywords 
    ON topic_memory USING GIN(keywords);

CREATE INDEX IF NOT EXISTS idx_topic_embedding 
    ON topic_memory 
    USING ivfflat (embedding vector_cosine_ops)
    WITH (lists = 100);

-- 用户画像索引
CREATE INDEX IF NOT EXISTS idx_profile_user 
    ON profile_memory(user_id);

CREATE INDEX IF NOT EXISTS idx_profile_embedding 
    ON profile_memory 
    USING ivfflat (embedding vector_cosine_ops)
    WITH (lists = 50);

\echo '✅ 索引创建成功: 元数据索引 + 向量索引'

-- ==================== 4. 创建更新时间触发器 ====================

-- 触发器函数：自动更新 updated_at 字段
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- 应用触发器到各表
CREATE TRIGGER update_dialogue_updated_at
    BEFORE UPDATE ON dialogue_memory
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_topic_updated_at
    BEFORE UPDATE ON topic_memory
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

CREATE TRIGGER update_profile_updated_at
    BEFORE UPDATE ON profile_memory
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

\echo '✅ 触发器创建成功: 自动更新 updated_at'

-- ==================== 5. 创建视图（可选）====================

-- 最近对话视图
CREATE OR REPLACE VIEW recent_dialogues AS
SELECT 
    id, 
    user_id, 
    content, 
    role, 
    created_at,
    embedding <=> '[0,0,0]'::vector AS similarity  -- 占位符
FROM dialogue_memory
ORDER BY created_at DESC
LIMIT 1000;

\echo '✅ 视图创建成功: recent_dialogues'

-- ==================== 6. 插入测试数据（开发环境）====================

-- 插入示例对话
INSERT INTO dialogue_memory (user_id, session_id, content, role) VALUES
    ('test_user', 'session_001', '你好，我想了解一下AI记忆系统', 'user'),
    ('test_user', 'session_001', '你好！AI记忆系统是一个基于RAG架构的长期记忆解决方案', 'assistant'),
    ('test_user', 'session_002', '我喜欢使用Go语言开发', 'user');

\echo '✅ 测试数据插入成功'

-- ==================== 7. 权限配置 ====================

-- 确保应用用户有所有表的权限
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO memoryos;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO memoryos;

\echo '✅ 权限配置成功'

-- ==================== 初始化完成 ====================

\echo ''
\echo '=========================================='
\echo '🎉 MemoryOS 数据库初始化完成！'
\echo '=========================================='
\echo '数据库信息:'
\echo '  - 扩展: pgvector, uuid-ossp'
\echo '  - 表数量: 3 (dialogue, topic, profile)'
\echo '  - 索引: 元数据索引 + IVF 向量索引'
\echo '  - 触发器: 自动更新时间戳'
\echo '=========================================='
