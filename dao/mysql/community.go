package mysql

import (
	"errors"

	"gorm.io/gorm"

	"agricultural_vision/constants"
	"agricultural_vision/models/entity"
	"agricultural_vision/models/response"
)

// 查询社区列表
func GetCommunityList() ([]*response.CommunityBriefResponse, error) {
	var communities []*response.CommunityBriefResponse

	result := DB.Model(&entity.Community{}).
		Select("id", "community_name").
		Find(&communities)

	// 如果未查询到结果
	if result.RowsAffected == 0 {
		return nil, constants.ErrorNoResult
	}

	return communities, result.Error
}

// 根据ID获取社区详情
func GetCommunityById(id int64) (*entity.Community, error) {
	var community entity.Community

	result := DB.First(&community, id)

	// 如果未查询到结果
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, constants.ErrorNoResult
	}

	return &community, result.Error
}
