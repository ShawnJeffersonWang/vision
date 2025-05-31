package main

import (
	"fmt"
	"math/rand"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"
)

// 日志颜色
const (
	colorReset  = "\033[0m"
	colorRed    = "\033[31m"
	colorGreen  = "\033[32m"
	colorYellow = "\033[33m"
	colorBlue   = "\033[34m"
	colorPurple = "\033[35m"
	colorCyan   = "\033[36m"
	colorWhite  = "\033[37m"
)

// 模拟的API路径
var apiPaths = []string{
	"/api/users",
	"/api/users/{id}",
	"/api/products",
	"/api/orders",
	"/api/auth/login",
	"/api/auth/logout",
	"/api/dashboard/stats",
}

// 模拟的SQL查询
var sqlQueries = []string{
	"SELECT * FROM users WHERE id = ?",
	"SELECT id, name, email FROM users WHERE status = 'active'",
	"INSERT INTO orders (user_id, product_id, amount) VALUES (?, ?, ?)",
	"UPDATE users SET last_login = NOW() WHERE id = ?",
	"SELECT COUNT(*) FROM products WHERE category_id = ?",
	"DELETE FROM sessions WHERE expired_at < NOW()",
}

// HTTP方法
var httpMethods = []string{"GET", "POST", "PUT", "DELETE"}

// Logger 结构体
type Logger struct {
	mu sync.Mutex
}

func (l *Logger) Info(format string, args ...interface{}) {
	l.log("INFO", colorGreen, format, args...)
}

func (l *Logger) Error(format string, args ...interface{}) {
	l.log("ERROR", colorRed, format, args...)
}

func (l *Logger) Slow(format string, args ...interface{}) {
	l.log("SLOW", colorYellow, format, args...)
}

func (l *Logger) Stat(format string, args ...interface{}) {
	l.log("STAT", colorBlue, format, args...)
}

func (l *Logger) log(level, color, format string, args ...interface{}) {
	l.mu.Lock()
	defer l.mu.Unlock()

	timestamp := time.Now().Format("2006-01-02 15:04:05.000")
	message := fmt.Sprintf(format, args...)
	fmt.Printf("%s%s [%s]%s %s\n", color, timestamp, level, colorReset, message)
}

var logger = &Logger{}

// 模拟HTTP请求处理
func simulateHTTPRequest() {
	path := apiPaths[rand.Intn(len(apiPaths))]
	method := httpMethods[rand.Intn(len(httpMethods))]
	duration := time.Duration(rand.Intn(300)+50) * time.Millisecond
	statusCode := 200

	// 10%的概率返回错误
	if rand.Float32() < 0.1 {
		statusCodes := []int{400, 404, 500, 503}
		statusCode = statusCodes[rand.Intn(len(statusCodes))]
	}

	// 开始请求
	startTime := time.Now()
	requestID := fmt.Sprintf("%d", rand.Int63())

	logger.Info("[HTTP] %s %s - started [requestId: %s]", method, path, requestID)

	// 模拟处理时间
	time.Sleep(duration)

	// 记录请求完成
	elapsed := time.Since(startTime)
	if elapsed > 200*time.Millisecond {
		logger.Slow("[HTTP] %s %s - %dms - status: %d [requestId: %s]",
			method, path, elapsed.Milliseconds(), statusCode, requestID)
	} else {
		logger.Info("[HTTP] %s %s - %dms - status: %d [requestId: %s]",
			method, path, elapsed.Milliseconds(), statusCode, requestID)
	}

	// 如果是错误状态码，记录错误
	if statusCode >= 400 {
		logger.Error("[HTTP] %s %s failed with status %d", method, path, statusCode)
	}
}

// 模拟SQL查询
func simulateSQLQuery() {
	query := sqlQueries[rand.Intn(len(sqlQueries))]
	duration := time.Duration(rand.Intn(100)+10) * time.Millisecond

	startTime := time.Now()

	// 模拟查询执行
	time.Sleep(duration)

	elapsed := time.Since(startTime)
	rows := rand.Intn(100)

	if elapsed > 50*time.Millisecond {
		logger.Slow("[SQL] exec: %s - duration: %dms, rows: %d",
			query, elapsed.Milliseconds(), rows)
	} else {
		logger.Info("[SQL] exec: %s - duration: %dms, rows: %d",
			query, elapsed.Milliseconds(), rows)
	}
}

// 模拟缓存操作
func simulateCacheOperation() {
	operations := []string{"GET", "SET", "DEL"}
	operation := operations[rand.Intn(len(operations))]
	key := fmt.Sprintf("cache:user:%d", rand.Intn(1000))

	hit := rand.Float32() < 0.7 // 70%命中率

	if operation == "GET" {
		if hit {
			logger.Info("[CACHE] %s %s - HIT", operation, key)
		} else {
			logger.Info("[CACHE] %s %s - MISS", operation, key)
		}
	} else {
		logger.Info("[CACHE] %s %s - OK", operation, key)
	}
}

// 定期输出统计信息
func printStats(requestCount, sqlCount, cacheCount *int64) {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		logger.Stat("Server Stats - HTTP requests: %d, SQL queries: %d, Cache operations: %d",
			*requestCount, *sqlCount, *cacheCount)
	}
}

// 模拟服务健康检查
func healthCheck() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for range ticker.C {
		logger.Info("[HEALTH] Server is healthy - CPU: %.1f%%, Memory: %.1f%%, Goroutines: %d",
			rand.Float32()*30+10, rand.Float32()*40+20, rand.Intn(50)+10)
	}
}

func main() {
	logger.Info("Starting server on :8080")
	logger.Info("Environment: production")
	logger.Info("Version: 1.0.0")

	// 计数器
	var requestCount, sqlCount, cacheCount int64

	// 启动统计协程
	go printStats(&requestCount, &sqlCount, &cacheCount)
	go healthCheck()

	// 设置信号处理
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)

	// 主循环
	ticker := time.NewTicker(time.Millisecond * 100)
	defer ticker.Stop()

	logger.Info("Server started successfully")

	for {
		select {
		case <-sigChan:
			logger.Info("Shutting down server...")
			time.Sleep(1 * time.Second)
			logger.Info("Server stopped")
			return

		case <-ticker.C:
			// 随机执行不同类型的操作
			switch rand.Intn(10) {
			case 0, 1, 2, 3: // 40% HTTP请求
				go func() {
					simulateHTTPRequest()
					requestCount++
				}()
			case 4, 5, 6: // 30% SQL查询
				go func() {
					simulateSQLQuery()
					sqlCount++
				}()
			case 7, 8: // 20% 缓存操作
				go func() {
					simulateCacheOperation()
					cacheCount++
				}()
			case 9: // 10% 同时多个操作
				go func() {
					simulateHTTPRequest()
					simulateSQLQuery()
					simulateCacheOperation()
					requestCount++
					sqlCount++
					cacheCount++
				}()
			}
		}
	}
}
