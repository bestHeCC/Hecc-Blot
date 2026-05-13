package api

import (
	"context"

	coreError "core/contract/error"
)

type IResponse interface {
	Regular(ctx context.Context, data interface{}, err coreError.IError)
}
