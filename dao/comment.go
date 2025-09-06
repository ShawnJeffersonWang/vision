package dao

import (
	"fmt"
	"strings"
	"vision/dao/postgres"

	"vision/constants"
	"vision/models/entity"
)

// 创建评论
func CreateComment(comment *entity.Comment) error {
	result := postgres.DB.Create(comment)

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

	result := postgres.DB.Delete(&entity.Comment{}, commentID)
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
//func GetCommentListByIDs(ids []string) ([]*entity.Comment, error) {
//	var comments []*entity.Comment
//
//	// 将 []string 转换为 []interface{}，gorm 会自动处理类型匹配
//	var idsInterface []interface{}
//	for _, id := range ids {
//		idsInterface = append(idsInterface, id)
//	}
//
//	//order by FIND_IN_SET(post_id, ?) 表示根据 post_id 在另一个给定字符串列表中的位置进行排序。
//	//? 是另一个占位符，将被替换为一个包含多个ID的字符串，例如 "1,3,2"。
//	result := postgres.DB.
//		Where("id IN ?", idsInterface).
//		Order(fmt.Sprintf("FIELD(id, %s)", strings.Join(ids, ","))).
//		Find(&comments)
//
//	return comments, result.Error
//}

// 根据评论ID列表获取评论列表 (已修复)
func GetCommentListByIDs(ids []string) ([]*entity.Comment, error) {
	var comments []*entity.Comment
	if len(ids) == 0 {
		return comments, nil // 如果 ids 为空，直接返回，避免执行无效的查询
	}

	// 使用 CASE 语句动态生成排序逻辑，以兼容 PostgreSQL
	// 生成的 SQL 排序子句类似于: ORDER BY CASE id WHEN 'id1' THEN 1 WHEN 'id2' THEN 2 ... END
	var orderClause strings.Builder
	orderClause.WriteString("CASE id ")

	for i, id := range ids {
		// 将每个 id 映射到一个顺序数字。
		// 使用 fmt.Sprintf 并对 id 加单引号，可以正确处理字符串类型的ID并防止SQL注入。
		orderClause.WriteString(fmt.Sprintf("WHEN '%s' THEN %d ", id, i+1))
	}
	orderClause.WriteString("END")

	// 执行 GORM 查询
	// GORM 的 IN 查询可以直接接收 []string，无需手动转换为 []interface{}
	result := postgres.DB.
		Where("id IN ?", ids).
		Order(orderClause.String()). // 应用我们动态生成的 CASE 排序语句
		Find(&comments)

	return comments, result.Error
}

// 根据分页查询子评论
func GetSonCommentList(rootID, page, size int64) ([]*entity.Comment, int64, error) {
	var comments []*entity.Comment
	var total int64

	// 查询总数
	if err := postgres.DB.Model(&entity.Comment{}).Where("root_id = ?", rootID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 计算偏移量
	offset := (page - 1) * size

	// 查询子评论（root_id 为 rootID）
	result := postgres.DB.
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
