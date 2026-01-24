package main

import (
	"fmt"
	"log"

	"github.com/yourusername/MemoryOs/internal/config"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	cfg, err := config.Load("config/config.yaml")
	if err != nil {
		log.Fatal("加载配置失败:", err)
	}

	fmt.Printf("尝试连接数据库:\n")
	fmt.Printf("  Host: %s\n", cfg.Database.Postgres.Host)
	fmt.Printf("  Port: %d\n", cfg.Database.Postgres.Port)
	fmt.Printf("  User: %s\n", cfg.Database.Postgres.User)
	fmt.Printf("  DB: %s\n", cfg.Database.Postgres.DBName)
	fmt.Printf("  DSN: %s\n", cfg.Database.Postgres.DSN())

	db, err := gorm.Open(postgres.Open(cfg.Database.Postgres.DSN()), &gorm.Config{})
	if err != nil {
		log.Fatal("❌ 连接失败:", err)
	}

	sqlDB, _ := db.DB()
	if err := sqlDB.Ping(); err != nil {
		log.Fatal("❌ Ping 失败:", err)
	}

	fmt.Println("✅ 数据库连接成功!")

	// 测试查询
	var count int64
	db.Raw("SELECT COUNT(*) FROM dialogue_memory").Scan(&count)
	fmt.Printf("✅ dialogue_memory 表中有 %d 条记录\n", count)
}
