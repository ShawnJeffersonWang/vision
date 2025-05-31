package request

// 帖子或评论投票
type VoteRequest struct {
	PostID    int64 `json:"post_id,omitempty"`                      // 可选，帖子ID
	CommentID int64 `json:"comment_id,omitempty"`                   // 可选，评论ID
	Direction int8  `json:"direction" bind:"oneof=-1 0 1;required"` // 赞成票(1)or反对票(-1)or取消投票(0)
}
