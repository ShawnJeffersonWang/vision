package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"agricultural_vision/constants"
	"agricultural_vision/dao/mysql"
)

func GetNewsHandler(c *gin.Context) {
	news, err := mysql.GetNews()
	if err != nil {
		zap.L().Error("获取新闻失败", zap.Error(err))
		ResponseError(c, http.StatusInternalServerError, constants.CodeServerBusy)
		return
	}

	ResponseSuccess(c, news)
	return
}

func GetProverbHandler(c *gin.Context) {
	proverbs, err := mysql.GetProverb()
	if err != nil {
		zap.L().Error("获取谚语失败", zap.Error(err))
		ResponseError(c, http.StatusInternalServerError, constants.CodeServerBusy)
		return
	}

	ResponseSuccess(c, proverbs)
	return
}

func GetCropHandler(c *gin.Context) {
	crops, err := mysql.GetCrop()
	if err != nil {
		zap.L().Error("获取农作物百科失败", zap.Error(err))
		ResponseError(c, http.StatusInternalServerError, constants.CodeServerBusy)
		return
	}

	ResponseSuccess(c, crops)
	return
}

func GetVideoHandler(c *gin.Context) {
	videos, err := mysql.GetVideo()
	if err != nil {
		zap.L().Error("获取视频失败", zap.Error(err))
		ResponseError(c, http.StatusInternalServerError, constants.CodeServerBusy)
		return
	}

	ResponseSuccess(c, videos)
	return
}

func GetPoetryHandler(c *gin.Context) {
	poetry, err := mysql.GetPoetry()
	if err != nil {
		zap.L().Error("获取古诗失败", zap.Error(err))
		ResponseError(c, http.StatusInternalServerError, constants.CodeServerBusy)
		return
	}

	ResponseSuccess(c, poetry)
	return
}
