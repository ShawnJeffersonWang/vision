package request

// 发布评论
type CreateCommentRequest struct {
	Content  string `json:"content" binding:"required"` // 内容
	PostID   int64  `json:"post_id" binding:"required"` // 帖子ID
	ParentID *int64 `json:"parent_id,omitempty"`        // 父评论ID（null表示自身是顶级评论）
	RootID   *int64 `json:"root_id,omitempty"`          // 根评论ID（null表示自身是顶级评论）
}
