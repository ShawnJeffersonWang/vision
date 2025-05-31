package entity

type User struct {
	BaseModel
	Username string `gorm:"type:varchar(64);not null" json:"username"`
	Email    string `gorm:"type:varchar(64);not null;unique" json:"email"`
	Password string `gorm:"type:varchar(64);not null" json:"-"`
	Avatar   string `gorm:"type:varchar(625);" json:"avatar"` // 用户头像
	PostNum  int64  `gorm:"-" json:"post_num"`                // 用户所发帖子的数量

	// 发过的帖子（HasMany关系）Post中的AuthorID是外键关联到User的ID
	Posts []Post `gorm:"foreignKey:AuthorID" json:"-"` // 定义关联关系

	// 发过的评论（HasMany关系）Comment中的AuthorID是外键关联到User的ID
	Comments []Comment `gorm:"foreignKey:AuthorID" json:"-"` // 定义关联关系

	// 点赞过的帖子（多对多）
	LikedPosts []Post `gorm:"many2many:user_likes_posts;" json:"-"`
}
