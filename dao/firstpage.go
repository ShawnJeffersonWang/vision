package dao

import (
	"vision/dao/postgres"
	"vision/models/entity"
)

func GetNews() (news []entity.News, err error) {
	if err = postgres.DB.Model(&entity.News{}).Find(&news).Error; err != nil {
		return
	}
	return
}

func GetProverb() (proverbs []entity.Proverb, err error) {
	if err = postgres.DB.Model(&entity.Proverb{}).Find(&proverbs).Error; err != nil {
		return
	}
	return
}

func GetCrop() (crops []entity.CropCategory, err error) {
	if err = postgres.DB.Preload("CropDetails").Find(&crops).Error; err != nil {
		return
	}
	return
}

func GetVideo() (videos []entity.Video, err error) {
	if err = postgres.DB.Model(&entity.Video{}).Find(&videos).Error; err != nil {
		return
	}
	return
}

func GetPoetry() (poetry []entity.Poetry, err error) {
	if err = postgres.DB.Model(&entity.Poetry{}).Find(&poetry).Error; err != nil {
		return
	}
	return
}

// AddNews 添加新闻
func AddNews(news *entity.News) (err error) {
	err = postgres.DB.Create(news).Error
	return
}

// AddProverb 添加谚语
func AddProverb(proverb *entity.Proverb) (err error) {
	err = postgres.DB.Create(proverb).Error
	return
}

// AddCropCategory 添加农作物种类
func AddCropCategory(crop *entity.CropCategory) (err error) {
	err = postgres.DB.Create(crop).Error
	return
}

// AddCropDetail 添加农作物细节
func AddCropDetail(detail *entity.CropDetail) (err error) {
	err = postgres.DB.Create(detail).Error
	return
}

// AddVideo 添加短视频
func AddVideo(video *entity.Video) (err error) {
	err = postgres.DB.Create(video).Error
	return
}

// AddPoetry 添加诗歌
func AddPoetry(poetry *entity.Poetry) (err error) {
	err = postgres.DB.Create(poetry).Error
	return
}
