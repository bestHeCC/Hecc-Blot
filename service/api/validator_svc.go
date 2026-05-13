package api

import (
	"errors"

	"core/contract/api"

	"github.com/go-playground/validator/v10"
)

// GetErrorMsg 获取错误信息
func GetErrorMsg(request interface{}, err error) string {
	if _, ok := errors.AsType[validator.ValidationErrors](err); ok {
		_, isValidator := request.(api.IValidator)

		for _, v := range err.(validator.ValidationErrors) {
			// 若 request 结构体实现 Validator 接口即可实现自定义错误信息
			if isValidator {
				if message, exist := request.(api.IValidator).GetMessages()[v.Field()+"."+v.Tag()]; exist {
					return message
				}
			}
			return v.Error()
		}
	}

	return "Parameter error"
}
