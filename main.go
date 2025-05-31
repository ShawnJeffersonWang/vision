package main

import (
	"agricultural_vision/pkg/jwt"
	"context"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"agricultural_vision/controller"
	"agricultural_vision/dao/mysql"
	"agricultural_vision/dao/redis"
	"agricultural_vision/logger"
	"agricultural_vision/routers"
	"agricultural_vision/settings"
	"agricultural_vision/utils"
)

func main() {
	// 初始化所有组件
	if err := setupApp(); err != nil {
		log.Fatal("Failed to setup app:", err)
	}

	// 启动服务器
	runServer()
}

// setupApp 初始化所有组件
func setupApp() error {
	// 初始化配置
	if err := settings.Init(); err != nil {
		return fmt.Errorf("init settings failed: %w", err)
	}

	// 初始化日志（日志初始化后，后续的错误可以使用日志记录）
	if err := logger.Init(settings.Conf.LogConfig, settings.Conf.Mode); err != nil {
		return fmt.Errorf("init logger failed: %w", err)
	}

	// 初始化数据库
	if err := mysql.Init(settings.Conf.MySQLConfig); err != nil {
		return fmt.Errorf("init mysql failed: %w", err)
	}

	// 建表
	if err := utils.InitSqlTable(); err != nil {
		return fmt.Errorf("init sql table failed: %w", err)
	}

	// 初始化Redis
	if err := redis.Init(settings.Conf.RedisConfig); err != nil {
		return fmt.Errorf("init redis failed: %w", err)
	}

	// 初始化JWT
	if err := jwt.Init(); err != nil {
		return fmt.Errorf("init jwt failed: %w", err)
	}

	// 初始化校验器翻译器
	if err := controller.InitTrans("zh"); err != nil {
		return fmt.Errorf("init validator trans failed: %w", err)
	}

	return nil
}

// runServer 启动HTTP服务器并处理优雅关闭
func runServer() {
	r := routers.SetupRouter(settings.Conf.Mode)
	r.Static("/static", "./static")

	srv := &http.Server{
		Addr:    fmt.Sprintf(":%d", settings.Conf.Port),
		Handler: r,
	}

	// 在goroutine中启动服务器
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Error("listen failed", err)
		}
	}()

	logger.Infof("Server started on port %d in %s mode", settings.Conf.Port, settings.Conf.Mode)

	// 等待中断信号
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	logger.Info("Shutting down server...")

	// 优雅关闭
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := srv.Shutdown(ctx); err != nil {
		logger.Error("Server forced to shutdown:", err)
	}

	// 清理资源
	cleanup()

	logger.Info("Server exited")
}

// cleanup 清理资源
func cleanup() {
	mysql.Close()
	redis.Close()
	// 其他需要清理的资源
}

//func main() {
//	// 初始化加载配置文件
//	if err := settings.Init(); err != nil {
//		fmt.Printf("load config failed, err:%v\n", err)
//		return
//	}
//
//	//初始化zap日志库
//	if err := logger.Init(settings.Conf.LogConfig, settings.Conf.Mode); err != nil {
//		fmt.Printf("init logger failed, err:%v\n", err)
//		return
//	}
//
//	//初始化mysql
//	if err := mysql.Init(settings.Conf.MySQLConfig); err != nil {
//		fmt.Printf("init mysql failed, err:%v\n", err)
//		return
//	}
//	defer mysql.Close() // 程序退出关闭数据库连接
//
//	// 建表
//	if err := utils.InitSqlTable(); err != nil {
//		fmt.Printf("init sql table failed, err:%v\n", err)
//		return
//	}
//
//	//初始化redis
//	if err := redis.Init(settings.Conf.RedisConfig); err != nil {
//		fmt.Printf("init redis failed, err:%v\n", err)
//		return
//	}
//	defer redis.Close()
//
//	// 初始化JWT（新增）
//	if err := jwt.Init(); err != nil {
//		fmt.Printf("init jwt failed, err:%v\n", err)
//		return
//	}
//
//	/*//初始化雪花算法
//	if err := snowflake.Init(settings.Conf.StartTime, settings.Conf.MachineID); err != nil {
//		fmt.Printf("init snowflake failed, err:%v\n", err)
//		return
//	}*/
//
//	//初始化gin框架内置的校验器使用的翻译器
//	if err := controller.InitTrans("zh"); err != nil {
//		fmt.Printf("init validator trans failed, err:%v\n", err)
//		return
//	}
//
//	// 注册路由
//	r := routers.SetupRouter(settings.Conf.Mode)
//	r.Static("/static", "./static")
//	err := r.Run(fmt.Sprintf(":%d", settings.Conf.Port))
//	if err != nil {
//		fmt.Printf("run server failed, err:%v\n", err)
//		return
//	}
//}
