package request

// CreateCommunityRequest 创建社区请求结构体
type CreateCommunityRequest struct {
	CommunityName string `json:"community_name" binding:"required,min=1,max=128"`
	Introduction  string `json:"introduction" binding:"required,min=1,max=625"`
}
