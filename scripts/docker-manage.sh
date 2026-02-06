#!/bin/bash
# MemoryOS Docker 快速命令脚本

echo "=========================================="
echo "MemoryOS Docker 管理工具"
echo "=========================================="
echo ""
echo "请选择操作:"
echo "1. 启动所有服务"
echo "2. 停止所有服务"
echo "3. 重启应用"
echo "4. 查看日志"
echo "5. 查看服务状态"
echo "6. 进入数据库"
echo "7. 测试环境"
echo "8. 完全清理（包括数据）"
echo "0. 退出"
echo ""
read -p "请输入选项 (0-8): " choice

case $choice in
    1)
        echo "启动所有服务..."
        docker compose -f deploy/compose/docker-compose.yml up -d
        ;;
    2)
        echo "停止所有服务..."
        docker compose -f deploy/compose/docker-compose.yml down
        ;;
    3)
        echo "重启应用..."
        docker compose -f deploy/compose/docker-compose.yml restart memoryos
        ;;
    4)
        echo "查看日志（按 Ctrl+C 退出）..."
        docker compose -f deploy/compose/docker-compose.yml logs -f memoryos
        ;;
    5)
        echo "服务状态:"
        docker compose -f deploy/compose/docker-compose.yml ps
        ;;
    6)
        echo "进入 PostgreSQL（输入 \q 退出）..."
        docker exec -it memoryos-postgres psql -U memoryos -d memoryos
        ;;
    7)
        echo "运行环境测试..."
        ./scripts/test/test_docker.bat
        ;;
    8)
        read -p "⚠️  确认删除所有数据? (y/N): " confirm
        if [ "$confirm" = "y" ]; then
            docker compose -f deploy/compose/docker-compose.yml down -v
            rm -rf data/
            echo "✅ 已清理所有数据"
        else
            echo "取消操作"
        fi
        ;;
    0)
        echo "退出"
        exit 0
        ;;
    *)
        echo "无效选项"
        ;;
esac
