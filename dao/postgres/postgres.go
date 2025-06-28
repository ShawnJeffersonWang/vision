package postgres // 包名从 mysql 修改为 postgres

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/postgres" // 关键：导入 PostgreSQL 驱动
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"agricultural_vision/settings" // 假设您的 settings 包路径
)

var DB *gorm.DB

// Init 初始化 PostgreSQL 连接
func Init(cfg *settings.PostgreSQLConfig) (err error) {
	// 打印配置，调试用
	//log.Printf("PostgreSQL Config: host=%s, port=%d, user=%s, dbname=%s, timezone=%s",
	//	cfg.Host, cfg.Port, cfg.User, cfg.DBName, cfg.TimeZone)

	// 构造 PostgreSQL 的 DSN
	// 格式: "host=localhost user=gorm password=gorm dbname=gorm port=9920 sslmode=disable TimeZone=Asia/Shanghai"
	dsn := fmt.Sprintf("host=%s user=%s password=%s dbname=%s port=%d",
		cfg.Host, cfg.User, cfg.Password, cfg.DBName, cfg.Port)

	// 自定义日志配置 (这部分无需修改)
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags),
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Info,
			IgnoreRecordNotFoundError: true,
			Colorful:                  true,
		},
	)

	// 配置 GORM 选项
	// 关键：使用 postgres.Open()
	DB, err = gorm.Open(postgres.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		Logger: newLogger,
	})
	if err != nil {
		log.Printf("Failed to connect to PostgreSQL with DSN: %s, error: %v", dsn, err)
		return
	}

	// 获取通用数据库连接池对象并设置连接池配置 (这部分无需修改)
	sqlDB, err := DB.DB()
	if err != nil {
		return
	}

	// 设置最大连接数
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	// 设置最大空闲连接数
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)

	return
}

// Close 关闭 PostgreSQL 连接 (函数体无需修改)
func Close() {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	}
}

// Ping 检查数据库连接 (函数体无需修改)
func Ping(ctx context.Context) error {
	if DB == nil {
		return fmt.Errorf("database not initialized")
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	return sqlDB.PingContext(ctx)
}

// GetStats 获取数据库连接池统计信息 (函数体无需修改)
func GetStats() (map[string]interface{}, error) {
	if DB == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	sqlDB, err := DB.DB()
	if err != nil {
		return nil, err
	}

	stats := sqlDB.Stats()
	return map[string]interface{}{
		"open_connections":    stats.OpenConnections,
		"in_use":              stats.InUse,
		"idle":                stats.Idle,
		"wait_count":          stats.WaitCount,
		"wait_duration":       stats.WaitDuration.String(),
		"max_idle_closed":     stats.MaxIdleClosed,
		"max_lifetime_closed": stats.MaxLifetimeClosed,
	}, nil
}
