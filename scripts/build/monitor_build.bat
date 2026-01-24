@echo off
chcp 65001 >nul
echo ==========================================
echo 📊 Docker 构建进度监控
echo ==========================================
echo.

:loop
cls
echo ==========================================
echo 📊 Docker 构建进度监控
echo ==========================================
echo 当前时间: %date% %time%
echo.

echo [1/3] 检查构建进程...
docker ps -a --filter "name=memoryos-app" --format "table {{.Names}}\t{{.Status}}" 2>nul

echo.
echo [2/3] 检查镜像...
docker images memoryos-memoryos 2>nul

echo.
echo [3/3] 服务状态...
docker compose ps 2>nul

echo.
echo ==========================================
echo 💡 提示:
echo   - 按 Ctrl+C 停止监控
echo   - 构建通常需要 10-15 分钟
echo   - 查看详细日志: docker compose logs -f
echo ==========================================
echo.

timeout /t 30 /nobreak >nul
goto loop
