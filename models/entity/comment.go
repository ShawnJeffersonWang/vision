package entity

// 评论
type Comment struct {
	BaseModel
	Content string `gorm:"type:text;not null"`

	// 评论关联
	ParentID *int64     `gorm:"index;default:null" json:"parent_id"`                       // 父评论ID（null表示自身是顶级评论）
	RootID   *int64     `gorm:"ind ex;default:null" json:"root_id"`                        // 根评论ID（null表示自身是顶级评论）
	Replies  []*Comment `gorm:"foreignKey:ParentID;constraint:OnDelete:CASCADE;" json:"-"` // 删除评论时，级联删除子评论

	// 用户关联
	AuthorID int64 `gorm:"index;not null"`
	Author   User  `gorm:"foreignKey:AuthorID"`

	// 帖子关联
	PostID int64 `gorm:"index;not null"`
	Post   Post  `gorm:"foreignKey:PostID"` // 删除帖子时级联删除评论

	// 记录点赞用户（多对多）
	LikedBy []*User `gorm:"many2many:user_likes_comments;"`
}
