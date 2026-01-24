@echo off
chcp 65001 >nul
echo ==========================================
echo 🔍 MemoryOS Docker 环境测试
echo ==========================================
echo.

echo [1/5] 测试 PostgreSQL 连接...
docker exec memoryos-postgres psql -U memoryos -d memoryos -c "SELECT 'PostgreSQL 连接成功' AS status;" 2>nul
if %errorlevel% neq 0 (
    echo ❌ PostgreSQL 连接失败
) else (
    echo ✅ PostgreSQL 连接成功
)
echo.

echo [2/5] 测试 pgvector 扩展...
docker exec memoryos-postgres psql -U memoryos -d memoryos -c "SELECT extname FROM pg_extension WHERE extname = 'vector';" 2>nul
if %errorlevel% neq 0 (
    echo ❌ pgvector 扩展未安装
) else (
    echo ✅ pgvector 扩展已安装
)
echo.

echo [3/5] 测试 Redis 连接...
docker exec memoryos-redis redis-cli PING 2>nul
if %errorlevel% neq 0 (
    echo ❌ Redis 连接失败
) else (
    echo ✅ Redis 连接成功
)
echo.

echo [4/5] 测试 Milvus 连接...
curl -s http://localhost:9091/healthz >nul 2>&1
if %errorlevel% neq 0 (
    echo ⏳ Milvus 可能还在启动中...
) else (
    echo ✅ Milvus 运行正常
)
echo.

echo [5/5] 测试 MemoryOS API...
curl -s http://localhost:8080/health >nul 2>&1
if %errorlevel% neq 0 (
    echo ⏳ MemoryOS 可能还在启动中...
    echo    使用 'docker compose logs -f memoryos' 查看日志
) else (
    echo ✅ MemoryOS API 运行正常
    echo.
    echo 📊 完整响应:
    curl -s http://localhost:8080/health
)
echo.

echo ==========================================
echo 测试完成！
echo ==========================================
echo.

pause
