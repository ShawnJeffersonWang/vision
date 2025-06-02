package kafka

import "time"

// PostCreationMessage 发布帖子的 Kafka 消息结构
type PostCreationMessage struct {
	MessageID   string    `json:"messageId"`   // 唯一消息 ID
	PostID      int64     `json:"postId"`      // POST ID
	UserID      int64     `json:"userId"`      // 作者 ID
	Content     string    `json:"content"`     // 帖子内容
	Image       string    `json:"image"`       // 图片 URL
	CommunityID int64     `json:"communityId"` // 社区 ID
	CreatedAt   time.Time `json:"createdAt"`   // 创建时间
}
