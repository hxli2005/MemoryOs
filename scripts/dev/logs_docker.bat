@echo off
chcp 65001 >nul
echo ==========================================
echo 📋 MemoryOS Docker 日志查看
echo ==========================================
echo.

REM 检查是否有参数
if "%1"=="" (
    echo 📊 查看所有服务日志...
    echo 按 Ctrl+C 退出
    echo.
    docker compose logs -f
) else (
    echo 📊 查看 %1 服务日志...
    echo 按 Ctrl+C 退出
    echo.
    docker compose logs -f %1
)
