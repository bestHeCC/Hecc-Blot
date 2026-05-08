package api

import (
	"context"

	coreError "core/contract/error"
)

type IApi interface {
	Call(ctx context.Context) (interface{}, coreError.IError)
}
