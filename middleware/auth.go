package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	"agricultural_vision/constants"
	"agricultural_vision/pkg/jwt"
)

// 基于JWT的认证中间件，对请求头中的token进行校验，并将用户id放在请求的上下文上
func JWTAuthMiddleware() func(c *gin.Context) {
	return func(c *gin.Context) {
		// 客户端携带Token有三种方式 1.放在请求头 2.放在请求体 3.放在URL
		// 这里Token放在请求头Header的Authorization中，并使用Bearer开头
		// 格式：Authorization: Bearer xxx.xxx.xxx
		// 这里的具体实现方式要依据你的实际业务情况决定
		authHeader := c.Request.Header.Get("Authorization")
		//如果未携带令牌，代表没有登录
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, constants.CodeNeedLogin)
			c.Abort()
			return
		}

		// 按空格分割
		parts := strings.SplitN(authHeader, " ", 2)
		if !(len(parts) == 2 && parts[0] == "Bearer") {
			c.JSON(http.StatusUnauthorized, constants.CodeInvalidToken)
			c.Abort()
			return
		}

		// parts[1]是获取到的tokenString，我们使用之前定义好的解析JWT的函数来解析它
		mc, err := jwt.ParseToken(parts[1])
		if err != nil {
			c.JSON(http.StatusUnauthorized, constants.CodeInvalidToken)
			c.Abort()
			return
		}

		// 将当前请求的userID信息保存到请求的上下文c上
		c.Set("userID", mc.UserID)
		c.Next() // 后续的处理函数可以用过c.Get(ContextUserIDKey)来获取当前请求的用户信息
	}
}
