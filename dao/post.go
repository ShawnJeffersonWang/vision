package dao

import (
	"fmt"
	"strconv"
	"strings"
	"vision/dao/postgres"
	"vision/models/request"

	"vision/constants"
	"vision/models/entity"
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

// GetPostIDs 根据请求参数查询符合条件的 ID 列表
func GetPostIDs(p *request.ListRequest) (ids []string, total int64, err error) {
	// 假设全局DB对象为 global.DB
	db := postgres.DB.Model(&entity.Post{}) // 对应你的 Post 模型

	// 1. 如果有筛选条件（如分区ID），在这里添加
	// if p.CommunityID != 0 {
	//     db = db.Where("community_id = ?", p.CommunityID)
	// }

	// 2. 查询总数
	if err = db.Count(&total).Error; err != nil {
		return
	}
	if total == 0 {
		return
	}

	// 3. 处理排序
	// 注意：如果按分数排序，前提是你的数据库表里有 score 字段，或者你可以接受按其他字段代替
	switch p.Order {
	case constants.OrderScore:
		// 假设数据库里没有 score 字段，通常业务会以降序时间作为兜底，或者你有 score 字段
		db = db.Order("score DESC, created_at DESC")
	default:
		// 默认按创建时间倒序
		db = db.Order("created_at DESC")
	}

	// 4. 分页并只取 ID
	// Pluck 会自动把查询到的 id 放入 ids 切片中
	offset := (p.Page - 1) * p.Size
	// 注意：这里要把 int ID 转为 string，因为你的 GetPostListByIDs 接收 []string
	// 如果数据库ID是int，GORM 的 Pluck 可能会报错类型不匹配，建议先取 []int64 再转 string
	var intIDs []int64
	err = db.Limit(int(p.Size)).Offset(int(offset)).Pluck("id", &intIDs).Error

	// 类型转换 int64 -> string
	ids = make([]string, len(intIDs))
	for i, v := range intIDs {
		ids[i] = strconv.FormatInt(v, 10)
	}

	return
}
