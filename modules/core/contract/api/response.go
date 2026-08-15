package api

import (
	"context"

	coreError "github.com/bestHeCC/hecc-core/contract/error"
)

type IResponse interface {
	Regular(ctx context.Context, data interface{}, err coreError.IError)
}
