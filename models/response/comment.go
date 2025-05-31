package response

// 评论
type CommentResponse struct {
	ID           int64              `json:"id"`
	Content      string             `json:"content"`                 // 内容
	Author       *UserBriefResponse `json:"author"`                  // 作者
	LikeCount    int64              `json:"like_count"`              // 点赞数
	Liked        bool               `json:"liked"`                   // 当前用户是否已点赞
	RepliesCount *int64             `json:"replies_count,omitempty"` // 子评论数（只有一级评论需要），指针类型可以使值为0时依然在json中返回
	Parent       *UserBriefResponse `json:"parent,omitempty"`        // 父评论的作者信息（只有二级以上评论需要）
	CreatedAt    string             `json:"created_at"`              // 发布时间
	RootID       int64              `json:"root_id,omitempty"`       // 根评论id（子评论都需要）
	ParentID     int64              `json:"parent_id,omitempty"`     // 父评论id（为适配前端，只有二级以上评论需要）
}

// 分页查询评论响应体
type CommentListResponse struct {
	Comments []*CommentResponse `json:"comments"`
	Total    int64              `json:"total"`
}
