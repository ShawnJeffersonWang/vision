package controller

import (
	"net/http"
	"vision/dao"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"vision/constants"
)

func GetNewsHandler(c *gin.Context) {
	news, err := dao.GetNews()
	if err != nil {
		zap.L().Error("获取新闻失败", zap.Error(err))
		ResponseError(c, http.StatusInternalServerError, constants.CodeServerBusy)
		return
	}

	ResponseSuccess(c, news)
	return
}

func GetProverbHandler(c *gin.Context) {
	proverbs, err := dao.GetProverb()
	if err != nil {
		zap.L().Error("获取谚语失败", zap.Error(err))
		ResponseError(c, http.StatusInternalServerError, constants.CodeServerBusy)
		return
	}

	ResponseSuccess(c, proverbs)
	return
}

func GetCropHandler(c *gin.Context) {
	crops, err := dao.GetCrop()
	if err != nil {
		zap.L().Error("获取农作物百科失败", zap.Error(err))
		ResponseError(c, http.StatusInternalServerError, constants.CodeServerBusy)
		return
	}

	ResponseSuccess(c, crops)
	return
}

func GetVideoHandler(c *gin.Context) {
	videos, err := dao.GetVideo()
	if err != nil {
		zap.L().Error("获取视频失败", zap.Error(err))
		ResponseError(c, http.StatusInternalServerError, constants.CodeServerBusy)
		return
	}

	ResponseSuccess(c, videos)
	return
}

func GetPoetryHandler(c *gin.Context) {
	poetry, err := dao.GetPoetry()
	if err != nil {
		zap.L().Error("获取古诗失败", zap.Error(err))
		ResponseError(c, http.StatusInternalServerError, constants.CodeServerBusy)
		return
	}

	ResponseSuccess(c, poetry)
	return
}
