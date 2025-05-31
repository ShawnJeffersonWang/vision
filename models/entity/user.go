package entity

import "time"

type User struct {
	BaseModel
	Username string `gorm:"type:varchar(64);not null" json:"username"`
	Email    string `gorm:"type:varchar(64);not null;unique" json:"email"`
	Password string `gorm:"type:varchar(64);not null" json:"-"`
	Role     string `gorm:"size:20;default:'user'" json:"role"` // 添加角色字段
	Avatar   string `gorm:"type:varchar(625);" json:"avatar"`   // 用户头像
	PostNum  int64  `gorm:"-" json:"post_num"`                  // 用户所发帖子的数量

	// 发过的帖子（HasMany关系）Post中的AuthorID是外键关联到User的ID
	Posts []Post `gorm:"foreignKey:AuthorID" json:"-"` // 定义关联关系

	// 发过的评论（HasMany关系）Comment中的AuthorID是外键关联到User的ID
	Comments []Comment `gorm:"foreignKey:AuthorID" json:"-"` // 定义关联关系

	// 点赞过的帖子（多对多）
	LikedPosts []Post `gorm:"many2many:user_likes_posts;" json:"-"`
}

// LoginHistory 登录历史
type LoginHistory struct {
	ID         int64     `gorm:"primaryKey;autoIncrement" json:"id"`
	UserID     int64     `gorm:"index;not null" json:"user_id"`
	Username   string    `gorm:"size:50" json:"username"` // 添加用户名字段
	LoginTime  time.Time `gorm:"not null" json:"login_time"`
	LoginIP    string    `gorm:"size:45" json:"login_ip"`
	UserAgent  string    `gorm:"size:255" json:"user_agent"`
	DeviceID   string    `gorm:"size:100" json:"device_id"`
	Success    bool      `gorm:"default:true" json:"success"`
	FailReason string    `gorm:"size:255" json:"fail_reason,omitempty"`
	CreatedAt  time.Time `json:"created_at"`
}
