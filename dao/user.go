package dao

import (
	"agricultural_vision/dao/postgres"
	"agricultural_vision/models/response"
	"errors"

	"gorm.io/gorm"

	"agricultural_vision/constants"
	"agricultural_vision/models/entity"
)

// 查询邮箱是否已注册
func CheckEmailExist(email string) (bool, error) {
	var count int64
	// 使用GORM进行查询，查找符合条件的用户数量
	err := postgres.DB.Model(&entity.User{}).Where("email = ?", email).Count(&count).Error
	if err != nil {
		return false, err
	}
	// 如果邮箱已存在
	if count > 0 {
		return true, nil
	}
	// 否则返回邮箱未注册
	return false, nil
}

// 新增用户
func InsertUser(user *entity.User) error {
	return postgres.DB.Create(user).Error
}

// 用户登录
func Login(email, password string) (*entity.User, error) {
	// 新建用户结构体，用来保存查询到的用户信息
	user := new(entity.User)

	// 根据邮箱查询用户
	err := postgres.DB.Where("email = ?", email).First(user).Error
	// 如果查询不到用户，返回用户不存在错误
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return user, constants.ErrorEmailNotExist
	}

	// 判断密码是否正确
	// 如果密码不正确，返回密码不正确错误
	if password != user.Password {
		return user, constants.ErrorInvalidPassword
	}

	return user, nil
}

// GetUserByEmail 根据邮箱获取用户完整信息（包括密码哈希）
func GetUserByEmail(email string) (*entity.User, error) {
	var user entity.User

	// 使用GORM查询用户，通过邮箱精确匹配
	err := postgres.DB.Where("email = ?", email).First(&user).Error

	// 处理记录未找到的情况
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, constants.ErrorUserNotExist
	}

	// 处理其他数据库错误
	if err != nil {
		return nil, err
	}

	return &user, nil
}

// 根据用户ID更新用户信息
func UpdateUserByID(user *entity.User) error {
	err := postgres.DB.Model(&entity.User{}).Where("id = ?", user.ID).Updates(user).Error
	return err
}

// 根据用户ID获取用户详细信息
func GetUserInfo(id int64) (*entity.User, error) {
	user := new(entity.User)
	err := postgres.DB.Where("id = ?", id).First(user).Error
	return user, err
}

// GetUserByID 根据ID获取用户信息
func GetUserByID(userID int64) (*entity.User, error) {
	var user entity.User

	err := postgres.DB.Where("id = ?", userID).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, constants.ErrorUserNotExist
		}
		return nil, err
	}

	return &user, nil
}

// 根据用户ID获取用户简略信息
func GetUserBriefInfo(id int64) (*response.UserBriefResponse, error) {
	userBriefResponse := new(response.UserBriefResponse)
	err := postgres.DB.Model(&entity.User{}).Select("id", "username", "avatar").Where("id = ?", id).First(userBriefResponse).Error
	return userBriefResponse, err
}

// 根据邮箱更新用户密码
func UpdatePassword(user *entity.User) error {
	// 忽略零值动态更新
	err := postgres.DB.Model(&entity.User{}).Where("email = ?", user.Email).Update("password", user.Password).Error
	return err
}

// 根据评论ID获取评论作者简略信息
func GetUserBriefInfoByCommentID(commentID int64) (*response.UserBriefResponse, error) {
	userBriefResponse := new(response.UserBriefResponse)
	err := postgres.DB.Model(&entity.User{}).Select("id", "username", "avatar").Where("id = (select author_id from comment where id = ?)", commentID).First(userBriefResponse).Error
	return userBriefResponse, err
}
