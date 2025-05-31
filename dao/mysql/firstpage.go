package mysql

import (
	"agricultural_vision/models/entity"
)

func GetNews() (news []entity.News, err error) {
	if err = DB.Model(&entity.News{}).Find(&news).Error; err != nil {
		return
	}
	return
}

func GetProverb() (proverbs []entity.Proverb, err error) {
	if err = DB.Model(&entity.Proverb{}).Find(&proverbs).Error; err != nil {
		return
	}
	return
}

func GetCrop() (crops []entity.CropCategory, err error) {
	if err = DB.Preload("CropDetails").Find(&crops).Error; err != nil {
		return
	}
	return
}

func GetVideo() (videos []entity.Video, err error) {
	if err = DB.Model(&entity.Video{}).Find(&videos).Error; err != nil {
		return
	}
	return
}

func GetPoetry() (poetry []entity.Poetry, err error) {
	if err = DB.Model(&entity.Poetry{}).Find(&poetry).Error; err != nil {
		return
	}
	return
}
