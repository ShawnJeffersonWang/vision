package entity

// 社区
type Community struct {
	BaseModel
	CommunityName string `gorm:"type:varchar(128);not null;uniqueIndex" json:"community_name"`
	Introduction  string `gorm:"type:varchar(625);not null" json:"introduction"`

	// 关联帖子（一对多关系）
	Posts []Post `gorm:"foreignKey:CommunityID" json:"-"`
}
