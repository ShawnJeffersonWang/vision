package routers

import (
	"net/http"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"

	"agricultural_vision/controller"
	"agricultural_vision/logger"
	"agricultural_vision/middleware"
)

func SetupRouter(mode string) *gin.Engine {
	// 如果当前代码是运行模式，则将gin设置成发布模式
	if mode == gin.ReleaseMode {
		gin.SetMode(gin.ReleaseMode) // gin设置成发布模式
	}

	r := gin.New()
	r.Use(logger.GinLogger(), logger.GinRecovery(true))
	r.Use(cors.New(cors.Config{
		// 允许的域名（前端地址）
		AllowOrigins: []string{"*"}, // 允许所有源
		// 允许的请求方法
		AllowMethods: []string{"GET", "POST", "PUT", "DELETE", "OPTIONS"},
		// 允许的请求头
		AllowHeaders: []string{"Content-Type", "Authorization"},
		// 允许携带认证信息
		AllowCredentials: true,
	}))

	// 用户模块
	userGroup := r.Group("/user")
	{
		// 用户注册
		userGroup.POST("/signup", controller.SignUpHandler)
		// 用户登录
		userGroup.POST("/login", controller.LoginHandler)
		// 发送邮箱验证码
		userGroup.POST("/email", controller.VerifyEmailHandler)
		// 修改密码
		userGroup.POST("/change-password", controller.ChangePasswordHandler)

		// jwt校验
		userGroup.Use(middleware.JWTAuthMiddleware())
		{
			// 查询个人信息
			userGroup.GET("/info", controller.GetUserInfoHandler)
			// 修改个人信息
			userGroup.PUT("/info", controller.UpdateUserInfoHandler)
			// 修改个人头像
			userGroup.POST("/avatar", controller.UpdateUserAvatarHandler)
			// 查询用户主页
			userGroup.GET("/home-page/:id", controller.GetUserHomePageHandler)
		}
	}

	// AI模块
	AIGroup := r.Group("/ai")
	{
		AIGroup.POST("", middleware.JWTAuthMiddleware(), controller.AiHandler)
	}

	// 搜索模块
	searchGroup := r.Group("/search")
	{
		searchGroup.GET("", controller.SearchHandler)
		searchGroup.GET("/:crop_id", controller.SearchCropHandler)
	}

	// 首页模块
	firstPageGroup := r.Group("/firstpage")
	{
		firstPageGroup.GET("/news", controller.GetNewsHandler)
		firstPageGroup.GET("/proverb", controller.GetProverbHandler)
		firstPageGroup.GET("/crop", controller.GetCropHandler)
		firstPageGroup.GET("/video", controller.GetVideoHandler)
		firstPageGroup.GET("/poetry", controller.GetPoetryHandler)
	}

	// 社区帖子模块
	communityPost := r.Group("/community-post")
	{
		/*公开接口，不需要登录*/
		// 查询帖子列表（指定排序方式）（游客登录）
		communityPost.GET("/posts/guest", controller.GetPostListHandler)
		// 查询帖子列表（指定社区）（指定排序方式，默认按时间倒序）（游客登录）
		communityPost.GET("/community/:id/posts/guest", controller.GetCommunityPostListHandler)
		// 查询社区列表
		communityPost.GET("/community", controller.CommunityHandler)
		// 查询社区详情
		communityPost.GET("/community/:id", controller.CommunityDetailHandler)

		/*需要登录的接口*/
		authCommunityPost := communityPost.Group("/")
		{

			// 使用jwt校验
			authCommunityPost.Use(middleware.JWTAuthMiddleware())

			// 查询帖子列表（指定排序方式）（用户登录）
			authCommunityPost.GET("/posts", controller.GetPostListHandler)
			// 查询帖子列表（指定社区）（指定排序方式，默认按时间倒序）（用户登录）
			authCommunityPost.GET("/community/:id/posts", controller.GetCommunityPostListHandler)
			// 发布帖子
			authCommunityPost.POST("/post", controller.CreatePostHandler)
			// 上传帖子图片
			authCommunityPost.POST("/upload", controller.UploadPostImageHandler)
			// 删除帖子
			authCommunityPost.DELETE("/post/:id", controller.DeletePostHandler)
			// 帖子投票
			authCommunityPost.POST("/post/vote", controller.PostVoteController)

			// 发布评论
			authCommunityPost.POST("/comment", controller.CreateCommentHandler)
			// 删除评论
			authCommunityPost.DELETE("/comment/:id", controller.DeleteCommentHandler)
			// 查询顶级评论（指定排序方式，默认按时间倒序）
			authCommunityPost.GET("/first-level-comment/:post_id", controller.GetTopCommentListHandler)
			// 查询子评论（按时间正序）
			authCommunityPost.GET("/second-level-comment/:comment_id", controller.GetSonCommentListHandler)
			// 查询帖子所有评论（指定排序方式，默认按时间倒序）
			authCommunityPost.GET("/comment/:post_id", controller.GetCommentListHandler)
			// 评论投票
			authCommunityPost.POST("/comment/vote", controller.CommentVoteController)
		}

		userCommunityPost := authCommunityPost.Group("/user")
		{
			// 查询用户的帖子列表（分页）
			userCommunityPost.GET("/posts", controller.GetUserPostListHandler)
			// 查询用户点赞的帖子列表（分页）
			userCommunityPost.GET("/likes", controller.GetUserLikedPostListHandler)
		}
	}

	r.NoRoute(func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"msg": "404",
		})
	})

	return r
}
