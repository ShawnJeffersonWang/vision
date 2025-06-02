package constants

//封装自定义的状态码和信息

type ResCode int64

const (
	CodeSuccess             string = "success"
	CodeInvalidParam        string = "请求参数错误"
	CodeEmailExist          string = "此邮箱已注册"
	CodeEmailNotExist       string = "此邮箱未注册"
	CodeInvalidPassword     string = "邮箱或密码错误"
	CodeInvalidEmailCode    string = "邮箱验证码错误"
	CodeInvalidCredentials  string = "校验失败"
	CodeNeedLogin           string = "用户需要登录"
	CodeInvalidToken        string = "无效的token"
	CodeAiNotAnswer         string = "AI未回答"
	CodeServerBusy          string = "服务繁忙"
	CodeNotAffectData       string = "未影响到数据"
	CodeNoResult            string = "未查询到结果"
	CodeNoPost              string = "此帖子不存在"
	CodeNoComment           string = "此评论不存在"
	CodeNoPermission        string = "没有权限"
	CodeEmptyKeyword        string = "关键词为空"
	CodeCommunityNameExists string = "社区名称已存在"
	CodeTokenExpired        string = "token已过期"
	CodeNeedRefreshToken    string = "需要提供刷新token"
	CodeUserNotExist        string = "用户不存在"
	CodeErrKafkaNotEnabled  string = "kafka未启动"
	CodeKafkaSendFailed     string = "kafka发送失败"
)

// Context keys
const (
	CtxUserIDKey   = "userID"
	CtxUsernameKey = "username"
	CtxUserRoleKey = "userRole"
)

// User roles
const (
	RoleUser  = "user"
	RoleAdmin = "admin"
)
