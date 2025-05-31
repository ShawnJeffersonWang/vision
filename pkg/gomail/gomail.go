package gomail

import (
	"fmt"
	"math/rand"
	"time"

	"agricultural_vision/constants"

	"github.com/patrickmn/go-cache"
	"gopkg.in/gomail.v2"
)

var (
	// 缓存，用于存储验证码和邮箱的对应关系
	// 5分钟内有效，且每隔10分钟进行一次清理
	verificationCodeCache = cache.New(5*time.Minute, 10*time.Minute)
)

// 生成6位随机验证码
func generateVerificationCode() string {
	rand.Seed(time.Now().UnixNano())
	return fmt.Sprintf("%06d", rand.Intn(1000000))
}

// 发送验证码到邮箱并存入缓存
func SendVerificationCode(toEmail string) error {
	// 配置发件人邮箱
	smtpHost := "smtp.qq.com"
	smtpPort := 465
	fromEmail := "2455494167@qq.com"
	authCode := "rpqcsjeyqoesecbd"

	// 生成验证码
	verificationCode := generateVerificationCode()

	// 创建邮件
	m := gomail.NewMessage()
	// m.FormatAddress将发件人邮箱和名称格式化为MIME编码的地址
	m.SetHeader("From", m.FormatAddress(fromEmail, "农视界"))
	m.SetHeader("To", toEmail)
	m.SetHeader("Subject", "验证码")
	m.SetBody("text/html", fmt.Sprintf(`
		<!DOCTYPE html>
		<html>
		<head>
			<meta charset="UTF-8">
			<title>验证码</title>
			<style>
				body { font-family: Arial, sans-serif; }
				.container { padding: 20px; border: 1px solid #ddd; border-radius: 5px; }
				h1 { color: #333; }
				.code { font-size: 24px; font-weight: bold; color: #007bff; }
				.footer { margin-top: 20px; font-size: 12px; color: #888; }
			</style>
		</head>
		<body>
			<div class="container">
				<h1>你的验证码</h1>
				<p class="code">%s</p>
				<p>有效时间为 5 分钟</p>
				<div class="footer">如果您没有请求此验证码，请忽略此邮件。</div>
			</div>
		</body>
		</html>
	`, verificationCode))

	// 发送邮件
	d := gomail.NewDialer(smtpHost, smtpPort, fromEmail, authCode)
	d.SSL = true
	if err := d.DialAndSend(m); err != nil {
		return err
	}

	// 将验证码存入缓存
	verificationCodeCache.Set(toEmail, verificationCode, cache.DefaultExpiration)
	return nil
}

// 校验验证码
func VerifyVerificationCode(email string, code string) error {
	cachedCode, found := verificationCodeCache.Get(email)

	// 如果找不到验证码或验证码已过期
	if !found {
		return constants.ErrorInvalidEmailCode
	}

	// 如果验证码不匹配
	if cachedCode != code {
		return constants.ErrorInvalidEmailCode
	}

	return nil
}
