package dao

import (
	"agricultural_vision/dao/postgres"
	"fmt"
	"strings"

	"agricultural_vision/constants"
	"agricultural_vision/models/entity"
)

// 创建帖子
func CreatePost(p *entity.Post) error {
	result := postgres.DB.Create(p)
	// 在执行 SQL 语句或与数据库交互过程中是否发生了错误
	if result.Error != nil {
		return result.Error
	}
	// 虽然没有发生错误，但插入操作没有成功插入任何数据
	if result.RowsAffected == 0 {
		return constants.ErrorNotAffectData
	}
	return nil
}

// 删除帖子
func DeletePost(id int64) error {
	// 先删除关联的评论
	if err := postgres.DB.Where("post_id = ?", id).Delete(&entity.Comment{}).Error; err != nil {
		return err
	}

	// 再删除帖子
	result := postgres.DB.Delete(&entity.Post{}, id)
	if result.Error != nil {
		return result.Error
	}
	if result.RowsAffected == 0 {
		return constants.ErrorNotAffectData
	}
	return nil
}

// 根据帖子id查询帖子详情
func GetPostById(pid int64) (*entity.Post, error) {
	var post *entity.Post
	result := postgres.DB.Where("id = ?", pid).First(&post)
	return post, result.Error
}

// 根据给定的id列表查询帖子数据
//func GetPostListByIDs(ids []string) ([]*entity.Post, error) {
//	var posts []*entity.Post
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
//		Find(&posts)
//
//	return posts, result.Error
//}

//func GetPostListByIDs(ids []string) ([]*entity.Post, error) {
//	var posts []*entity.Post
//	if len(ids) == 0 {
//		return posts, nil // 如果 ids 为空，直接返回，避免生成无效的 SQL
//	}
//
//	// 方案：使用 CASE 语句动态生成排序逻辑，兼容 PostgreSQL
//	// 生成的 SQL 类似于: ORDER BY CASE id WHEN 'id1' THEN 1 WHEN 'id2' THEN 2 ... END
//	var orderClause strings.Builder
//	orderClause.WriteString("CASE id ") // 注意：这里假设您的 id 字段在数据库中是 bigint 或 varchar 类型
//
//	for i, id := range ids {
//		// 将每个 id 映射到一个顺序数字。注意对 id 进行单引号包裹，以防 SQL 注入并正确处理字符串类型ID。
//		orderClause.WriteString(fmt.Sprintf("WHEN '%s' THEN %d ", id, i+1))
//	}
//	orderClause.WriteString("END")
//
//	// GORM 在处理 IN 查询时，可以直接接收 []string
//	result := postgres.DB.
//		Where("id IN ?", ids).
//		Order(orderClause.String()). // 使用动态生成的 CASE 语句进行排序
//		Find(&posts)
//
//	return posts, result.Error
//}

// 根据给定的id列表查询帖子数据 (PostgreSQL 特有方案)
func GetPostListByIDs(ids []string) ([]*entity.Post, error) {
	var posts []*entity.Post
	if len(ids) == 0 {
		return posts, nil
	}

	// 将 Go 的 string slice 格式化为 PostgreSQL 的 array literal 格式
	// 例如: ["a", "b"] -> "'a','b'"
	quotedIDs := make([]string, len(ids))
	for i, id := range ids {
		quotedIDs[i] = fmt.Sprintf("'%s'", id)
	}

	// 生成的 SQL 类似于: ORDER BY array_position(ARRAY['102','95','101'], id)
	orderClause := fmt.Sprintf("array_position(ARRAY[%s], id::text)", strings.Join(quotedIDs, ","))

	result := postgres.DB.
		Where("id IN ?", ids).
		Order(orderClause).
		Find(&posts)

	return posts, result.Error
}

// 根据userID，分页获取用户发布的帖子列表
func GetPostListByUserID(userID, page, size int64) ([]*entity.Post, int64, error) {
	var posts []*entity.Post
	var total int64

	if err := postgres.DB.Model(&entity.Post{}).Where("author_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 计算偏移量
	offset := (page - 1) * size

	// 查询二级评论（parent_id 为 commentID）
	result := postgres.DB.
		Where("author_id = ?", userID).
		Order("created_at DESC"). // 默认按时间倒序排序
		Limit(int(size)).
		Offset(int(offset)).
		Find(&posts)

	if result.Error != nil {
		return nil, 0, result.Error
	}
	return posts, total, nil
}
