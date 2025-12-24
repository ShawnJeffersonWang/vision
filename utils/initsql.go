package utils

import (
	"vision/dao/postgres"
	"vision/models/entity"
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
		&entity.PostVote{},
	)
	return
}
