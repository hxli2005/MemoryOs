@echo off
chcp 65001 >nul
echo ========================================
echo   Milvus VectorStore 测试启动
echo ========================================
echo.

echo [1/4] 检查 Docker 状态...
docker info >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ Docker 未运行，请先启动 Docker Desktop
    echo.
    echo    打开 Docker Desktop 后重新运行此脚本
    pause
    exit /b 1
)
echo ✅ Docker 运行正常
echo.

echo [2/4] 启动 Milvus 及依赖服务...
docker-compose up -d etcd minio milvus
if %errorlevel% neq 0 (
    echo ❌ Milvus 启动失败
    pause
    exit /b 1
)
echo ✅ Milvus 启动成功
echo.

echo [3/4] 等待 Milvus 就绪...
timeout /t 10 /nobreak >nul
echo ✅ 等待完成
echo.

echo [4/4] 运行集成测试...
go run ./test/test_milvus.go
echo.

echo ========================================
echo   测试完成！
echo ========================================
echo.
echo 查看结果：
echo   - 如果看到 "✅ Milvus VectorStore 集成成功"，说明一切正常
echo   - 如果看到 "⚠️  VectorStore 可能使用 Mock 模式"，请检查 Milvus 日志
echo.
echo 查看 Milvus 日志:
echo   docker logs memoryos-milvus
echo.
pause
