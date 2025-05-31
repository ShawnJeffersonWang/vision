package controller

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"agricultural_vision/constants"
	"agricultural_vision/logic"
)

// 社区模块

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
