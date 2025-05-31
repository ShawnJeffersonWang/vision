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
	// 构造 DSN (Comments Source Email)
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		cfg.User, cfg.Password, cfg.Host, cfg.Port, cfg.DB)

	// 自定义日志配置
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // 日志输出目标
		logger.Config{
			SlowThreshold:             time.Second, // 慢 SQL 阈值
			LogLevel:                  logger.Info, // 日志级别
			IgnoreRecordNotFoundError: true,        // 忽略记录未找到错误
			Colorful:                  true,        // 彩色打印
		},
	)

	// 配置 GORM 选项
	DB, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		NamingStrategy: schema.NamingStrategy{
			SingularTable: true, //是否单数表名
		},
		Logger: newLogger,
	})
	if err != nil {
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
		//获取底层的 SQL 数据库连接，并检查是否有错误
		sqlDB, err := DB.DB()
		//如果没错误
		if err == nil {
			_ = sqlDB.Close()
		}
	}
}
