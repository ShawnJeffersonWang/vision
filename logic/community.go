package logic

import (
	"agricultural_vision/dao/mysql"
	"agricultural_vision/models/entity"
	"agricultural_vision/models/response"
)

func GetCommunityList() ([]*response.CommunityBriefResponse, error) {
	return mysql.GetCommunityList()
}

func GetCommunityDetail(id int64) (*entity.Community, error) {
	return mysql.GetCommunityById(id)
}
