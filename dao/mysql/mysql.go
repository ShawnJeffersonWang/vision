package mysql

import (
	"fmt"
	"log"
	"os"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"gorm.io/gorm/schema"

	"agricultural_vision/settings"
)

var DB *gorm.DB

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
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
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

// Close 关闭 MySQL 连接
func Close() {
	if DB != nil {
		sqlDB, err := DB.DB()
		if err == nil {
			_ = sqlDB.Close()
		}
	}
}
