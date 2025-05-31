package controller

import (
	"errors"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.uber.org/zap"

	"agricultural_vision/constants"
	"agricultural_vision/logic"
	"agricultural_vision/middleware"
	"agricultural_vision/models/request"
)

// 帖子投票
func PostVoteController(c *gin.Context) {
	p := new(request.VoteRequest)
	//参数校验
	err := c.ShouldBindJSON(p)
	if err != nil {
		errs, ok := err.(validator.ValidationErrors)
		if !ok {
			ResponseError(c, http.StatusBadRequest, constants.CodeInvalidParam)
			return
		} else {
			errData := removeTopStruct(errs.Translate(trans)) //翻译错误
			ResponseError(c, http.StatusBadRequest, errData)
			return
		}
	}

	//业务逻辑
	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		ResponseError(c, http.StatusBadRequest, constants.CodeInvalidParam)
		return
	}

	err = logic.VoteForPost(userID, p)
	if err != nil {
		zap.L().Error("投票失败", zap.Error(err))
		if errors.Is(err, constants.ErrorNoPost) {
			ResponseError(c, http.StatusBadRequest, constants.CodeNoPost)
			return
		}
		ResponseError(c, http.StatusBadRequest, constants.CodeInvalidParam)
		return
	}

	//返回响应
	ResponseSuccess(c, nil)
}

// 评论投票
func CommentVoteController(c *gin.Context) {
	votePostRequest := new(request.VoteRequest)
	err := c.ShouldBindJSON(votePostRequest)
	if err != nil {
		zap.L().Error("参数不正确", zap.Error(err))
		ResponseError(c, http.StatusBadRequest, constants.CodeInvalidParam)
		return
	}

	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		ResponseError(c, http.StatusInternalServerError, constants.CodeServerBusy)
		return
	}

	err = logic.CommentVote(userID, votePostRequest)
	if err != nil {
		zap.L().Error("评论投票失败", zap.Error(err))
		if errors.Is(err, constants.ErrorNoComment) {
			ResponseError(c, http.StatusBadRequest, constants.CodeNoComment)
			return
		}
		ResponseError(c, http.StatusInternalServerError, constants.CodeServerBusy)
		return
	}

	ResponseSuccess(c, nil)
}
