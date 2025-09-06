package dao

import (
	"errors"
	"vision/dao/postgres"

	"gorm.io/gorm"

	"vision/constants"
	"vision/models/entity"
	"vision/models/response"
)

// CreateCommunity 创建社区
func CreateCommunity(community *entity.Community) error {
	result := postgres.DB.Create(community)
	return result.Error
}

// CheckCommunityNameExists 检查社区名称是否已存在
func CheckCommunityNameExists(name string) (bool, error) {
	var count int64
	result := postgres.DB.Model(&entity.Community{}).
		Where("community_name = ?", name).
		Count(&count)

	if result.Error != nil {
		return false, result.Error
	}

	return count > 0, nil
}

// 查询社区列表
func GetCommunityList() ([]*response.CommunityBriefResponse, error) {
	var communities []*response.CommunityBriefResponse

	result := postgres.DB.Model(&entity.Community{}).
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

	result := postgres.DB.First(&community, id)

	// 如果未查询到结果
	if errors.Is(result.Error, gorm.ErrRecordNotFound) {
		return nil, constants.ErrorNoResult
	}

	return &community, result.Error
}
