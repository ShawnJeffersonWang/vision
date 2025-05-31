package response

// 帖子列表
type PostResponse struct {
	ID           int64                  `json:"id"`
	Content      string                 `json:"content"`
	Image        string                 `json:"image"`
	Author       UserBriefResponse      `json:"author"`        // 作者
	LikeCount    int64                  `json:"like_count"`    // 点赞数
	Liked        bool                   `json:"liked"`         // 当前用户是否已点赞
	CommentCount int64                  `json:"comment_count"` // 评论数
	CreatedAt    string                 `json:"created_at"`    // 发布时间
	Community    CommunityBriefResponse `json:"community"`     // 所属社区信息
}

type PostListResponse struct {
	Posts []*PostResponse `json:"posts"`
	Total int64           `json:"total"`
}
