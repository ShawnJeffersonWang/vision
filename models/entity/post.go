package entity

// 帖子
type Post struct {
	BaseModel
	Content string `gorm:"type:text;not_null" json:"content"`
	Image   string `gorm:"type:text" json:"image"`

	// 用户关联（BelongsTo关系）
	AuthorID int64 `gorm:"index;not null" json:"author_id"`
	Author   User  `gorm:"foreignKey:AuthorID" json:"-"` // 实现预加载用户信息

	// 社区关联（BelongsTo关系）
	CommunityID int64     `gorm:"index;not null" json:"community_id"`
	Community   Community `gorm:"foreignKey:CommunityID"` // 实现预加载社区信息

	// 评论关联（HasMany关系，级联删除）
	Comments []Comment `gorm:"foreignKey:PostID;constraint:OnDelete:CASCADE;"` // 定义关联关系

	// 记录收藏用户（多对多）
	CollectedBy []*User `gorm:"many2many:user_likes_comments;"`
}
