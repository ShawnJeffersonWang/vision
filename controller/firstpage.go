package controller

import (
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"time"
	"vision/dao"
	"vision/models/entity"
	"vision/pkg/alioss"
	"vision/settings"

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

// UploadMaterialHandler 通用素材上传接口 (图片/视频)
// 参考了 UploadPostImageHandler 的逻辑
func UploadMaterialHandler(c *gin.Context) {
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		zap.L().Error("获取上传文件失败", zap.Error(err))
		ResponseError(c, http.StatusBadRequest, constants.CodeInvalidParam)
		return
	}
	defer func(file multipart.File) {
		_ = file.Close()
	}(file)

	// 限制文件大小（例如 10MB，根据需求调整）
	if header.Size > 10*1024*1024 {
		ResponseError(c, http.StatusBadRequest, "文件大小超出限制")
		return
	}

	// 获取扩展名
	ext := filepath.Ext(header.Filename)
	// 生成唯一文件名
	newFileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)

	// 上传到 OSS (假设 settings 中有一个通用的 MaterialPath，如果没有请替换为具体的字符串路径)
	// 如果没有 MaterialPath，可以暂时复用 PostImagePtah 或写死一个路径如 "materials/"
	path := settings.Conf.AliossConfig.PostImagePtah

	fileURL, err := alioss.UploadFile(file, newFileName, path)
	if err != nil {
		zap.L().Error("上传文件失败", zap.Error(err))
		ResponseError(c, http.StatusInternalServerError, constants.CodeServerBusy)
		return
	}
	ResponseSuccess(c, fileURL)
}

// AddNewsHandler 添加新闻
func AddNewsHandler(c *gin.Context) {
	var p entity.News
	if err := c.ShouldBindJSON(&p); err != nil {
		ResponseError(c, http.StatusBadRequest, constants.CodeInvalidParam)
		return
	}
	if err := dao.AddNews(&p); err != nil {
		zap.L().Error("添加新闻失败", zap.Error(err))
		ResponseError(c, http.StatusInternalServerError, constants.CodeServerBusy)
		return
	}
	ResponseSuccess(c, "添加成功")
}

// AddProverbHandler 添加谚语
func AddProverbHandler(c *gin.Context) {
	var p entity.Proverb
	if err := c.ShouldBindJSON(&p); err != nil {
		ResponseError(c, http.StatusBadRequest, constants.CodeInvalidParam)
		return
	}
	if err := dao.AddProverb(&p); err != nil {
		zap.L().Error("添加谚语失败", zap.Error(err))
		ResponseError(c, http.StatusInternalServerError, constants.CodeServerBusy)
		return
	}
	ResponseSuccess(c, "添加成功")
}

// AddCropCategoryHandler 添加农作物种类
func AddCropCategoryHandler(c *gin.Context) {
	var p entity.CropCategory
	if err := c.ShouldBindJSON(&p); err != nil {
		ResponseError(c, http.StatusBadRequest, constants.CodeInvalidParam)
		return
	}
	if err := dao.AddCropCategory(&p); err != nil {
		zap.L().Error("添加农作物种类失败", zap.Error(err))
		ResponseError(c, http.StatusInternalServerError, constants.CodeServerBusy)
		return
	}
	ResponseSuccess(c, "添加成功")
}

// AddCropDetailHandler 添加农作物细节
// 使用 Query 参数传递 category_id，例如: POST /cms/crop/detail?category_id=1
func AddCropDetailHandler(c *gin.Context) {
	var p entity.CropDetail
	if err := c.ShouldBindJSON(&p); err != nil {
		ResponseError(c, http.StatusBadRequest, constants.CodeInvalidParam)
		return
	}

	// 处理 CategoryId (因为 json:"-" 忽略了该字段)
	cidStr := c.Query("category_id")
	cid, err := strconv.ParseInt(cidStr, 10, 64)
	if err != nil || cid == 0 {
		ResponseError(c, http.StatusBadRequest, constants.CodeInvalidParam)
		return
	}
	p.CategoryId = cid

	if err := dao.AddCropDetail(&p); err != nil {
		zap.L().Error("添加农作物细节失败", zap.Error(err))
		ResponseError(c, http.StatusInternalServerError, constants.CodeServerBusy)
		return
	}
	ResponseSuccess(c, "添加成功")
}

// AddVideoHandler 添加视频
func AddVideoHandler(c *gin.Context) {
	var p entity.Video
	if err := c.ShouldBindJSON(&p); err != nil {
		ResponseError(c, http.StatusBadRequest, constants.CodeInvalidParam)
		return
	}
	if err := dao.AddVideo(&p); err != nil {
		zap.L().Error("添加视频失败", zap.Error(err))
		ResponseError(c, http.StatusInternalServerError, constants.CodeServerBusy)
		return
	}
	ResponseSuccess(c, "添加成功")
}

// AddPoetryHandler 添加诗歌
func AddPoetryHandler(c *gin.Context) {
	var p entity.Poetry
	if err := c.ShouldBindJSON(&p); err != nil {
		ResponseError(c, http.StatusBadRequest, constants.CodeInvalidParam)
		return
	}
	if err := dao.AddPoetry(&p); err != nil {
		zap.L().Error("添加诗歌失败", zap.Error(err))
		ResponseError(c, http.StatusInternalServerError, constants.CodeServerBusy)
		return
	}
	ResponseSuccess(c, "添加成功")
}
