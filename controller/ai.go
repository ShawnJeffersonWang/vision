package controller

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"agricultural_vision/constants"
	"agricultural_vision/logic"
	"agricultural_vision/middleware"
	"agricultural_vision/models/request"
)

func AiHandler(c *gin.Context) {
	userID, err := middleware.GetCurrentUserID(c)
	if err != nil {
		zap.L().Error("获取userID失败", zap.Error(err))
		ResponseError(c, http.StatusInternalServerError, constants.CodeServerBusy)
		return
	}

	// 解析前端传来的请求体
	var aiRequest request.AiRequest
	if err := c.ShouldBindJSON(&aiRequest); err != nil {
		zap.L().Error("参数校验失败", zap.Error(err))
		ResponseError(c, http.StatusBadRequest, constants.CodeInvalidParam)
	}

	// 调用逻辑层
	aiResponse, err := logic.AiTalk(aiRequest.UserInput, userID, aiRequest.Role)
	if err != nil {
		zap.L().Error("AI对话失败", zap.Error(err))
		ResponseError(c, http.StatusInternalServerError, constants.CodeServerBusy)
	} else {
		ResponseSuccess(c, aiResponse)
	}
}
