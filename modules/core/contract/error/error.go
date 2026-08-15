package error

import "github.com/bestHeCC/hecc-core/enum/response"

type IError interface {
	error

	GetCode() response.Value
	GetData() interface{}
}
