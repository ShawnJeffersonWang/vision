package middleware

import (
	"agricultural_vision/dao/mysql"
	"go.uber.org/zap"
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

// AdminAuthMiddleware 管理员权限中间件
func AdminAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 获取当前用户角色
		role, exists := c.Get(constants.CtxUserRoleKey)
		if !exists {
			// 如果context中没有角色信息，尝试从数据库获取
			userID, exists := c.Get(constants.CtxUserIDKey)
			if !exists {
				c.JSON(http.StatusUnauthorized, constants.CodeNeedLogin)
				c.Abort()
				return
			}

			uid, ok := userID.(int64)
			if !ok {
				c.JSON(http.StatusUnauthorized, constants.CodeNeedLogin)
				c.Abort()
				return
			}

			// 从数据库获取用户信息
			user, err := mysql.GetUserByID(uid)
			if err != nil {
				zap.L().Error("获取用户信息失败",
					zap.Int64("user_id", uid),
					zap.Error(err))
				c.JSON(http.StatusInternalServerError, constants.CodeServerBusy)
				c.Abort()
				return
			}

			role = user.Role
			c.Set(constants.CtxUserRoleKey, role)
		}

		// 检查是否是管理员
		roleStr, ok := role.(string)
		if !ok || roleStr != constants.RoleAdmin {
			c.JSON(http.StatusForbidden, constants.CodeNoPermission)
			c.Abort()
			return
		}

		c.Next()
	}
}

// PermissionMiddleware 权限中间件（更灵活的权限控制）
func PermissionMiddleware(requiredRoles ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role, exists := c.Get(constants.CtxUserRoleKey)
		if !exists {
			c.JSON(http.StatusUnauthorized, constants.CodeNeedLogin)
			c.Abort()
			return
		}

		roleStr, ok := role.(string)
		if !ok {
			c.JSON(http.StatusUnauthorized, constants.CodeNeedLogin)
			c.Abort()
			return
		}

		// 检查角色是否在允许的角色列表中
		allowed := false
		for _, allowedRole := range requiredRoles {
			if roleStr == allowedRole {
				allowed = true
				break
			}
		}

		if !allowed {
			c.JSON(http.StatusForbidden, constants.CodeNoPermission)
			c.Abort()
			return
		}

		c.Next()
	}
}
