package api

import (
	"context"
	"net/http"

	"core/contract/api"
	coreError "core/contract/error"
	"core/enum/response"

	"github.com/gin-gonic/gin"
)

type ResponseSvc struct{}

type responseBody struct {
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

	g.JSON(http.StatusOK, responseBody{
		Code:    code,
		Message: response.CodeMap[code],
		Data:    data,
	})
}

func NewResponseSvc() api.IResponse {
	return ResponseSvc{}
}
