@echo off
chcp 65001 >nul
echo ==========================================
echo 🛑 MemoryOS Docker Compose 停止脚本
echo ==========================================
echo.

docker compose down

if %errorlevel% neq 0 (
    echo ❌ 停止失败
    pause
    exit /b 1
)

echo.
echo ✅ 所有服务已停止
echo.
echo 💾 数据已保留在 ./data 目录
echo.
echo 🗑️  如需完全清理（包括数据），运行:
echo     docker compose down -v
echo     rmdir /s /q data
echo.

pause
