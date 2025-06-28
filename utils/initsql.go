package utils

import (
	"agricultural_vision/dao/postgres"
	"agricultural_vision/models/entity"
)

func InitSqlTable() (err error) {
	err = postgres.DB.AutoMigrate(
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
		&entity.LoginHistory{},
	)
	return
}
