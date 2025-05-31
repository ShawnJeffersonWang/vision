package mysql

import (
	"fmt"
	"strings"

	"agricultural_vision/constants"
	"agricultural_vision/models/entity"
)

// 创建帖子
func CreatePost(p *entity.Post) error {
	result := DB.Create(p)
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
	if err := DB.Where("post_id = ?", id).Delete(&entity.Comment{}).Error; err != nil {
		return err
	}

	// 再删除帖子
	result := DB.Delete(&entity.Post{}, id)
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
	result := DB.Where("id = ?", pid).First(&post)
	return post, result.Error
}

// 根据给定的id列表查询帖子数据
func GetPostListByIDs(ids []string) ([]*entity.Post, error) {
	var posts []*entity.Post

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
		Find(&posts)

	return posts, result.Error
}

// 根据userID，分页获取用户发布的帖子列表
func GetPostListByUserID(userID, page, size int64) ([]*entity.Post, int64, error) {
	var posts []*entity.Post
	var total int64

	if err := DB.Model(&entity.Post{}).Where("author_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	// 计算偏移量
	offset := (page - 1) * size

	// 查询二级评论（parent_id 为 commentID）
	result := DB.
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
