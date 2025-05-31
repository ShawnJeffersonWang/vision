package middleware

import (
	"github.com/gin-gonic/gin"

	"agricultural_vision/constants"
)

// 从请求上下文中获取当前用户ID
func GetCurrentUserID(c *gin.Context) (int64, error) {
	// 如果用户已登录则可以在请求上下文中获取到userID, 如果获取不到则用户未登录
	uid, ok := c.Get("userID")
	// 如果获取不到则返回错误
	if !ok {
		return 0, constants.ErrorNeedLogin
	}

	// 类型断言失败则直接返回错误
	userID, _ := uid.(int64)

	// 没有错误，直接返回
	return userID, nil
}
