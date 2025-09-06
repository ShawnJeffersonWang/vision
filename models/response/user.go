package response

import "vision/models/entity"

// 用户详情
type UserHomePageResponse struct {
	ID         int64             `json:"id"`
	Username   string            `json:"username"`
	Email      string            `json:"email"`
	Avatar     string            `json:"avatar"`
	Posts      *PostListResponse `json:"posts"`       // 该用户发布的帖子
	LikedPosts *PostListResponse `json:"liked_posts"` // 该用户点赞的帖子
}

// 用户简略信息（帖子和评论中展示）
type UserBriefResponse struct {
	ID       int64  `json:"id"`
	Username string `json:"username"` // 用户名
	Avatar   string `json:"avatar"`   // 头像
}

// 定义token响应结构
type TokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
	ExpiresIn    int64  `json:"expires_in"`
	TokenType    string `json:"token_type"`
}

type LoginHistoryResponse struct {
	Total    int64                  `json:"total"`
	Page     int                    `json:"page"`
	PageSize int                    `json:"page_size"`
	List     []*entity.LoginHistory `json:"list"`
}
