package logic

import (
	"agricultural_vision/constants"
	"agricultural_vision/dao"
	"agricultural_vision/models/entity"
	"agricultural_vision/models/request"
	"agricultural_vision/models/response"
)

// CreateCommunity 创建社区
func CreateCommunity(req *request.CreateCommunityRequest) error {
	// 检查社区名称是否已存在
	exists, err := dao.CheckCommunityNameExists(req.CommunityName)
	if err != nil {
		return err
	}
	if exists {
		return constants.ErrorCommunityNameExists // 需要定义这个错误常量
	}

	// 创建社区实体
	community := &entity.Community{
		CommunityName: req.CommunityName,
		Introduction:  req.Introduction,
	}

	// 调用数据层创建社区
	return dao.CreateCommunity(community)
}

func GetCommunityList() ([]*response.CommunityBriefResponse, error) {
	return dao.GetCommunityList()
}

func GetCommunityDetail(id int64) (*entity.Community, error) {
	return dao.GetCommunityById(id)
}
