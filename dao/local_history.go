// mysql/login_history.go
package dao

import (
	"time"
	"vision/dao/postgres"
	"vision/models/entity"
)

// RecordLoginHistory 记录登录历史
func RecordLoginHistory(history *entity.LoginHistory) error {
	return postgres.DB.Create(history).Error
}

// GetLoginHistory 获取登录历史（分页）
func GetLoginHistory(userID int64, offset, limit int) ([]*entity.LoginHistory, error) {
	var histories []*entity.LoginHistory

	err := postgres.DB.Where("user_id = ?", userID).
		Order("login_time DESC").
		Offset(offset).
		Limit(limit).
		Find(&histories).Error

	return histories, err
}

// GetRecentFailedAttempts 获取最近的失败尝试次数
func GetRecentFailedAttempts(email string, duration time.Duration) (int64, error) {
	var count int64

	err := postgres.DB.Model(&entity.LoginHistory{}).
		Joins("JOIN users ON users.id = login_history.user_id").
		Where("users.email = ? AND login_history.success = ? AND login_history.login_time > ?",
			email, false, time.Now().Add(-duration)).
		Count(&count).Error

	return count, err
}

// GetLoginHistoryCount 获取登录历史总数
func GetLoginHistoryCount(userID int64) (int64, error) {
	var count int64

	err := postgres.DB.Model(&entity.LoginHistory{}).
		Where("user_id = ?", userID).
		Count(&count).Error

	return count, err
}

// GetRecentLoginHistory 获取最近的登录历史（不分页）
func GetRecentLoginHistory(userID int64, limit int) ([]*entity.LoginHistory, error) {
	var histories []*entity.LoginHistory

	err := postgres.DB.Where("user_id = ?", userID).
		Order("login_time DESC").
		Limit(limit).
		Find(&histories).Error

	return histories, err
}
