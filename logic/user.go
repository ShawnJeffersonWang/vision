package logic

import (
	"fmt"
	"go.uber.org/zap"
	"time"
	"vision/constants"
	"vision/dao"
	"vision/dao/redis"
	"vision/models/entity"
	"vision/models/request"
	"vision/models/response"
	"vision/pkg/bcrypt"
	"vision/pkg/gomail"
	auth "vision/pkg/jwt"
)

// 用户注册
func SingUp(p *request.SignUpRequest) error {
	// 1.判断邮箱是否已注册
	flag, err := dao.CheckEmailExist(p.Email)
	// 如果数据库查询出错
	if err != nil {
		return err
	}
	// 如果邮箱已注册
	if flag {
		return constants.ErrorEmailExist
	}

	// 2.校验邮箱
	if err = gomail.VerifyVerificationCode(p.Email, p.Code); err != nil {
		return constants.ErrorInvalidEmailCode
	}

	user := entity.User{
		Username: p.Username,
		Password: bcrypt.EncryptPassword(p.Password),
		Email:    p.Email,
	}

	// 3.保存进数据库
	err = dao.InsertUser(&user)
	return err
}

// 修改密码
func ChangePassword(p *request.ChangePasswordRequest) error {
	// 验证邮箱是否已注册
	flag, err := dao.CheckEmailExist(p.Email)
	// 如果数据库查询出错
	if err != nil {
		return err
	}
	// 如果邮箱未注册
	if !flag {
		return constants.ErrorEmailNotExist
	}

	// 验证邮箱验证码是否正确
	if err = gomail.VerifyVerificationCode(p.Email, p.Code); err != nil {
		return constants.ErrorInvalidEmailCode
	}

	// 修改密码
	// 先对密码明文进行加密
	p.Password = bcrypt.EncryptPassword(p.Password)
	user := entity.User{
		Password: p.Password,
		Email:    p.Email,
	}

	// 再更新数据库
	return dao.UpdatePassword(&user)
}

// 用户登录
//func Login(p *request.LoginRequest) (string, error) {
//	//可以从user中拿到UserID
//	user, err := mysql.Login(p.Email, md5.EncryptPassword(p.Password))
//	if err != nil {
//		return "", err
//	}
//
//	//生成JWT
//	token, err := auth.GenToken(user.ID, user.Username)
//	return token, err
//}

// Login 用户登录 - 返回访问token和刷新token
func Login(p *request.LoginRequest) (string, error) {
	// 记录登录尝试的开始时间
	loginTime := time.Now()

	// 验证用户
	//user, err := mysql.Login(p.Email, md5.EncryptPassword(p.Password))
	//if err != nil {
	//	return "", err
	//}

	// 1. 先查询用户信息（注意：这里不验证密码）
	user, err := dao.GetUserByEmail(p.Email)
	if err != nil {
		// 统一返回模糊错误，避免信息泄露
		return "", constants.ErrorInvalidCredentials
	}

	// 2. 验证密码（使用 bcrypt 比较明文和数据库中的哈希值）
	// 登录逻辑中使用自定义函数
	if !bcrypt.VerifyPassword(user.Password, p.Password) {
		return "", constants.ErrorInvalidCredentials
	}

	// 生成访问token（短期）
	accessToken, err := auth.GenAccessToken(user.ID, user.Username)
	if err != nil {
		zap.L().Error("生成访问token失败", zap.Error(err))
		return "", err
	}

	// 生成刷新token（长期）
	refreshToken, err := auth.GenRefreshToken(user.ID, user.Username)
	if err != nil {
		zap.L().Error("生成刷新token失败", zap.Error(err))
		return "", err
	}

	// 保存刷新token到Redis
	if err := saveRefreshToken(user.ID, refreshToken, p.DeviceID, p.DeviceInfo); err != nil {
		// 记录错误但不影响登录
		zap.L().Error("保存刷新token失败",
			zap.Int64("user_id", user.ID),
			zap.Error(err))
	}

	// 记录成功的登录历史和更新用户信息（异步执行）
	go func() {
		// 记录登录历史
		history := &entity.LoginHistory{
			UserID:    user.ID,
			Username:  user.Username, // 添加用户名
			LoginTime: loginTime,
			LoginIP:   p.ClientIP,
			UserAgent: p.UserAgent,
			DeviceID:  p.DeviceID,
			Success:   true,
		}

		if err := dao.RecordLoginHistory(history); err != nil {
			zap.L().Error("记录登录历史失败",
				zap.Int64("user_id", user.ID),
				zap.Error(err))
		}
	}()

	// 记录登录日志
	zap.L().Info("用户登录成功",
		zap.Int64("user_id", user.ID),
		zap.String("username", user.Username),
		zap.String("email", user.Email),
		zap.String("device_id", p.DeviceID),
		zap.String("client_ip", p.ClientIP))

	return accessToken, nil
}

// GetLoginHistory 获取用户登录历史
func GetLoginHistory(userID int64, page, pageSize int) (*response.LoginHistoryResponse, error) {
	// 参数验证
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	if pageSize > 100 {
		pageSize = 100 // 限制最大值
	}

	// 计算偏移量
	offset := (page - 1) * pageSize

	// 获取登录历史
	histories, err := dao.GetLoginHistory(userID, offset, pageSize)
	if err != nil {
		zap.L().Error("获取登录历史失败",
			zap.Int64("user_id", userID),
			zap.Error(err))
		return nil, err
	}

	// 获取总数
	total, err := dao.GetLoginHistoryCount(userID)
	if err != nil {
		zap.L().Error("获取登录历史总数失败",
			zap.Int64("user_id", userID),
			zap.Error(err))
		return nil, err
	}

	// 构建响应
	return &response.LoginHistoryResponse{
		Total:    total,
		Page:     page,
		PageSize: pageSize,
		List:     histories,
	}, nil
}

// GetMyLoginHistory 获取当前用户的登录历史
func GetMyLoginHistory(userID int64, req *request.GetLoginHistoryRequest) (*response.LoginHistoryResponse, error) {
	return GetLoginHistory(userID, req.Page, req.PageSize)
}

// RefreshToken 刷新token
func RefreshToken(refreshToken string, deviceID string) (*response.TokenResponse, error) {
	// 解析刷新token
	claims, err := auth.ParseRefreshToken(refreshToken)
	if err != nil {
		return nil, constants.ErrorInvalidToken
	}

	// 验证刷新token是否有效
	if !isRefreshTokenValid(claims.UserID, refreshToken, deviceID) {
		return nil, constants.ErrorTokenExpired
	}

	// 检查用户状态
	user, err := dao.GetUserByID(claims.UserID)
	if err != nil {
		return nil, err
	}

	// 生成新的访问token
	newAccessToken, err := auth.GenAccessToken(user.ID, user.Username)
	if err != nil {
		return nil, err
	}

	// 生成新的刷新token（可选：轮转机制）
	newRefreshToken, err := auth.GenRefreshToken(user.ID, user.Username)
	if err != nil {
		return nil, err
	}

	// 更新Redis中的刷新token
	if err := updateRefreshToken(user.ID, refreshToken, newRefreshToken, deviceID); err != nil {
		zap.L().Error("更新刷新token失败", zap.Error(err))
	}

	zap.L().Info("刷新token成功",
		zap.Int64("user_id", user.ID),
		zap.String("device_id", deviceID))

	return &response.TokenResponse{
		AccessToken:  newAccessToken,
		RefreshToken: newRefreshToken,
		ExpiresIn:    7200,
		TokenType:    "Bearer",
	}, nil
}

// Logout 用户登出
func Logout(userID int64, deviceID string) error {
	key := fmt.Sprintf("refresh_token:%d:%s", userID, deviceID)
	return redis.Del(key)
}

// LogoutAllDevices 登出所有设备
func LogoutAllDevices(userID int64) error {
	pattern := fmt.Sprintf("refresh_token:%d:*", userID)
	keys, err := redis.Keys(pattern)
	if err != nil {
		return err
	}

	for _, key := range keys {
		if err := redis.Del(key); err != nil {
			zap.L().Error("删除refresh token失败",
				zap.String("key", key),
				zap.Error(err))
		}
	}

	return nil
}

// 辅助函数
func saveRefreshToken(userID int64, token, deviceID, deviceInfo string) error {
	key := fmt.Sprintf("refresh_token:%d:%s", userID, deviceID)

	tokenInfo := map[string]interface{}{
		"token":       token,
		"user_id":     userID,
		"device_id":   deviceID,
		"device_info": deviceInfo,
		"created_at":  time.Now().Unix(),
		"last_used":   time.Now().Unix(),
	}

	// 设置30天过期
	return redis.SetJSON(key, tokenInfo, 30*24*time.Hour)
}

func isRefreshTokenValid(userID int64, token, deviceID string) bool {
	key := fmt.Sprintf("refresh_token:%d:%s", userID, deviceID)

	var tokenInfo map[string]interface{}
	if err := redis.GetJSON(key, &tokenInfo); err != nil {
		return false
	}

	// 验证token是否匹配
	if tokenInfo["token"] != token {
		return false
	}

	// 更新最后使用时间
	tokenInfo["last_used"] = time.Now().Unix()
	redis.SetJSON(key, tokenInfo, 30*24*time.Hour)

	return true
}

func updateRefreshToken(userID int64, oldToken, newToken, deviceID string) error {
	// 先删除旧的
	oldKey := fmt.Sprintf("refresh_token:%d:%s", userID, deviceID)
	redis.Del(oldKey)

	// 保存新的
	return saveRefreshToken(userID, newToken, deviceID, "")
}

// 获取用户信息
func GetUserInfo(id int64) (*entity.User, error) {
	return dao.GetUserInfo(id)
}

// 更新用户信息
func UpdateUserInfo(p *request.UpdateUserInfoRequest, id int64) error {
	// 1. 邮箱校验
	// 查询用户原本的邮箱
	user, err := dao.GetUserInfo(id)
	if err != nil {
		return err
	}
	// 判断邮箱是否已注册
	flag, err := dao.CheckEmailExist(p.Email)
	if err != nil {
		return err
	}
	// 如果新邮箱和原本邮箱不同，且新邮箱已被其他用户注册
	if user.Email != p.Email && flag {
		return constants.ErrorEmailExist
	}

	// 2. 更新用户信息
	newUser := entity.User{
		BaseModel: entity.BaseModel{ID: id},
		Username:  p.Username,
		Email:     p.Email,
		Avatar:    p.Avatar,
	}

	return dao.UpdateUserByID(&newUser)
}

// 查询用户主页
func GetUserHomePage(targetUserID int64) (*response.UserHomePageResponse, error) {
	userHomePageResponse := &response.UserHomePageResponse{
		Posts: &response.PostListResponse{
			Posts: []*response.PostResponse{},
		},
		LikedPosts: &response.PostListResponse{
			Posts: []*response.PostResponse{},
		},
	}

	// 获取用户基本信息并填充
	userInfo, err := dao.GetUserInfo(targetUserID)
	if err != nil {
		return nil, err
	}
	userHomePageResponse.ID = userInfo.ID
	userHomePageResponse.Username = userInfo.Username
	userHomePageResponse.Email = userInfo.Email
	userHomePageResponse.Avatar = userInfo.Avatar

	// 获取用户帖子列表并填充
	userPostList, err := GetUserPostList(targetUserID, &request.ListRequest{
		Page: 1,
		Size: 10,
	})
	if err != nil {
		return nil, err
	}

	// 获取用户点赞帖子列表并填充
	userLikedPostList, err := GetUserLikedPostList(targetUserID, &request.ListRequest{
		Page: 1,
		Size: 10,
	})
	if err != nil {
		return nil, err
	}

	userHomePageResponse.Posts = userPostList
	userHomePageResponse.LikedPosts = userLikedPostList
	return userHomePageResponse, nil
}
