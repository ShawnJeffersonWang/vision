package controller

import (
	"agricultural_vision/pkg/alioss"
	"agricultural_vision/settings"
	"errors"
	"fmt"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"agricultural_vision/constants"
	"agricultural_vision/logic"
	"agricultural_vision/middleware"
	"agricultural_vision/models/request"
)

// 发布帖子
func CreatePostHandler(c *gin.Context) {
	//1.获取参数及参数的校验
	p := new(request.CreatePostRequest)
	//将参数绑定到p中
	if err := c.ShouldBindJSON(p); err != nil {
		zap.L().Error("请求参数错误", zap.Error(err))
		ResponseError(c, http.StatusBadRequest, constants.CodeInvalidParam)
		return
	}
	//在请求上下文中获取userID
	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, constants.CodeServerBusy)
		return
	}

	//2.创建帖子
	data, err := logic.CreatePost(p, userID)
	if err != nil {
		zap.L().Error("创建帖子失败", zap.Error(err))
		ResponseError(c, http.StatusInternalServerError, constants.CodeServerBusy)
		return
	}

	//3.返回响应
	ResponseSuccess(c, data)
}

// 删除帖子
func DeletePostHandler(c *gin.Context) {
	postID := c.Param("id")

	postIDStr, err := strconv.ParseInt(postID, 10, 64)
	if err != nil {
		zap.L().Error("请求参数错误", zap.Error(err))
		ResponseError(c, http.StatusBadRequest, constants.CodeInvalidParam)
		return
	}

	//在请求上下文中获取userID
	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, constants.CodeServerBusy)
		return
	}

	if err := logic.DeletePost(postIDStr, userID); err != nil {
		zap.L().Error("删除帖子失败", zap.Error(err))
		if errors.Is(err, constants.ErrorNoPermission) {
			ResponseError(c, http.StatusForbidden, constants.CodeNoPermission)
			return
		} else if errors.Is(err, constants.ErrorNoPost) {
			ResponseError(c, http.StatusBadRequest, constants.CodeNoPost)
			return
		}
		ResponseError(c, http.StatusInternalServerError, constants.CodeServerBusy)
		return
	}
	ResponseSuccess(c, nil)
}

// 查询帖子列表
func GetPostListHandler(c *gin.Context) {
	//初始化结构体时指定初始默认参数
	p := &request.ListRequest{
		Page:  1,
		Size:  10,
		Order: constants.OrderScore,
	}
	err := c.ShouldBindQuery(p)
	if err != nil {
		zap.L().Error("请求参数错误", zap.Error(err))
		ResponseError(c, http.StatusBadRequest, constants.CodeInvalidParam)
		return
	}

	// 获取userID
	userID, err := middleware.GetCurrentUserID(c)
	if errors.Is(err, constants.ErrorNeedLogin) {
		// 如果用户未登录，也可以查询帖子列表
		data, err := logic.GetPostList(p, 0)
		if err != nil {
			zap.L().Error("指定顺序查询帖子列表失败", zap.Error(err))
			ResponseError(c, http.StatusInternalServerError, constants.CodeServerBusy)
			return
		}
		ResponseSuccess(c, data)
		return
	}

	// 用户已登录，查询帖子列表
	data, err := logic.GetPostList(p, userID)
	if err != nil {
		zap.L().Error("指定顺序查询帖子列表失败", zap.Error(err))
		ResponseError(c, http.StatusInternalServerError, constants.CodeServerBusy)
		return
	}
	ResponseSuccess(c, data)
	return
}

// 查询该社区分类下的帖子详情列表
func GetCommunityPostListHandler(c *gin.Context) {
	//初始化结构体时指定初始默认参数
	p := &request.ListRequest{
		Page:  1,
		Size:  10,
		Order: constants.OrderTime,
	}
	err1 := c.ShouldBindQuery(p)

	communityIDStr := c.Param("id")
	communityID, err2 := strconv.ParseInt(communityIDStr, 10, 64)

	if err1 != nil || err2 != nil {
		zap.L().Error("请求参数错误", zap.Error(err1), zap.Error(err2))
		ResponseError(c, http.StatusBadRequest, constants.CodeInvalidParam)
		return
	}

	// 获取userID
	userID, err := middleware.GetCurrentUserID(c)
	if errors.Is(err, constants.ErrorNeedLogin) {
		data, err := logic.GetCommunityPostList(p, communityID, 0)
		if err != nil {
			zap.L().Error("根据社区查询帖子列表失败", zap.Error(err))
			ResponseError(c, http.StatusInternalServerError, constants.CodeServerBusy)
			return
		}
		ResponseSuccess(c, data)
		return
	}

	//根据社区查询该社区分类下的帖子列表
	data, err := logic.GetCommunityPostList(p, communityID, userID)
	if err != nil {
		zap.L().Error("根据社区查询帖子列表失败", zap.Error(err))
		ResponseError(c, http.StatusInternalServerError, constants.CodeServerBusy)
		return
	}
	ResponseSuccess(c, data)
	return
}

// 获取用户帖子列表
func GetUserPostListHandler(c *gin.Context) {
	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		zap.L().Error("获取userID失败", zap.Error(err))
		ResponseError(c, http.StatusInternalServerError, constants.CodeServerBusy)
		return
	}

	listRequest := &request.ListRequest{
		Page:  1,
		Size:  10,
		Order: constants.OrderTime,
	}
	if err := c.ShouldBindQuery(listRequest); err != nil {
		zap.L().Error("参数校验失败", zap.Error(err))
		ResponseError(c, http.StatusBadRequest, constants.CodeInvalidParam)
		return
	}

	data, err := logic.GetUserPostList(userID, listRequest)
	if err != nil {
		zap.L().Error("获取用户帖子列表失败", zap.Error(err))
		ResponseError(c, http.StatusInternalServerError, constants.CodeServerBusy)
		return
	}
	ResponseSuccess(c, data)
}

// 获取用户点赞帖子列表
func GetUserLikedPostListHandler(c *gin.Context) {
	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		zap.L().Error("获取userID失败", zap.Error(err))
		ResponseError(c, http.StatusInternalServerError, constants.CodeServerBusy)
		return
	}

	listRequest := &request.ListRequest{
		Page:  1,
		Size:  10,
		Order: constants.OrderTime,
	}
	if err := c.ShouldBindQuery(listRequest); err != nil {
		zap.L().Error("参数校验失败", zap.Error(err))
		ResponseError(c, http.StatusBadRequest, constants.CodeInvalidParam)
		return
	}

	data, err := logic.GetUserLikedPostList(userID, listRequest)
	if err != nil {
		zap.L().Error("获取用户点赞帖子列表失败", zap.Error(err))
		ResponseError(c, http.StatusInternalServerError, constants.CodeServerBusy)
		return
	}
	ResponseSuccess(c, data)
}

// 上传帖子图片
func UploadPostImageHandler(c *gin.Context) {
	// 获取上传的文件
	file, header, err := c.Request.FormFile("file")
	if err != nil {
		zap.L().Error("获取上传文件失败", zap.Error(err))
		ResponseError(c, http.StatusBadRequest, constants.CodeInvalidParam)
		return
	}
	defer func(file multipart.File) {
		_ = file.Close()
	}(file)

	// 限制文件大小（5MB）
	if header.Size > 5*1024*1024 {
		zap.L().Error("文件大小超出5MB", zap.Error(err))
		ResponseError(c, http.StatusBadRequest, "文件大小超出5MB")
		return
	}

	// 获取文件扩展名ext
	ext := filepath.Ext(header.Filename)
	if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
		zap.L().Error("文件格式不支持", zap.Error(err))
		ResponseError(c, http.StatusBadRequest, "文件格式不支持")
		return
	}

	// 生成唯一文件名
	newFileName := fmt.Sprintf("%d%s", time.Now().UnixNano(), ext)

	// 上传到 OSS
	fileURL, err := alioss.UploadFile(file, newFileName, settings.Conf.AliossConfig.PostImagePtah)
	if err != nil {
		zap.L().Error("上传文件失败", zap.Error(err))
		ResponseError(c, http.StatusInternalServerError, constants.CodeServerBusy)
		return
	}
	ResponseSuccess(c, fileURL)
}
