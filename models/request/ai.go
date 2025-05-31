package request

// ai对话
type AiRequest struct {
	UserInput string `json:"user_input" binding:"required"` // 前端传来的问题
	Role      int64  `json:"role" binding:"required"`       // 选择的模型角色
}
