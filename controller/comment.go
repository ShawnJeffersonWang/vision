package controller

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	"agricultural_vision/constants"
	"agricultural_vision/logic"
	"agricultural_vision/middleware"
	"agricultural_vision/models/request"
)

// 创建评论
func CreateCommentHandler(c *gin.Context) {
	// 获取参数
	createCommentRequest := &request.CreateCommentRequest{}
	if err := c.ShouldBindJSON(createCommentRequest); err != nil {
		zap.L().Error("参数不正确", zap.Error(err))
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			ResponseError(c, http.StatusBadRequest, constants.CodeInvalidParam)
			return
		}
		ResponseError(c, http.StatusBadRequest, removeTopStruct(errs.Translate(trans)))
		return
	}

	// 获取userID
	id, err := middleware.GetCurrentUserID(c)
	if err != nil {
		zap.L().Error("获取userID失败", zap.Error(err))
		ResponseError(c, http.StatusInternalServerError, constants.CodeServerBusy)
		return
	}

	// 创建评论
	data, err := logic.CreateComment(createCommentRequest, id)
	if err != nil {
		zap.L().Error("创建评论失败", zap.Error(err))
		if errors.Is(err, constants.ErrorNoPost) {
			ResponseError(c, http.StatusBadRequest, constants.CodeNoPost)
			return
		}
		ResponseError(c, http.StatusInternalServerError, constants.CodeServerBusy)
		return
	}
	ResponseSuccess(c, data)
}

// 删除评论
func DeleteCommentHandler(c *gin.Context) {
	// 获取参数
	commentIDStr := c.Param("id")
	commentID, err := strconv.ParseInt(commentIDStr, 10, 64)
	if err != nil {
		zap.L().Error("参数不正确", zap.Error(err))
		ResponseError(c, http.StatusBadRequest, constants.CodeInvalidParam)
		return
	}

	// 获取userID
	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		zap.L().Error("获取userID失败", zap.Error(err))
		ResponseError(c, http.StatusInternalServerError, constants.CodeServerBusy)
		return
	}

	// 删除评论
	err = logic.DeleteComment(commentID, userID)
	if err != nil {
		zap.L().Error("删除评论失败", zap.Error(err))
		if errors.Is(err, constants.ErrorNoComment) {
			ResponseError(c, http.StatusBadRequest, constants.CodeNoComment)
			return
		} else if errors.Is(err, constants.ErrorNoPermission) {
			ResponseError(c, http.StatusBadRequest, constants.CodeNoPermission)
			return
		}
		ResponseError(c, http.StatusInternalServerError, constants.CodeServerBusy)
		return
	}

	ResponseSuccess(c, nil)
}

// 查询顶级评论
func GetTopCommentListHandler(c *gin.Context) {
	// 获取参数
	postIDStr := c.Param("post_id")
	postID, err1 := strconv.ParseInt(postIDStr, 10, 64)

	listRequest := &request.ListRequest{
		Page:  1,
		Size:  10,
		Order: constants.OrderTime,
	}
	err2 := c.ShouldBindQuery(listRequest)

	if err1 != nil || err2 != nil {
		zap.L().Error("参数不正确", zap.Error(err1), zap.Error(err2))
		ResponseError(c, http.StatusBadRequest, constants.CodeInvalidParam)
		return
	}

	// 获取userID
	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		zap.L().Error("获取userID失败", zap.Error(err))
		ResponseError(c, http.StatusInternalServerError, constants.CodeServerBusy)
		return
	}

	// 查询顶级评论
	commentListResponse, err := logic.GetTopCommentList(postID, listRequest, userID)
	if err != nil {
		zap.L().Error("查询顶级评论失败", zap.Error(err))
		ResponseError(c, http.StatusInternalServerError, constants.CodeServerBusy)
		return
	}

	ResponseSuccess(c, commentListResponse)
}

// 查询子评论
func GetSonCommentListHandler(c *gin.Context) {
	commentIDStr := c.Param("comment_id")
	commentID, err1 := strconv.ParseInt(commentIDStr, 10, 64)

	listRequest := &request.ListRequest{
		Page: 1,
		Size: 10,
	}
	err2 := c.ShouldBindQuery(listRequest)

	if err1 != nil || err2 != nil {
		zap.L().Error("参数不正确", zap.Error(err1), zap.Error(err2))
		ResponseError(c, http.StatusBadRequest, constants.CodeInvalidParam)
		return
	}

	// 获取userID
	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		zap.L().Error("获取userID失败", zap.Error(err))
		ResponseError(c, http.StatusInternalServerError, constants.CodeServerBusy)
		return
	}

	commentListResponse, err := logic.GetSonCommentList(commentID, listRequest, userID)
	if err != nil {
		zap.L().Error("查询子评论失败", zap.Error(err))
		ResponseError(c, http.StatusInternalServerError, constants.CodeServerBusy)
		return
	}
	ResponseSuccess(c, commentListResponse)
}

// 查询帖子的所有评论
func GetCommentListHandler(c *gin.Context) {
	// 绑定参数
	postIDStr := c.Param("post_id")
	postID, err1 := strconv.ParseInt(postIDStr, 10, 64)
	listRequest := &request.ListRequest{
		Page:  1,
		Size:  10,
		Order: constants.OrderTime,
	}
	err2 := c.ShouldBindQuery(&listRequest)
	if err1 != nil || err2 != nil {
		zap.L().Error("参数不正确", zap.Error(err1), zap.Error(err2))
		ResponseError(c, http.StatusBadRequest, constants.CodeInvalidParam)
		return
	}

	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		zap.L().Error("获取userID失败", zap.Error(err))
		return
	}

	commentListResponse, err := logic.GetCommentList(postID, listRequest, userID)
	if err != nil {
		zap.L().Error("查询评论失败", zap.Error(err))
		ResponseError(c, http.StatusInternalServerError, constants.CodeServerBusy)
		return
	}
	ResponseSuccess(c, commentListResponse)
}
