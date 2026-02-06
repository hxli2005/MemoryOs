# MemoryOS Docker 快速管理脚本 (PowerShell)

Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "MemoryOS Docker 管理工具" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "请选择操作:" -ForegroundColor Yellow
Write-Host "1. 启动所有服务"
Write-Host "2. 停止所有服务"
Write-Host "3. 重启应用"
Write-Host "4. 查看日志"
Write-Host "5. 查看服务状态"
Write-Host "6. 进入数据库"
Write-Host "7. 测试环境"
Write-Host "8. 完全清理（包括数据）"
Write-Host "0. 退出"
Write-Host ""

$choice = Read-Host "请输入选项 (0-8)"

switch ($choice) {
    "1" {
        Write-Host "`n启动所有服务..." -ForegroundColor Green
        docker compose -f deploy/compose/docker-compose.yml up -d
    }
    "2" {
        Write-Host "`n停止所有服务..." -ForegroundColor Yellow
        docker compose -f deploy/compose/docker-compose.yml down
    }
    "3" {
        Write-Host "`n重启应用..." -ForegroundColor Green
        docker compose -f deploy/compose/docker-compose.yml restart memoryos
    }
    "4" {
        Write-Host "`n查看日志（按 Ctrl+C 退出）..." -ForegroundColor Cyan
        docker compose -f deploy/compose/docker-compose.yml logs -f memoryos
    }
    "5" {
        Write-Host "`n服务状态:" -ForegroundColor Cyan
        docker compose -f deploy/compose/docker-compose.yml ps
    }
    "6" {
        Write-Host "`n进入 PostgreSQL（输入 \q 退出）..." -ForegroundColor Cyan
        docker exec -it memoryos-postgres psql -U memoryos -d memoryos
    }
    "7" {
        Write-Host "`n运行环境测试..." -ForegroundColor Green
        & ".\scripts\test\test_docker.bat"
    }
    "8" {
        $confirm = Read-Host "`n⚠️  确认删除所有数据? (y/N)"
        if ($confirm -eq "y" -or $confirm -eq "Y") {
            docker compose -f deploy/compose/docker-compose.yml down -v
            Remove-Item -Recurse -Force ".\data" -ErrorAction SilentlyContinue
            Write-Host "✅ 已清理所有数据" -ForegroundColor Green
        } else {
            Write-Host "取消操作" -ForegroundColor Yellow
        }
    }
    "0" {
        Write-Host "退出" -ForegroundColor Gray
        exit 0
    }
    default {
        Write-Host "无效选项" -ForegroundColor Red
    }
}
