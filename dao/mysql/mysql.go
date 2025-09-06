package mysql

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"vision/settings"
)

var DBMySQL *gorm.DB

// Init 初始化 MySQL 连接
func Init(cfg *settings.MySQLConfig) (err error) {
	// 打印配置，调试用
	log.Printf("MySQL Config: host=%s, port=%d, user=%s, db=%s",
		cfg.Host, cfg.Port, cfg.User, cfg.DB)
	// 构造 DSN
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DB)

	// 自定义日志配置
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
	DBMySQL, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true,
		},
		Logger: newLogger,
	})
	if err != nil {
		log.Printf("Failed to connect to MySQL with DSN: %s, error: %v", dsn, err)
		return
	}

	// 获取通用数据库连接池对象并设置连接池配置
	sqlDB, err := DBMySQL.DB()
	if err != nil {
		return
	}

	// 设置最大连接数
	sqlDB.SetMaxOpenConns(cfg.MaxOpenConns)
	// 设置最大空闲连接数
	sqlDB.SetMaxIdleConns(cfg.MaxIdleConns)

	return
}

// Close 关闭 MySQL 连接
func Close() {
	if DBMySQL != nil {
		sqlDB, err := DBMySQL.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	}
}

// Ping 检查数据库连接
func Ping(ctx context.Context) error {
	if DBMySQL == nil {
		return fmt.Errorf("database not initialized")
	}

	sqlDB, err := DBMySQL.DB()
	if err != nil {
		return fmt.Errorf("failed to get sql.DB: %w", err)
	}

	return sqlDB.PingContext(ctx)
}

// GetStats 获取数据库连接池统计信息
func GetStats() (map[string]interface{}, error) {
	if DBMySQL == nil {
		return nil, fmt.Errorf("database not initialized")
	}

	sqlDB, err := DBMySQL.DB()
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
