package error

import "hecc-blot/enum/response"

type IError interface {
	error

	GetCode() response.Value
	GetData() interface{}
}
