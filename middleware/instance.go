package middleware

import (
	"os"
	"sync"

	"github.com/gin-gonic/gin"
)

var (
	instanceID string
	once       sync.Once
)

func getInstanceID() string {
	once.Do(func() {
		instanceID = os.Getenv("INSTANCE_NAME")
		if instanceID == "" {
			instanceID, _ = os.Hostname()
		}
		if instanceID == "" {
			instanceID = "unknown"
		}
	})
	return instanceID
}

// InstanceIDMiddleware 添加实例ID到响应头
func InstanceIDMiddleware() gin.HandlerFunc {
	id := getInstanceID()
	return func(c *gin.Context) {
		// 在响应头中添加实例ID
		c.Header("X-Instance-ID", id)
		// 在上下文中保存，供需要时使用
		c.Set("instanceID", id)
		c.Next()
	}
}
