package error

import "core/enum/response"

type IError interface {
	error

	GetCode() response.Value
	GetData() interface{}
}
