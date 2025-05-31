package mysql

import (
	"fmt"
	"strings"

	"agricultural_vision/constants"
	"agricultural_vision/models/entity"
)

// 创建评论
func CreateComment(comment *entity.Comment) error {
	result := DB.Create(comment)

	if result.Error != nil {
		return result.Error
	}

	if result.RowsAffected == 0 {
		return constants.ErrorNotAffectData
	}

	return nil
}

// 删除评论
func DeleteComment(commentID int64) error {
	// 判断是否是一级评论

	result := DB.Delete(&entity.Comment{}, commentID)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return constants.ErrorNotAffectData
	}

	return nil
}

/*// 根据评论ID获取父评论ID和帖子ID
func GetParentIDAndPostIDByCommentID(commentID int64) (*int64, *int64, error) {
	// 定义一个结构体来接收查询结果
	var result struct {
		ParentID *int64 `json:"parent_id"`
		PostID   *int64 `json:"post_id"`
	}

	// 查询父评论ID和帖子ID
	err := DB.Model(&entity.Comment{}).
		Select("parent_id", "post_id").
		Where("id = ?", commentID).
		Scan(&result).Error
	if err != nil {
		return nil, nil, err
	}

	return result.ParentID, result.PostID, nil
}*/

// 根据评论ID列表获取评论列表
func GetCommentListByIDs(ids []string) ([]*entity.Comment, error) {
	var comments []*entity.Comment

	// 将 []string 转换为 []interface{}，gorm 会自动处理类型匹配
	var idsInterface []interface{}
	for _, id := range ids {
		idsInterface = append(idsInterface, id)
	}

	//order by FIND_IN_SET(post_id, ?) 表示根据 post_id 在另一个给定字符串列表中的位置进行排序。
	//? 是另一个占位符，将被替换为一个包含多个ID的字符串，例如 "1,3,2"。
	result := DB.
		Where("id IN ?", idsInterface).
		Order(fmt.Sprintf("FIELD(id, %s)", strings.Join(ids, ","))).
		Find(&comments)

	return comments, result.Error
}

// 根据分页查询子评论
func GetSonCommentList(rootID, page, size int64) ([]*entity.Comment, int64, error) {
	var comments []*entity.Comment
	var total int64

	// 查询总数
	if err := DB.Model(&entity.Comment{}).Where("root_id = ?", rootID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 计算偏移量
	offset := (page - 1) * size

	// 查询子评论（root_id 为 rootID）
	result := DB.
		Where("root_id = ?", rootID).
		Order("created_at ASC"). // 默认按时间正序排序（新发布的在后面）
		Limit(int(size)).
		Offset(int(offset)).
		Find(&comments)

	if result.Error != nil {
		return nil, 0, result.Error
	}
	return comments, total, nil
}
