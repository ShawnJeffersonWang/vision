package redis

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/go-redis/redis"

	"agricultural_vision/settings"
)

var (
	client *redis.Client
)

type SliceCmd = redis.SliceCmd
type StringStringMapCmd = redis.StringStringMapCmd

// Init 初始化连接
func Init(cfg *settings.RedisConfig) (err error) {
	client = redis.NewClient(&redis.Options{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password:     cfg.Password, // no password set
		DB:           cfg.DB,       // use default DB
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
	})

	_, err = client.Ping().Result()
	if err != nil {
		return
	}
	return
}

func Close() {
	_ = client.Close()
}

// Ping 检查 Redis 连接
func Ping() error {
	if client == nil {
		return fmt.Errorf("redis client not initialized")
	}

	_, err := client.Ping().Result()
	return err
}

// GetClient 获取 Redis 客户端（如果需要直接访问）
func GetClient() *redis.Client {
	return client
}

// GetInfo 获取 Redis 信息
func GetInfo() (map[string]string, error) {
	if client == nil {
		return nil, fmt.Errorf("redis client not initialized")
	}

	info, err := client.Info().Result()
	if err != nil {
		return nil, err
	}

	// 简单解析 INFO 输出
	result := make(map[string]string)
	lines := strings.Split(info, "\n")
	for _, line := range lines {
		if strings.Contains(line, ":") {
			parts := strings.SplitN(line, ":", 2)
			if len(parts) == 2 {
				result[strings.TrimSpace(parts[0])] = strings.TrimSpace(parts[1])
			}
		}
	}

	return result, nil
}

// ========== 通用操作 ==========

// Set 设置键值对
func Set(key string, value interface{}, expiration time.Duration) error {
	return client.Set(key, value, expiration).Err()
}

// Get 获取值
func Get(key string) (string, error) {
	return client.Get(key).Result()
}

// Del 删除键
func Del(key string) error {
	return client.Del(key).Err()
}

// Exists 检查键是否存在
func Exists(key string) (bool, error) {
	val, err := client.Exists(key).Result()
	if err != nil {
		return false, err
	}
	return val > 0, nil
}

// Expire 设置过期时间
func Expire(key string, expiration time.Duration) error {
	return client.Expire(key, expiration).Err()
}

// TTL 获取剩余生存时间
func TTL(key string) (time.Duration, error) {
	return client.TTL(key).Result()
}

// Keys 获取匹配的键
func Keys(pattern string) ([]string, error) {
	return client.Keys(pattern).Result()
}

// ========== JSON 操作 ==========

// SetJSON 存储JSON数据
func SetJSON(key string, value interface{}, expiration time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return err
	}
	return Set(key, data, expiration)
}

// GetJSON 获取JSON数据
func GetJSON(key string, dest interface{}) error {
	data, err := Get(key)
	if err != nil {
		return err
	}
	return json.Unmarshal([]byte(data), dest)
}

// ========== Hash 操作 ==========

// HSet 设置hash字段
func HSet(key, field string, value interface{}) error {
	return client.HSet(key, field, value).Err()
}

// HGet 获取hash字段
func HGet(key, field string) (string, error) {
	return client.HGet(key, field).Result()
}

// HGetAll 获取所有hash字段
func HGetAll(key string) (map[string]string, error) {
	return client.HGetAll(key).Result()
}

// HDel 删除hash字段
func HDel(key string, fields ...string) error {
	return client.HDel(key, fields...).Err()
}

// HExists 检查hash字段是否存在
func HExists(key, field string) (bool, error) {
	return client.HExists(key, field).Result()
}

// ========== 专门用于 Token 管理的函数 ==========

// SaveRefreshToken 保存刷新token
func SaveRefreshToken(userID int64, deviceID, token string, expiration time.Duration) error {
	key := fmt.Sprintf("refresh_token:%d:%s", userID, deviceID)

	tokenData := map[string]interface{}{
		"token":      token,
		"user_id":    userID,
		"device_id":  deviceID,
		"created_at": time.Now().Unix(),
		"last_used":  time.Now().Unix(),
	}

	return SetJSON(key, tokenData, expiration)
}

// GetRefreshToken 获取刷新token
func GetRefreshToken(userID int64, deviceID string) (map[string]interface{}, error) {
	key := fmt.Sprintf("refresh_token:%d:%s", userID, deviceID)

	var tokenData map[string]interface{}
	err := GetJSON(key, &tokenData)
	if err != nil {
		return nil, err
	}

	return tokenData, nil
}

// DeleteRefreshToken 删除刷新token
func DeleteRefreshToken(userID int64, deviceID string) error {
	key := fmt.Sprintf("refresh_token:%d:%s", userID, deviceID)
	return Del(key)
}

// DeleteUserAllTokens 删除用户所有token
func DeleteUserAllTokens(userID int64) error {
	pattern := fmt.Sprintf("refresh_token:%d:*", userID)
	keys, err := Keys(pattern)
	if err != nil {
		return err
	}

	for _, key := range keys {
		if err := Del(key); err != nil {
			return err
		}
	}

	return nil
}

// UpdateRefreshTokenLastUsed 更新token最后使用时间
func UpdateRefreshTokenLastUsed(userID int64, deviceID string) error {
	key := fmt.Sprintf("refresh_token:%d:%s", userID, deviceID)

	var tokenData map[string]interface{}
	if err := GetJSON(key, &tokenData); err != nil {
		return err
	}

	tokenData["last_used"] = time.Now().Unix()

	// 重新设置，保持原有的过期时间
	ttl, err := TTL(key)
	if err != nil {
		return err
	}

	return SetJSON(key, tokenData, ttl)
}

// IsRefreshTokenValid 验证刷新token是否有效
func IsRefreshTokenValid(userID int64, deviceID, token string) bool {
	tokenData, err := GetRefreshToken(userID, deviceID)
	if err != nil {
		return false
	}

	storedToken, ok := tokenData["token"].(string)
	if !ok || storedToken != token {
		return false
	}

	// 更新最后使用时间
	_ = UpdateRefreshTokenLastUsed(userID, deviceID)

	return true
}

// SaveUserSession 保存用户会话信息（可选）
func SaveUserSession(userID int64, sessionData map[string]interface{}, expiration time.Duration) error {
	key := fmt.Sprintf("session:%d", userID)
	return SetJSON(key, sessionData, expiration)
}

// GetUserSession 获取用户会话信息
func GetUserSession(userID int64) (map[string]interface{}, error) {
	key := fmt.Sprintf("session:%d", userID)

	var sessionData map[string]interface{}
	err := GetJSON(key, &sessionData)
	if err != nil {
		return nil, err
	}

	return sessionData, nil
}
