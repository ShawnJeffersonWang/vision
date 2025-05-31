package controller

import (
	"agricultural_vision/models/request"
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"agricultural_vision/constants"
	"agricultural_vision/logic"
)

// 社区模块

// CreateCommunityHandler 创建社区
func CreateCommunityHandler(c *gin.Context) {
	// 获取请求参数
	var req request.CreateCommunityRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		zap.L().Error("创建社区请求参数绑定失败", zap.Error(err))
		ResponseError(c, http.StatusBadRequest, constants.CodeInvalidParam)
		return
	}

	// 调用业务逻辑创建社区
	err := logic.CreateCommunity(&req)
	if err != nil {
		zap.L().Error("创建社区失败", zap.Error(err))

		// 根据不同错误类型返回不同响应
		if errors.Is(err, constants.ErrorCommunityNameExists) {
			ResponseError(c, http.StatusConflict, constants.CodeCommunityNameExists)
			return
		}

		ResponseError(c, http.StatusInternalServerError, constants.CodeServerBusy)
		return
	}

	// 创建成功
	ResponseSuccess(c, "社区创建成功")
}

// 查询所有社区
func CommunityHandler(c *gin.Context) {
	data, err := logic.GetCommunityList()
	if err != nil {
		zap.L().Error("获取社区列表失败", zap.Error(err))
		if errors.Is(err, constants.ErrorNoResult) {
			ResponseError(c, http.StatusOK, constants.CodeNoResult)
			return
		}
		ResponseError(c, http.StatusInternalServerError, constants.CodeServerBusy)
		return
	}
	ResponseSuccess(c, data)
}

// 查询社区详情
func CommunityDetailHandler(c *gin.Context) {
	//1.获取社区id
	idStr := c.Param("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	//如果获取请求参数失败
	if err != nil {
		zap.L().Error("获取社区详情的参数不正确", zap.Error(err))
		ResponseError(c, http.StatusBadRequest, constants.CodeInvalidParam)
		return
	}

	//查询到所有的社区，以列表形式返回
	data, err := logic.GetCommunityDetail(id)
	if err != nil {
		zap.L().Error("获取社区详情失败", zap.Error(err))
		if errors.Is(err, constants.ErrorNoResult) {
			ResponseError(c, http.StatusOK, constants.CodeNoResult)
			return
		}
		ResponseError(c, http.StatusInternalServerError, constants.CodeServerBusy)
		return
	}
	ResponseSuccess(c, data)
}
