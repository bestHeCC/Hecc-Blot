package api

import (
	coreError "hecc-blot/contract/error"

	"github.com/gin-gonic/gin"
)

type IApi interface {
	Call(ctx *gin.Context) (interface{}, coreError.IError)
}
