package md5

import (
	"crypto/md5"
	"encoding/hex"
)

// 对密码进行加密的盐
const secret = "agricultural_vision"

// 对密码进行加密
func EncryptPassword(oPassword string) string {
	h := md5.New()          // 创建一个 MD5 哈希对象
	h.Write([]byte(secret)) // 向哈希对象中写入 `secret` 的字节数据
	//把 secret 的字节数据写入到 MD5 哈希的内部状态，开始计算哈希值。 相当于让 secret 成为一个固定输入。
	return hex.EncodeToString(h.Sum([]byte(oPassword)))
	// h.Sum([]byte(oPassword))：将 oPassword 的字节作为已有哈希值的“附加值”，生成最终的哈希
	//hex.EncodeToString：将计算出的 MD5 哈希值（16 字节）转换成一个可读的十六进制字符串，便于存储或显示。
}
