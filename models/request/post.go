package request

// 发布帖子
type CreatePostRequest struct {
	Content     string `json:"content" binding:"required"`      // 内容
	Image       string `json:"image"`                           // 图片（可选）
	CommunityID int64  `json:"community_id" binding:"required"` // 归属社区
}

// 分页批量查询
type ListRequest struct {
	Page  int64  `json:"page" form:"page"`   //查询第几页的数据
	Size  int64  `json:"size" form:"size"`   //每页数据条数
	Order string `json:"order" form:"order"` //排序方式
}
