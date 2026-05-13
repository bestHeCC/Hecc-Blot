package api

import (
	"context"

	coreError "hecc-blot/contract/error"
)

type IResponse interface {
	Regular(ctx context.Context, data interface{}, err coreError.IError)
}
