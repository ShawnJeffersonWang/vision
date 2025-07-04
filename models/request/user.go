package request

// 用户注册
type SignUpRequest struct {
	Username string `json:"username" binding:"required"`
	Email    string `json:"email" binding:"required,email"`
	Code     string `json:"code" binding:"required"` // 邮箱验证码
	Password string `json:"password" binding:"required"`
}

// 用户登录
type LoginRequest struct {
	Email      string `json:"email" binding:"required,email"`
	Password   string `json:"password" binding:"required"`
	DeviceID   string `json:"device_id,omitempty"`
	DeviceInfo string `json:"device_info,omitempty"`
	ClientIP   string `json:"-"` // 从context中获取，不从JSON解析
	UserAgent  string `json:"-"` // 从context中获取，不从JSON解析
}

// 发送验证码
type SendVerificationCodeRequest struct {
	Email string `json:"email" binding:"required"`
}

// 修改密码
type ChangePasswordRequest struct {
	Email    string `json:"email" binding:"required"`
	Code     string `json:"code" binding:"required"` // 邮箱验证码
	Password string `json:"password" binding:"required"`
}

// 修改用户信息
type UpdateUserInfoRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Avatar   string `json:"avatar"`
}

type GetLoginHistoryRequest struct {
	Page     int `form:"page" binding:"min=1"`
	PageSize int `form:"page_size" binding:"min=1,max=100"`
}
