package utils

import (
	"agricultural_vision/dao/mysql"
	"agricultural_vision/models/entity"
)

func InitSqlTable() (err error) {
	err = mysql.DB.AutoMigrate(
		&entity.User{},

		&entity.News{},
		&entity.Proverb{},
		&entity.CropCategory{},
		&entity.CropDetail{},
		&entity.Video{},
		&entity.Poetry{},

		&entity.Post{},
		&entity.Community{},
		&entity.Comment{},
	)
	return
}
