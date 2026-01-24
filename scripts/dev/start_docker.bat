@echo off
chcp 65001 >nul
echo ==========================================
echo 🚀 MemoryOS Docker Compose 启动脚本
echo ==========================================
echo.

REM 检查 Docker 是否运行
docker info >nul 2>&1
if %errorlevel% neq 0 (
    echo ❌ Docker 未运行，请先启动 Docker Desktop
    pause
    exit /b 1
)

echo ✅ Docker 运行中
echo.

REM 创建数据目录（如果不存在）
echo 📁 创建数据目录...
if not exist "data\postgres" mkdir "data\postgres"
if not exist "data\redis" mkdir "data\redis"
if not exist "data\milvus" mkdir "data\milvus"
if not exist "data\etcd" mkdir "data\etcd"
if not exist "data\minio" mkdir "data\minio"
if not exist "logs" mkdir "logs"
echo ✅ 数据目录创建完成
echo.

REM 启动所有服务
echo 🔄 启动 Docker Compose 服务...
echo 这可能需要几分钟（首次启动需下载镜像）
echo.
docker compose up -d

if %errorlevel% neq 0 (
    echo ❌ 启动失败，请查看错误信息
    pause
    exit /b 1
)

echo.
echo ==========================================
echo 🎉 启动成功！
echo ==========================================
echo.
echo 📊 服务状态:
docker compose ps
echo.
echo 🌐 访问地址:
echo   - MemoryOS API:    http://localhost:8080
echo   - Health Check:    http://localhost:8080/health
echo   - MinIO 控制台:    http://localhost:9001 (minioadmin/minioadmin)
echo   - Milvus Metrics:  http://localhost:9091
echo.
echo 📝 常用命令:
echo   - 查看日志:        docker compose logs -f
echo   - 停止服务:        docker compose down
echo   - 重启服务:        docker compose restart
echo   - 查看状态:        docker compose ps
echo.
echo 💡 提示: 首次启动 Milvus 可能需要 1-2 分钟初始化
echo          使用 'docker compose logs -f milvus' 查看进度
echo.

pause
