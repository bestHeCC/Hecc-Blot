package api

import (
	"context"
	"net/http"

	"hecc-blot/contract/api"
	coreError "hecc-blot/contract/error"
	"hecc-blot/enum/response"

	"github.com/gin-gonic/gin"
)

type ResponseSvc struct {
	Code    response.Value `json:"code"`
	Message string         `json:"message"`
	Data    interface{}    `json:"data"`
}

func (r ResponseSvc) Regular(ctx context.Context, data interface{}, err coreError.IError) {
	g := ctx.(*gin.Context)
	code := response.Success
	if err != nil {
		code = err.GetCode()
		data = err.GetData()
	}

	r.Code = code
	r.Message = response.CodeMap[code]
	r.Data = data

	g.JSON(http.StatusOK, r)
}

func NewResponseSvc() api.IResponse {
	return ResponseSvc{}
}
