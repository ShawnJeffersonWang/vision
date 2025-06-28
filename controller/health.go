package controller

import (
	"agricultural_vision/dao/postgres"
	"context"
	"net/http"
	"time"

	"agricultural_vision/dao/redis"
	"github.com/gin-gonic/gin"
)

// HealthCheckHandler 健康检查处理函数
func HealthCheckHandler(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	health := gin.H{
		"status":    "healthy",
		"service":   "agricultural_vision",
		"timestamp": time.Now().Unix(),
		"checks":    gin.H{},
	}

	isHealthy := true

	// 检查 MySQL 连接
	mysqlHealth := checkMySQL(ctx)
	health["checks"].(gin.H)["mysql"] = mysqlHealth
	if mysqlHealth["status"] != "healthy" {
		isHealthy = false
	}

	// 检查 Redis 连接
	redisHealth := checkRedis()
	health["checks"].(gin.H)["redis"] = redisHealth
	if redisHealth["status"] != "healthy" {
		isHealthy = false
	}

	// 设置总体健康状态
	if !isHealthy {
		health["status"] = "unhealthy"
		c.JSON(http.StatusServiceUnavailable, health)
		return
	}

	c.JSON(http.StatusOK, health)
}

// checkMySQL 检查 MySQL 连接
func checkMySQL(ctx context.Context) gin.H {
	if postgres.DB == nil {
		return gin.H{
			"status": "unhealthy",
			"error":  "database not initialized",
		}
	}

	// 获取底层的 *sql.DB
	sqlDB, err := postgres.DB.DB()
	if err != nil {
		return gin.H{
			"status": "unhealthy",
			"error":  err.Error(),
		}
	}

	// 执行 Ping 检查
	if err := sqlDB.PingContext(ctx); err != nil {
		return gin.H{
			"status": "unhealthy",
			"error":  err.Error(),
		}
	}

	// 获取连接池统计信息
	stats := sqlDB.Stats()
	return gin.H{
		"status": "healthy",
		"stats": gin.H{
			"open_connections":    stats.OpenConnections,
			"in_use":              stats.InUse,
			"idle":                stats.Idle,
			"wait_count":          stats.WaitCount,
			"wait_duration":       stats.WaitDuration.String(),
			"max_idle_closed":     stats.MaxIdleClosed,
			"max_lifetime_closed": stats.MaxLifetimeClosed,
		},
	}
}

// checkRedis 检查 Redis 连接
func checkRedis() gin.H {
	// 使用 redis 包中的 Get 函数来测试连接
	// 尝试获取一个不存在的键，这会触发连接检查
	_, err := redis.Get("health_check_test_key")

	// 如果错误是 redis.Nil，说明连接正常，只是键不存在
	if err != nil && err.Error() != "redis: nil" {
		return gin.H{
			"status": "unhealthy",
			"error":  err.Error(),
		}
	}

	// 或者我们可以设置并获取一个测试值
	testKey := "health_check_ping"
	testValue := time.Now().Unix()

	// 设置一个短期过期的测试键
	if err := redis.Set(testKey, testValue, 10*time.Second); err != nil {
		return gin.H{
			"status": "unhealthy",
			"error":  "failed to set test key: " + err.Error(),
		}
	}

	// 读取测试键
	if _, err := redis.Get(testKey); err != nil {
		return gin.H{
			"status": "unhealthy",
			"error":  "failed to get test key: " + err.Error(),
		}
	}

	// 删除测试键
	_ = redis.Del(testKey)

	return gin.H{
		"status": "healthy",
	}
}
