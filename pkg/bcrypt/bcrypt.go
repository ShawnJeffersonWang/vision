package bcrypt

import (
	"golang.org/x/crypto/bcrypt"
)

/*
bcrypt 算法会自动生成随机盐值并将其嵌入到最终的哈希结果中，所以无需手动添加盐。
技术细节说明：
自动加盐机制：
bcrypt 在生成哈希时会自动生成一个 16 字节的随机盐
盐值会被直接存储在最终的哈希字符串中（格式示例：$2a$14$SALT_HASHED_PASSWORD）
验证时会从哈希字符串中提取盐值，无需额外存储

哈希格式解析：
$2a$14$SALT_HASHED_PASSWORD
|  |  |     |
|  |  |     └── 实际的哈希值（184位Base64编码）
|  |  └──────── 自动生成的22字符盐值（Base64编码）
|  └─────────── 成本因子（默认为14）
└────────────── 算法版本标识

bcrypt.CompareHashAndPassword 会自动：
1. 从 hashedPassword 中提取盐值
2. 使用相同的盐和成本因子对 plainPassword 重新哈希
3. 比较两个哈希结果是否一致
*/
// 对密码进行加密
func EncryptPassword(oPassword string) string {
	// 生成密码哈希，使用默认成本因子（当前为14）
	hashedBytes, _ := bcrypt.GenerateFromPassword([]byte(oPassword), bcrypt.DefaultCost)

	// 返回Base64编码的字符串格式
	return string(hashedBytes)
}

// 验证密码是否匹配
func VerifyPassword(hashedPassword, plainPassword string) bool {
	err := bcrypt.CompareHashAndPassword([]byte(hashedPassword), []byte(plainPassword))
	return err == nil
}
