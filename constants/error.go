package constants

import (
	"errors"
)

// 封装常见的错误信息，向上层进行返回。这里的错误是业务逻辑错误
// 在controller层中对错误信息进行判断，向前端返回最终错误信息
var (
	ErrorEmailExist       = errors.New(CodeEmailExist)
	ErrorEmailNotExist    = errors.New(CodeEmailNotExist)
	ErrorInvalidPassword  = errors.New(CodeInvalidPassword)
	ErrorInvalidEmailCode = errors.New(CodeInvalidEmailCode)
	ErrorNeedLogin        = errors.New(CodeNeedLogin)
	ErrorInvalidAToken    = errors.New(CodeInvalidToken)
	ErrorAiNotAnswer      = errors.New(CodeAiNotAnswer)
	ErrorNotAffectData    = errors.New(CodeNotAffectData)
	ErrorNoResult         = errors.New(CodeNoResult)
	ErrorNoPermission     = errors.New(CodeNoPermission)
	ErrorNoPost           = errors.New(CodeNoPost)
	ErrorNoComment        = errors.New(CodeNoComment)
)
