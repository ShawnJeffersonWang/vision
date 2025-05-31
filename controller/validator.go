package controller

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/locales/en"
	"github.com/go-playground/locales/zh"
	ut "github.com/go-playground/universal-translator"
	"github.com/go-playground/validator/v10"
	enTranslations "github.com/go-playground/validator/v10/translations/en"
	zhTranslations "github.com/go-playground/validator/v10/translations/zh"

	"agricultural_vision/models/request"
)

// 定义一个全局翻译器
var trans ut.Translator

// 初始化翻译器
func InitTrans(locale string) (err error) {
	// 修改gin框架中的Validator引擎属性，实现自定制
	/*这行代码从 Gin 框架中获取当前的验证器引擎，确保它是 go-playground/validator 类型。
	Gin 中的 binding.Validator 是用于请求数据绑定和校验的工具。*/
	if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
		// 注册一个获取json tag的自定义方法
		/*注册了一个自定义的标签解析函数，使得 json 标签可以在结构体字段中正确地获取，
		默认情况下，Go 通过 json 标签来解析结构体字段。"-" 表示忽略该字段。*/
		v.RegisterTagNameFunc(func(fld reflect.StructField) string {
			name := strings.SplitN(fld.Tag.Get("json"), ",", 2)[0]
			if name == "-" {
				return ""
			}
			return name
		})

		zhT := zh.New() // 中文翻译器
		enT := en.New() // 英文翻译器

		// 第一个参数是备用（fallback）的语言环境
		// 后面的参数是应该支持的语言环境（支持多个）
		// uni := ut.New(zhT, zhT) 也是可以的
		uni := ut.New(enT, zhT, enT)

		// locale 通常取决于 http 请求头的 'Accept-Language'
		var ok bool
		// 也可以使用 uni.FindTranslator(...) 传入多个locale进行查找
		trans, ok = uni.GetTranslator(locale)
		if !ok {
			return fmt.Errorf("uni.GetTranslator(%s) failed", locale)
		}

		// 注册翻译器
		switch locale {
		case "en":
			err = enTranslations.RegisterDefaultTranslations(v, trans)
		case "zh":
			err = zhTranslations.RegisterDefaultTranslations(v, trans)
		default:
			err = enTranslations.RegisterDefaultTranslations(v, trans)
		}
		if err != nil {
			return err
		}

		v.RegisterStructValidation(validateParentRootID, request.CreateCommentRequest{})
		v.RegisterStructValidation(validatePostCommentID, request.VoteRequest{})

		// 方法参数：
		//1.自定义验证规则的标签
		//2.翻译器实例
		//3.注册翻译文本的函数，用于将自定义错误信息添加到翻译器中
		//4.生成最终错误信息的函数，用于在验证失败时返回翻译后的错误信息
		err = v.RegisterTranslation("parent_root", trans, func(ut ut.Translator) error {
			return ut.Add("parent_root", "ParentID 和 RootID 必须同时提供或同时为空", true)
		}, func(ut ut.Translator, fe validator.FieldError) string {
			return "ParentID 和 RootID 必须同时提供或同时为空"
		})
		if err != nil {
			return err
		}

		err = v.RegisterTranslation("post_comment", trans, func(ut ut.Translator) error {
			return ut.Add("post_comment", "PostID 和 CommentID 只能填一个值", true)
		}, func(ut ut.Translator, fe validator.FieldError) string {
			return "PostID 和 CommentID 只能填一个值"
		})
		return
	}
	return
}

// 去除提示信息中的结构体名称
/*移除结构体名称前缀，从而清理校验错误信息。
比如，如果有一个字段 User.CommunityName，那么通过此函数，会将其转换为 CommunityName，方便返回更简洁的错误信息*/
func removeTopStruct(fields map[string]string) map[string]string {
	res := map[string]string{}
	for field, err := range fields {
		res[field[strings.Index(field, ".")+1:]] = err
	}
	return res
}

// 自定义校验函数，确保 ParentID 和 RootID 要么都为空，要么都不为空
func validateParentRootID(sl validator.StructLevel) {
	// 获取当前验证的结构体实例，并断言为具体类型
	su := sl.Current().Interface().(request.CreateCommentRequest)

	if (su.ParentID == nil) != (su.RootID == nil) { // 只有一个为空，说明有问题
		// 参数分别为：字段值，字段在结构体中的名称，字段在json中的名称，自定义错误标签，附加参数（空）
		sl.ReportError(su.ParentID, "ParentID", "ParentID", "parent_root", "")
		sl.ReportError(su.RootID, "RootID", "RootID", "parent_root", "")
	}
}

// 自定义校验函数，确保 PostID 和 CommentID 必须二选一
func validatePostCommentID(sl validator.StructLevel) {
	su := sl.Current().Interface().(request.VoteRequest)
	// 如果 PostID 和 CommentID 都为 0，或者都不为 0，则报错
	if su.PostID == 0 && su.CommentID == 0 {
		sl.ReportError(su.PostID, "PostID", "PostID", "post_comment", "")
		sl.ReportError(su.CommentID, "CommentID", "CommentID", "post_comment", "")
	} else if su.PostID != 0 && su.CommentID != 0 {
		sl.ReportError(su.PostID, "PostID", "PostID", "post_comment", "")
		sl.ReportError(su.CommentID, "CommentID", "CommentID", "post_comment", "")
	}
}
